package pdf

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SimplePDFStorage provides basic PDF storage functionality
type SimplePDFStorage struct {
	baseDir string
}

// StoredPDF represents a stored PDF with metadata
type StoredPDF struct {
	DocumentID      string                 `json:"document_id"`
	DocumentName    string                 `json:"document_name"`
	DocumentSize    uint64                 `json:"document_size"`
	ContentHash     string                 `json:"content_hash"`
	UploadTimestamp time.Time              `json:"upload_timestamp"`
	UploaderID      string                 `json:"uploader_id"`
	Title           string                 `json:"title,omitempty"`
	Author          string                 `json:"author,omitempty"`
	Description     string                 `json:"description,omitempty"`
	Keywords        []string               `json:"keywords,omitempty"`
	Tags            []string               `json:"tags,omitempty"`
	CustomFields    map[string]string     `json:"custom_fields,omitempty"`
}

// NewSimplePDFStorage creates a new simple PDF storage instance
func NewSimplePDFStorage(baseDir string) (*SimplePDFStorage, error) {
	// Create directories
	dirs := []string{baseDir, filepath.Join(baseDir, "content"), filepath.Join(baseDir, "metadata")}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	
	return &SimplePDFStorage{baseDir: baseDir}, nil
}

// StorePDF stores a PDF document
func (s *SimplePDFStorage) StorePDF(
	documentContent []byte,
	documentName string,
	uploaderID string,
	metadata PDFMetadata,
) (*StoredPDF, error) {
	
	// Calculate hash
	hash := sha256.Sum256(documentContent)
	documentID := hex.EncodeToString(hash[:])
	
	// Create stored PDF info
	storedPDF := &StoredPDF{
		DocumentID:      documentID,
		DocumentName:    documentName,
		DocumentSize:    uint64(len(documentContent)),
		ContentHash:     hex.EncodeToString(hash[:]),
		UploadTimestamp: time.Now().UTC(),
		UploaderID:      uploaderID,
		Title:           metadata.Title,
		Author:          metadata.Author,
		Description:     metadata.Description,
		Keywords:        metadata.Keywords,
		Tags:            metadata.Tags,
		CustomFields:    metadata.CustomFields,
	}
	
	// Store content
	contentPath := filepath.Join(s.baseDir, "content", documentID)
	if err := os.WriteFile(contentPath, documentContent, 0644); err != nil {
		return nil, fmt.Errorf("failed to store content: %w", err)
	}
	
	// Store metadata
	metadataPath := filepath.Join(s.baseDir, "metadata", documentID+".json")
	metadataData, err := json.MarshalIndent(storedPDF, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}
	
	if err := os.WriteFile(metadataPath, metadataData, 0644); err != nil {
		return nil, fmt.Errorf("failed to store metadata: %w", err)
	}
	
	return storedPDF, nil
}

// GetPDF retrieves a PDF document
func (s *SimplePDFStorage) GetPDF(documentID string) ([]byte, *StoredPDF, error) {
	// Load metadata
	metadataPath := filepath.Join(s.baseDir, "metadata", documentID+".json")
	metadataData, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read metadata: %w", err)
	}
	
	var storedPDF StoredPDF
	if err := json.Unmarshal(metadataData, &storedPDF); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}
	
	// Load content
	contentPath := filepath.Join(s.baseDir, "content", documentID)
	content, err := os.ReadFile(contentPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read content: %w", err)
	}
	
	// Verify hash
	calculatedHash := sha256.Sum256(content)
	if hex.EncodeToString(calculatedHash[:]) != storedPDF.ContentHash {
		return nil, nil, fmt.Errorf("content hash mismatch")
	}
	
	return content, &storedPDF, nil
}

// ListPDFs returns a list of stored PDFs
func (s *SimplePDFStorage) ListPDFs() ([]*StoredPDF, error) {
	metadataDir := filepath.Join(s.baseDir, "metadata")
	
	files, err := os.ReadDir(metadataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata directory: %w", err)
	}
	
	var pdfs []*StoredPDF
	
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}
		
		documentID := file.Name()[:len(file.Name())-5] // Remove .json
		_, metadata, err := s.GetPDF(documentID)
		if err != nil {
			continue // Skip corrupted files
		}
		
		pdfs = append(pdfs, metadata)
	}
	
	return pdfs, nil
}
