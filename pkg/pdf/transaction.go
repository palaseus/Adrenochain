package pdf

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
	"bytes"
	"encoding/binary"

	"github.com/palaseus/adrenochain/pkg/block"
)

// PDFTransaction represents a special transaction type for storing PDF documents
// on the blockchain with immutability, hashing, and timestamping
type PDFTransaction struct {
	*block.Transaction                    // Embed the base transaction
	DocumentHash     []byte        // SHA256 hash of the PDF content
	DocumentSize     uint64        // Size of the PDF in bytes
	DocumentType     string        // MIME type (e.g., "application/pdf")
	DocumentName     string        // Original filename
	UploadTimestamp  time.Time     // When the document was uploaded
	ContentHash      []byte        // Hash of the actual PDF content
	Metadata         PDFMetadata   // Additional metadata about the document
	Signature        []byte        // Digital signature of the uploader
	PublicKey       []byte        // Public key of the uploader
}

// PDFMetadata contains additional information about the PDF document
type PDFMetadata struct {
	Title           string            `json:"title,omitempty"`
	Author          string            `json:"author,omitempty"`
	Subject         string            `json:"subject,omitempty"`
	Keywords        []string          `json:"keywords,omitempty"`
	PageCount       uint32            `json:"page_count,omitempty"`
	FileSize        uint64            `json:"file_size"`
	UploaderID      string            `json:"uploader_id"`
	Tags            []string          `json:"tags,omitempty"`
	Description     string            `json:"description,omitempty"`
	CustomFields    map[string]string `json:"custom_fields,omitempty"`
}

// NewPDFTransaction creates a new PDF transaction
func NewPDFTransaction(
	documentContent []byte,
	documentName string,
	uploaderID string,
	metadata PDFMetadata,
	inputs []*block.TxInput,
	outputs []*block.TxOutput,
	fee uint64,
) *PDFTransaction {
	
	// Calculate document hash
	documentHash := sha256.Sum256(documentContent)
	
	// Create the base transaction
	baseTx := block.NewTransaction(inputs, outputs, fee)
	
	// Create PDF transaction
	pdfTx := &PDFTransaction{
		Transaction:     baseTx,
		DocumentHash:    documentHash[:],
		DocumentSize:    uint64(len(documentContent)),
		DocumentType:    "application/pdf",
		DocumentName:    documentName,
		UploadTimestamp: time.Now().UTC(),
		ContentHash:     documentHash[:],
		Metadata:        metadata,
		PublicKey:       nil, // Will be set when signing
	}
	
	// Update metadata with calculated values
	pdfTx.Metadata.FileSize = pdfTx.DocumentSize
	pdfTx.Metadata.UploaderID = uploaderID
	
	// Calculate the transaction hash including PDF-specific data
	pdfTx.Hash = pdfTx.CalculatePDFHash()
	
	// Update the base transaction hash to match the PDF transaction hash
	baseTx.Hash = pdfTx.Hash
	
	return pdfTx
}

// CalculatePDFHash calculates the hash of the PDF transaction including all PDF-specific fields
func (pt *PDFTransaction) CalculatePDFHash() []byte {
	data := make([]byte, 0)
	
	// Include base transaction hash
	data = append(data, pt.Transaction.Hash...)
	
	// Include document hash
	data = append(data, pt.DocumentHash...)
	
	// Include document size
	sizeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(sizeBytes, pt.DocumentSize)
	data = append(data, sizeBytes...)
	
	// Include document type
	data = append(data, []byte(pt.DocumentType)...)
	
	// Include document name
	data = append(data, []byte(pt.DocumentName)...)
	
	// Include upload timestamp
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, uint64(pt.UploadTimestamp.UnixNano()))
	data = append(data, timestampBytes...)
	
	// Include content hash
	data = append(data, pt.ContentHash...)
	
	// Include metadata hash (hash of serialized metadata)
	metadataHash := pt.calculateMetadataHash()
	data = append(data, metadataHash...)
	
	// Calculate final hash
	hash := sha256.Sum256(data)
	return hash[:]
}

// calculateMetadataHash calculates the hash of the metadata
func (pt *PDFTransaction) calculateMetadataHash() []byte {
	// For simplicity, we'll hash the metadata as a JSON string
	// In production, you might want to use a more sophisticated approach
	metadataStr := fmt.Sprintf("%s|%s|%s|%d|%s|%s",
		pt.Metadata.Title,
		pt.Metadata.Author,
		pt.Metadata.Subject,
		pt.Metadata.PageCount,
		pt.Metadata.UploaderID,
		pt.Metadata.Description,
	)
	
	hash := sha256.Sum256([]byte(metadataStr))
	return hash[:]
}

// VerifyDocumentIntegrity verifies that the document content matches the stored hash
func (pt *PDFTransaction) VerifyDocumentIntegrity(documentContent []byte) bool {
	calculatedHash := sha256.Sum256(documentContent)
	return bytes.Equal(calculatedHash[:], pt.ContentHash)
}

// GetDocumentID returns a unique identifier for the document
func (pt *PDFTransaction) GetDocumentID() string {
	return hex.EncodeToString(pt.DocumentHash)
}

// GetUploadTimestamp returns the upload timestamp as a formatted string
func (pt *PDFTransaction) GetUploadTimestamp() string {
	return pt.UploadTimestamp.Format(time.RFC3339)
}

// GetDocumentInfo returns a summary of document information
func (pt *PDFTransaction) GetDocumentInfo() map[string]interface{} {
	return map[string]interface{}{
		"document_id":      pt.GetDocumentID(),
		"document_name":    pt.DocumentName,
		"document_size":    pt.DocumentSize,
		"document_type":    pt.DocumentType,
		"upload_timestamp": pt.GetUploadTimestamp(),
		"content_hash":     hex.EncodeToString(pt.ContentHash),
		"metadata":         pt.Metadata,
		"transaction_hash": hex.EncodeToString(pt.Hash),
	}
}

// IsValid checks if the PDF transaction is valid
func (pt *PDFTransaction) IsValid() error {
	if pt.DocumentHash == nil || len(pt.DocumentHash) == 0 {
		return fmt.Errorf("document hash is required")
	}
	
	if pt.DocumentSize == 0 {
		return fmt.Errorf("document size must be greater than 0")
	}
	
	if pt.DocumentName == "" {
		return fmt.Errorf("document name is required")
	}
	
	if pt.ContentHash == nil || len(pt.ContentHash) == 0 {
		return fmt.Errorf("content hash is required")
	}
	
	if pt.UploadTimestamp.IsZero() {
		return fmt.Errorf("upload timestamp is required")
	}
	
	// Verify that the transaction hash is not empty
	if pt.Hash == nil || len(pt.Hash) == 0 {
		return fmt.Errorf("transaction hash is required")
	}
	
	// Verify that the base transaction hash matches the PDF transaction hash
	if !bytes.Equal(pt.Transaction.Hash, pt.Hash) {
		return fmt.Errorf("base transaction hash mismatch")
	}
	
	return nil
}

// SerializePDF serializes the PDF transaction for storage
func (pt *PDFTransaction) SerializePDF() ([]byte, error) {
	if err := pt.IsValid(); err != nil {
		return nil, fmt.Errorf("invalid PDF transaction: %w", err)
	}
	
	// Serialize base transaction first
	baseTxData, err := pt.Transaction.Serialize()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize base transaction: %w", err)
	}
	
	// Create PDF-specific data structure
	pdfData := struct {
		BaseTransaction []byte        `json:"base_transaction"`
		DocumentHash    string        `json:"document_hash"`
		DocumentSize    uint64        `json:"document_size"`
		DocumentType    string        `json:"document_type"`
		DocumentName    string        `json:"document_name"`
		UploadTimestamp string        `json:"upload_timestamp"`
		ContentHash     string        `json:"content_hash"`
		Metadata        PDFMetadata   `json:"metadata"`
		Signature       string        `json:"signature,omitempty"`
		PublicKey       string        `json:"public_key,omitempty"`
	}{
		BaseTransaction: baseTxData,
		DocumentHash:    hex.EncodeToString(pt.DocumentHash),
		DocumentSize:    pt.DocumentSize,
		DocumentType:    pt.DocumentType,
		DocumentName:    pt.DocumentName,
		UploadTimestamp: pt.UploadTimestamp.Format(time.RFC3339Nano),
		ContentHash:     hex.EncodeToString(pt.ContentHash),
		Metadata:        pt.Metadata,
		Signature:       hex.EncodeToString(pt.Signature),
		PublicKey:       hex.EncodeToString(pt.PublicKey),
	}
	
	// Serialize to JSON
	data, err := json.Marshal(pdfData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal PDF transaction: %w", err)
	}
	
	return data, nil
}

// DeserializePDF deserializes a PDF transaction from storage
func DeserializePDF(data []byte) (*PDFTransaction, error) {
	var pdfData struct {
		BaseTransaction []byte        `json:"base_transaction"`
		DocumentHash    string        `json:"document_hash"`
		DocumentSize    uint64        `json:"document_size"`
		DocumentType    string        `json:"document_type"`
		DocumentName    string        `json:"document_name"`
		UploadTimestamp string        `json:"upload_timestamp"`
		ContentHash     string        `json:"content_hash"`
		Metadata        PDFMetadata   `json:"metadata"`
		Signature       string        `json:"signature"`
		PublicKey       string        `json:"public_key"`
	}
	
	if err := json.Unmarshal(data, &pdfData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal PDF transaction: %w", err)
	}
	
	// Deserialize base transaction
	baseTx := &block.Transaction{}
	if err := baseTx.Deserialize(pdfData.BaseTransaction); err != nil {
		return nil, fmt.Errorf("failed to deserialize base transaction: %w", err)
	}
	
	// Parse document hash
	documentHash, err := hex.DecodeString(pdfData.DocumentHash)
	if err != nil {
		return nil, fmt.Errorf("invalid document hash: %w", err)
	}
	
	// Parse content hash
	contentHash, err := hex.DecodeString(pdfData.ContentHash)
	if err != nil {
		return nil, fmt.Errorf("invalid content hash: %w", err)
	}
	
	// Parse upload timestamp
	uploadTimestamp, err := time.Parse(time.RFC3339Nano, pdfData.UploadTimestamp)
	if err != nil {
		return nil, fmt.Errorf("invalid upload timestamp: %w", err)
	}
	
	// Parse signature
	var signature []byte
	if pdfData.Signature != "" {
		signature, err = hex.DecodeString(pdfData.Signature)
		if err != nil {
			return nil, fmt.Errorf("invalid signature: %w", err)
		}
	}
	
	// Parse public key
	var publicKey []byte
	if pdfData.PublicKey != "" {
		publicKey, err = hex.DecodeString(pdfData.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("invalid public key: %w", err)
		}
	}
	
	// Create PDF transaction
	pdfTx := &PDFTransaction{
		Transaction:     baseTx,
		DocumentHash:    documentHash,
		DocumentSize:    pdfData.DocumentSize,
		DocumentType:    pdfData.DocumentType,
		DocumentName:    pdfData.DocumentName,
		UploadTimestamp: uploadTimestamp,
		ContentHash:     contentHash,
		Metadata:        pdfData.Metadata,
		Signature:       signature,
		PublicKey:       publicKey,
	}
	
	// Verify the transaction is valid
	if err := pdfTx.IsValid(); err != nil {
		return nil, fmt.Errorf("invalid PDF transaction: %w", err)
	}
	
	return pdfTx, nil
}
