package pdf

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/palaseus/adrenochain/pkg/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPDFTransaction_Creation(t *testing.T) {
	// Create test PDF content
	documentContent := []byte("This is a test PDF content for demonstration")
	documentName := "test_document.pdf"
	uploaderID := "test_user_123"

	// Create metadata
	metadata := PDFMetadata{
		Title:       "Test Document",
		Author:      "Test Author",
		Subject:     "PDF Testing",
		Keywords:    []string{"test", "pdf", "blockchain"},
		Description: "A test document for blockchain storage",
		Tags:        []string{"test", "document"},
		CustomFields: map[string]string{
			"category": "test",
			"version":  "1.0",
		},
	}

	// Create inputs and outputs
	inputs := []*block.TxInput{}
	outputs := []*block.TxOutput{}
	fee := uint64(1000)

	// Create PDF transaction
	pdfTx := NewPDFTransaction(documentContent, documentName, uploaderID, metadata, inputs, outputs, fee)

	// Test basic properties
	assert.NotNil(t, pdfTx)
	assert.Equal(t, documentName, pdfTx.DocumentName)
	assert.Equal(t, uint64(len(documentContent)), pdfTx.DocumentSize)
	assert.Equal(t, "application/pdf", pdfTx.DocumentType)
	assert.Equal(t, uploaderID, pdfTx.Metadata.UploaderID)
	assert.Equal(t, uint64(len(documentContent)), pdfTx.Metadata.FileSize)
	assert.False(t, pdfTx.UploadTimestamp.IsZero())

	// Test document hash calculation
	expectedHash := sha256.Sum256(documentContent)
	assert.Equal(t, expectedHash[:], pdfTx.DocumentHash)
	assert.Equal(t, expectedHash[:], pdfTx.ContentHash)

	// Test document ID
	expectedID := hex.EncodeToString(expectedHash[:])
	assert.Equal(t, expectedID, pdfTx.GetDocumentID())
}

func TestPDFTransaction_Validation(t *testing.T) {
	// Create valid PDF transaction
	documentContent := []byte("Test content")
	metadata := PDFMetadata{
		Title:  "Test Document",
		Author: "Test Author",
	}

	pdfTx := NewPDFTransaction(
		documentContent,
		"test.pdf",
		"user123",
		metadata,
		[]*block.TxInput{},
		[]*block.TxOutput{},
		1000,
	)

	// Test valid transaction
	err := pdfTx.IsValid()
	assert.NoError(t, err)

	// Test with empty document hash
	pdfTx.DocumentHash = nil
	err = pdfTx.IsValid()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document hash is required")

	// Reset and test with zero document size
	pdfTx = NewPDFTransaction(
		documentContent,
		"test.pdf",
		"user123",
		metadata,
		[]*block.TxInput{},
		[]*block.TxOutput{},
		1000,
	)
	pdfTx.DocumentSize = 0
	err = pdfTx.IsValid()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document size must be greater than 0")

	// Reset and test with empty document name
	pdfTx = NewPDFTransaction(
		documentContent,
		"test.pdf",
		"user123",
		metadata,
		[]*block.TxInput{},
		[]*block.TxOutput{},
		1000,
	)
	pdfTx.DocumentName = ""
	err = pdfTx.IsValid()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document name is required")
}

func TestPDFTransaction_IntegrityVerification(t *testing.T) {
	// Create test content
	originalContent := []byte("Original PDF content")
	modifiedContent := []byte("Modified PDF content")

	metadata := PDFMetadata{
		Title: "Test Document",
	}

	pdfTx := NewPDFTransaction(
		originalContent,
		"test.pdf",
		"user123",
		metadata,
		[]*block.TxInput{},
		[]*block.TxOutput{},
		1000,
	)

	// Test integrity verification with original content
	assert.True(t, pdfTx.VerifyDocumentIntegrity(originalContent))

	// Test integrity verification with modified content
	assert.False(t, pdfTx.VerifyDocumentIntegrity(modifiedContent))

	// Test with empty content
	assert.False(t, pdfTx.VerifyDocumentIntegrity([]byte{}))
}

func TestPDFTransaction_Serialization(t *testing.T) {
	// Create test PDF transaction
	documentContent := []byte("Test PDF content for serialization")
	metadata := PDFMetadata{
		Title:    "Serialization Test",
		Author:   "Test Author",
		Keywords: []string{"test", "serialization"},
		CustomFields: map[string]string{
			"test_field": "test_value",
		},
	}

	pdfTx := NewPDFTransaction(
		documentContent,
		"serialization_test.pdf",
		"user456",
		metadata,
		[]*block.TxInput{},
		[]*block.TxOutput{},
		2000,
	)

	// Test serialization
	serializedData, err := pdfTx.SerializePDF()
	require.NoError(t, err)
	assert.NotEmpty(t, serializedData)

	// Test deserialization
	deserializedTx, err := DeserializePDF(serializedData)
	require.NoError(t, err)
	assert.NotNil(t, deserializedTx)

	// Verify deserialized transaction matches original
	assert.Equal(t, pdfTx.DocumentName, deserializedTx.DocumentName)
	assert.Equal(t, pdfTx.DocumentSize, deserializedTx.DocumentSize)
	assert.Equal(t, pdfTx.DocumentType, deserializedTx.DocumentType)
	assert.Equal(t, pdfTx.DocumentHash, deserializedTx.DocumentHash)
	assert.Equal(t, pdfTx.ContentHash, deserializedTx.ContentHash)
	assert.Equal(t, pdfTx.Metadata.Title, deserializedTx.Metadata.Title)
	assert.Equal(t, pdfTx.Metadata.Author, deserializedTx.Metadata.Author)
	assert.Equal(t, pdfTx.Metadata.Keywords, deserializedTx.Metadata.Keywords)
	assert.Equal(t, pdfTx.Metadata.CustomFields, deserializedTx.Metadata.CustomFields)

	// Test that deserialized transaction is valid
	err = deserializedTx.IsValid()
	assert.NoError(t, err)
}

func TestPDFTransaction_WithRealPDFFile(t *testing.T) {
	// Read the test PDF file we created
	pdfData, err := os.ReadFile("../../test.pdf")
	require.NoError(t, err)
	require.NotEmpty(t, pdfData)

	// Create metadata for the real PDF
	metadata := PDFMetadata{
		Title:       "Real PDF Test Document",
		Author:      "Test Suite",
		Subject:     "Testing with real PDF file",
		Keywords:    []string{"real", "pdf", "test"},
		Description: "A test using an actual PDF file",
		PageCount:   1,
		Tags:        []string{"real", "pdf"},
		CustomFields: map[string]string{
			"source": "test_suite",
			"type":   "real_pdf",
		},
	}

	// Create PDF transaction with real PDF data
	pdfTx := NewPDFTransaction(
		pdfData,
		"test.pdf",
		"test_suite_user",
		metadata,
		[]*block.TxInput{},
		[]*block.TxOutput{},
		5000,
	)

	// Test basic properties
	assert.NotNil(t, pdfTx)
	assert.Equal(t, "test.pdf", pdfTx.DocumentName)
	assert.Equal(t, uint64(len(pdfData)), pdfTx.DocumentSize)
	assert.Equal(t, "application/pdf", pdfTx.DocumentType)
	assert.Equal(t, uint32(1), pdfTx.Metadata.PageCount)

	// Test document hash
	expectedHash := sha256.Sum256(pdfData)
	assert.Equal(t, expectedHash[:], pdfTx.DocumentHash)
	assert.Equal(t, expectedHash[:], pdfTx.ContentHash)

	// Test integrity verification
	assert.True(t, pdfTx.VerifyDocumentIntegrity(pdfData))

	// Test document info
	info := pdfTx.GetDocumentInfo()
	assert.NotNil(t, info)
	assert.Equal(t, "test.pdf", info["document_name"])
	assert.Equal(t, uint64(len(pdfData)), info["document_size"])
	assert.Equal(t, "application/pdf", info["document_type"])

	// Test serialization with real PDF
	serializedData, err := pdfTx.SerializePDF()
	require.NoError(t, err)
	assert.NotEmpty(t, serializedData)

	// Test deserialization
	deserializedTx, err := DeserializePDF(serializedData)
	require.NoError(t, err)
	assert.NotNil(t, deserializedTx)

	// Verify the deserialized transaction works with the original PDF data
	assert.True(t, deserializedTx.VerifyDocumentIntegrity(pdfData))
}

func TestPDFTransaction_GetDocumentInfo(t *testing.T) {
	documentContent := []byte("Test content for document info")
	metadata := PDFMetadata{
		Title:       "Document Info Test",
		Author:      "Info Tester",
		Description: "Testing document info retrieval",
	}

	pdfTx := NewPDFTransaction(
		documentContent,
		"info_test.pdf",
		"info_user",
		metadata,
		[]*block.TxInput{},
		[]*block.TxOutput{},
		1000,
	)

	info := pdfTx.GetDocumentInfo()

	// Test that all expected fields are present
	assert.Contains(t, info, "document_id")
	assert.Contains(t, info, "document_name")
	assert.Contains(t, info, "document_size")
	assert.Contains(t, info, "document_type")
	assert.Contains(t, info, "upload_timestamp")
	assert.Contains(t, info, "content_hash")
	assert.Contains(t, info, "metadata")
	assert.Contains(t, info, "transaction_hash")

	// Test specific values
	assert.Equal(t, "info_test.pdf", info["document_name"])
	assert.Equal(t, uint64(len(documentContent)), info["document_size"])
	assert.Equal(t, "application/pdf", info["document_type"])

	// Test metadata is included
	metadataInfo, ok := info["metadata"].(PDFMetadata)
	assert.True(t, ok)
	assert.Equal(t, "Document Info Test", metadataInfo.Title)
	assert.Equal(t, "Info Tester", metadataInfo.Author)
}

func TestPDFTransaction_GetUploadTimestamp(t *testing.T) {
	documentContent := []byte("Timestamp test content")
	metadata := PDFMetadata{Title: "Timestamp Test"}

	pdfTx := NewPDFTransaction(
		documentContent,
		"timestamp_test.pdf",
		"timestamp_user",
		metadata,
		[]*block.TxInput{},
		[]*block.TxOutput{},
		1000,
	)

	timestampStr := pdfTx.GetUploadTimestamp()
	assert.NotEmpty(t, timestampStr)

	// Test that the timestamp can be parsed
	parsedTime, err := time.Parse(time.RFC3339, timestampStr)
	assert.NoError(t, err)
	assert.False(t, parsedTime.IsZero())

	// Test that the timestamp is recent (within last minute)
	now := time.Now()
	diff := now.Sub(parsedTime)
	assert.True(t, diff < time.Minute)
}

func TestPDFTransaction_CalculatePDFHash(t *testing.T) {
	documentContent := []byte("Hash calculation test content")
	metadata := PDFMetadata{
		Title:  "Hash Test",
		Author: "Hash Tester",
	}

	pdfTx := NewPDFTransaction(
		documentContent,
		"hash_test.pdf",
		"hash_user",
		metadata,
		[]*block.TxInput{},
		[]*block.TxOutput{},
		1000,
	)

	// Test that hash is calculated
	hash1 := pdfTx.CalculatePDFHash()
	assert.NotNil(t, hash1)
	assert.Equal(t, 32, len(hash1)) // SHA256 hash length

	// Test that hash is consistent
	hash2 := pdfTx.CalculatePDFHash()
	assert.Equal(t, hash1, hash2)

	// Test that hash changes when content changes
	pdfTx.DocumentName = "modified_name.pdf"
	hash3 := pdfTx.CalculatePDFHash()
	assert.NotEqual(t, hash1, hash3)
}

func TestPDFTransaction_EdgeCases(t *testing.T) {
	// Test with empty content
	emptyContent := []byte("")
	metadata := PDFMetadata{Title: "Empty Test"}

	pdfTx := NewPDFTransaction(
		emptyContent,
		"empty.pdf",
		"empty_user",
		metadata,
		[]*block.TxInput{},
		[]*block.TxOutput{},
		1000,
	)

	assert.Equal(t, uint64(0), pdfTx.DocumentSize)
	assert.True(t, pdfTx.VerifyDocumentIntegrity(emptyContent))

	// Test with very large content (simulate large PDF)
	largeContent := bytes.Repeat([]byte("A"), 1024*1024) // 1MB
	largeMetadata := PDFMetadata{
		Title:       "Large Document Test",
		Description: "Testing with large content",
	}

	largePdfTx := NewPDFTransaction(
		largeContent,
		"large.pdf",
		"large_user",
		largeMetadata,
		[]*block.TxInput{},
		[]*block.TxOutput{},
		10000,
	)

	assert.Equal(t, uint64(len(largeContent)), largePdfTx.DocumentSize)
	assert.True(t, largePdfTx.VerifyDocumentIntegrity(largeContent))

	// Test with special characters in document name
	specialName := "test document with spaces & symbols!@#$%.pdf"
	specialPdfTx := NewPDFTransaction(
		[]byte("Special name test"),
		specialName,
		"special_user",
		PDFMetadata{Title: "Special Name Test"},
		[]*block.TxInput{},
		[]*block.TxOutput{},
		1000,
	)

	assert.Equal(t, specialName, specialPdfTx.DocumentName)
}

func TestPDFTransaction_ConcurrentAccess(t *testing.T) {
	documentContent := []byte("Concurrent access test content")
	metadata := PDFMetadata{Title: "Concurrent Test"}

	pdfTx := NewPDFTransaction(
		documentContent,
		"concurrent_test.pdf",
		"concurrent_user",
		metadata,
		[]*block.TxInput{},
		[]*block.TxOutput{},
		1000,
	)

	// Test concurrent access to various methods
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			// Test multiple concurrent operations
			_ = pdfTx.GetDocumentID()
			_ = pdfTx.GetUploadTimestamp()
			_ = pdfTx.GetDocumentInfo()
			_ = pdfTx.VerifyDocumentIntegrity(documentContent)
			_ = pdfTx.CalculatePDFHash()
			_ = pdfTx.IsValid()
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify the transaction is still valid after concurrent access
	err := pdfTx.IsValid()
	assert.NoError(t, err)
}

func TestPDFTransaction_InvalidSerialization(t *testing.T) {
	// Test deserialization with invalid data
	invalidData := []byte("invalid json data")
	_, err := DeserializePDF(invalidData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal PDF transaction")

	// Test deserialization with empty data
	_, err = DeserializePDF([]byte{})
	assert.Error(t, err)

	// Test deserialization with nil data
	_, err = DeserializePDF(nil)
	assert.Error(t, err)
}

func TestPDFMetadata_Serialization(t *testing.T) {
	// Test metadata with all fields populated
	metadata := PDFMetadata{
		Title:       "Complete Metadata Test",
		Author:      "Metadata Tester",
		Subject:     "Testing complete metadata",
		Keywords:    []string{"complete", "metadata", "test"},
		PageCount:   42,
		FileSize:    12345,
		UploaderID:  "metadata_user",
		Tags:        []string{"complete", "test"},
		Description: "A complete metadata test",
		CustomFields: map[string]string{
			"field1": "value1",
			"field2": "value2",
			"field3": "value3",
		},
	}

	documentContent := []byte("Complete metadata test content")
	pdfTx := NewPDFTransaction(
		documentContent,
		"complete_metadata_test.pdf",
		"complete_user",
		metadata,
		[]*block.TxInput{},
		[]*block.TxOutput{},
		1000,
	)

	// Test serialization and deserialization
	serializedData, err := pdfTx.SerializePDF()
	require.NoError(t, err)

	deserializedTx, err := DeserializePDF(serializedData)
	require.NoError(t, err)

	// Verify all metadata fields are preserved
	assert.Equal(t, metadata.Title, deserializedTx.Metadata.Title)
	assert.Equal(t, metadata.Author, deserializedTx.Metadata.Author)
	assert.Equal(t, metadata.Subject, deserializedTx.Metadata.Subject)
	assert.Equal(t, metadata.Keywords, deserializedTx.Metadata.Keywords)
	assert.Equal(t, metadata.PageCount, deserializedTx.Metadata.PageCount)
	// Note: FileSize and UploaderID are set by NewPDFTransaction, so check the actual values
	assert.Equal(t, uint64(len(documentContent)), deserializedTx.Metadata.FileSize)
	assert.Equal(t, "complete_user", deserializedTx.Metadata.UploaderID)
	assert.Equal(t, metadata.Tags, deserializedTx.Metadata.Tags)
	assert.Equal(t, metadata.Description, deserializedTx.Metadata.Description)
	assert.Equal(t, metadata.CustomFields, deserializedTx.Metadata.CustomFields)
}
