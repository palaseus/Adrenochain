package pdf

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimplePDFStorage_Creation(t *testing.T) {
	// Create a test directory
	testDir := "/tmp/pdf_storage_test"
	defer os.RemoveAll(testDir)

	// Test storage creation
	storage, err := NewSimplePDFStorage(testDir)
	assert.NoError(t, err)
	assert.NotNil(t, storage)

	// Verify directories were created
	assert.DirExists(t, testDir)
	assert.DirExists(t, filepath.Join(testDir, "content"))
	assert.DirExists(t, filepath.Join(testDir, "metadata"))
}

func TestSimplePDFStorage_StoreAndRetrieve(t *testing.T) {
	// Create a test directory
	testDir := "/tmp/pdf_storage_store_retrieve_test"
	defer os.RemoveAll(testDir)

	// Create storage
	storage, err := NewSimplePDFStorage(testDir)
	require.NoError(t, err)

	// Test data
	pdfData := []byte("Test PDF content for storage")
	documentName := "test_document.pdf"
	uploaderID := "test_user"
	metadata := PDFMetadata{
		Title:       "Test Document",
		Author:      "Test Author",
		Description: "A test document for storage testing",
		Keywords:    []string{"test", "storage", "pdf"},
		Tags:        []string{"test", "document"},
		CustomFields: map[string]string{
			"category": "test",
			"version":  "1.0",
		},
	}

	// Store PDF
	storedPDF, err := storage.StorePDF(pdfData, documentName, uploaderID, metadata)
	assert.NoError(t, err)
	assert.NotNil(t, storedPDF)
	assert.NotEmpty(t, storedPDF.DocumentID)

	// Retrieve PDF content
	retrievedData, retrievedStoredPDF, err := storage.GetPDF(storedPDF.DocumentID)
	assert.NoError(t, err)
	assert.Equal(t, pdfData, retrievedData)
	assert.NotNil(t, retrievedStoredPDF)

	// Verify retrieved metadata
	assert.Equal(t, documentName, retrievedStoredPDF.DocumentName)
	assert.Equal(t, uint64(len(pdfData)), retrievedStoredPDF.DocumentSize)
	assert.Equal(t, uploaderID, retrievedStoredPDF.UploaderID)
	assert.Equal(t, metadata.Title, retrievedStoredPDF.Title)
	assert.Equal(t, metadata.Author, retrievedStoredPDF.Author)
	assert.Equal(t, metadata.Description, retrievedStoredPDF.Description)
	assert.Equal(t, metadata.Keywords, retrievedStoredPDF.Keywords)
	assert.Equal(t, metadata.Tags, retrievedStoredPDF.Tags)
	assert.Equal(t, metadata.CustomFields, retrievedStoredPDF.CustomFields)
}

func TestSimplePDFStorage_ListPDFs(t *testing.T) {
	// Create a test directory
	testDir := "/tmp/pdf_storage_list_test"
	defer os.RemoveAll(testDir)

	// Create storage
	storage, err := NewSimplePDFStorage(testDir)
	require.NoError(t, err)

	// Store multiple PDFs
	storedPDFs := make([]*StoredPDF, 5)
	for i := 0; i < 5; i++ {
		pdfData := []byte(fmt.Sprintf("Test PDF content %d", i))
		documentName := fmt.Sprintf("test_document_%d.pdf", i)
		uploaderID := fmt.Sprintf("test_user_%d", i)
		metadata := PDFMetadata{
			Title:  fmt.Sprintf("Test Document %d", i),
			Author: fmt.Sprintf("Test Author %d", i),
		}

		storedPDF, err := storage.StorePDF(pdfData, documentName, uploaderID, metadata)
		assert.NoError(t, err)
		storedPDFs[i] = storedPDF
	}

	// List all PDFs
	allPDFs, err := storage.ListPDFs()
	assert.NoError(t, err)
	assert.Len(t, allPDFs, 5)

	// Verify all stored document IDs are in the list
	for _, storedPDF := range storedPDFs {
		found := false
		for _, listedPDF := range allPDFs {
			if listedPDF.DocumentID == storedPDF.DocumentID {
				found = true
				break
			}
		}
		assert.True(t, found, "Document ID %s should be in the list", storedPDF.DocumentID)
	}
}

func TestSimplePDFStorage_ErrorHandling(t *testing.T) {
	// Test with invalid directory
	invalidDir := "/invalid/path/that/does/not/exist"
	_, err := NewSimplePDFStorage(invalidDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid base directory path")

	// Create a test directory
	testDir := "/tmp/pdf_storage_error_test"
	defer os.RemoveAll(testDir)

	// Create storage
	storage, err := NewSimplePDFStorage(testDir)
	require.NoError(t, err)

	// Test retrieving non-existent PDF
	_, _, err = storage.GetPDF("non_existent_id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read metadata")
}

func TestSimplePDFStorage_EdgeCases(t *testing.T) {
	// Create a test directory
	testDir := "/tmp/pdf_storage_edge_test"
	defer os.RemoveAll(testDir)

	// Create storage
	storage, err := NewSimplePDFStorage(testDir)
	require.NoError(t, err)

	// Test with empty PDF content
	emptyData := []byte("")
	storedPDF, err := storage.StorePDF(emptyData, "empty.pdf", "test_user", PDFMetadata{})
	assert.NoError(t, err)
	assert.NotNil(t, storedPDF)
	assert.NotEmpty(t, storedPDF.DocumentID)

	retrievedData, _, err := storage.GetPDF(storedPDF.DocumentID)
	assert.NoError(t, err)
	assert.Equal(t, emptyData, retrievedData)

	// Test with very large PDF content
	largeData := make([]byte, 1024*1024) // 1MB
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	largeStoredPDF, err := storage.StorePDF(largeData, "large.pdf", "test_user", PDFMetadata{})
	assert.NoError(t, err)
	assert.NotNil(t, largeStoredPDF)
	assert.NotEmpty(t, largeStoredPDF.DocumentID)

	retrievedLargeData, _, err := storage.GetPDF(largeStoredPDF.DocumentID)
	assert.NoError(t, err)
	assert.Equal(t, largeData, retrievedLargeData)

	// Test with special characters in document name
	specialName := "test document with spaces & symbols!@#$%.pdf"
	specialStoredPDF, err := storage.StorePDF([]byte("special name test"), specialName, "test_user", PDFMetadata{})
	assert.NoError(t, err)
	assert.NotNil(t, specialStoredPDF)
	assert.NotEmpty(t, specialStoredPDF.DocumentID)

	_, retrievedMetadata, err := storage.GetPDF(specialStoredPDF.DocumentID)
	assert.NoError(t, err)
	assert.Equal(t, specialName, retrievedMetadata.DocumentName)
}

func TestSimplePDFStorage_ConcurrentAccess(t *testing.T) {
	// Create a test directory
	testDir := "/tmp/pdf_storage_concurrent_test"
	defer os.RemoveAll(testDir)

	// Create storage
	storage, err := NewSimplePDFStorage(testDir)
	require.NoError(t, err)

	// Test concurrent operations
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			defer func() { done <- true }()

			pdfData := []byte(fmt.Sprintf("Concurrent test PDF content %d", index))
			documentName := fmt.Sprintf("concurrent_test_%d.pdf", index)
			uploaderID := fmt.Sprintf("concurrent_user_%d", index)
			metadata := PDFMetadata{
				Title:  fmt.Sprintf("Concurrent Test Document %d", index),
				Author: fmt.Sprintf("Concurrent Author %d", index),
			}

			// Store PDF
			storedPDF, err := storage.StorePDF(pdfData, documentName, uploaderID, metadata)
			assert.NoError(t, err)

			// Retrieve PDF
			retrievedData, _, err := storage.GetPDF(storedPDF.DocumentID)
			assert.NoError(t, err)
			assert.Equal(t, pdfData, retrievedData)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all PDFs were stored
	allPDFs, err := storage.ListPDFs()
	assert.NoError(t, err)
	assert.Len(t, allPDFs, 10)
}

func TestSimplePDFStorage_Performance(t *testing.T) {
	// Create a test directory
	testDir := "/tmp/pdf_storage_performance_test"
	defer os.RemoveAll(testDir)

	// Create storage
	storage, err := NewSimplePDFStorage(testDir)
	require.NoError(t, err)

	// Test storage performance
	start := time.Now()

	// Store multiple PDFs
	storedPDFs := make([]*StoredPDF, 100)
	for i := 0; i < 100; i++ {
		pdfData := []byte(fmt.Sprintf("Performance test PDF content %d", i))
		documentName := fmt.Sprintf("performance_test_%d.pdf", i)
		uploaderID := fmt.Sprintf("perf_user_%d", i)
		metadata := PDFMetadata{
			Title:  fmt.Sprintf("Performance Test Document %d", i),
			Author: "Performance Tester",
		}

		storedPDF, err := storage.StorePDF(pdfData, documentName, uploaderID, metadata)
		assert.NoError(t, err)
		storedPDFs[i] = storedPDF
	}

	storageTime := time.Since(start)

	// Test retrieval performance
	start = time.Now()

	for i := 0; i < 100; i++ {
		_, _, err := storage.GetPDF(storedPDFs[i].DocumentID)
		assert.NoError(t, err)
	}

	retrievalTime := time.Since(start)

	// Verify performance is reasonable
	assert.Less(t, storageTime, 5*time.Second, "Storage should complete in less than 5 seconds")
	assert.Less(t, retrievalTime, 2*time.Second, "Retrieval should complete in less than 2 seconds")
}

func TestSimplePDFStorage_Integrity(t *testing.T) {
	// Create a test directory
	testDir := "/tmp/pdf_storage_integrity_test"
	defer os.RemoveAll(testDir)

	// Create storage
	storage, err := NewSimplePDFStorage(testDir)
	require.NoError(t, err)

	// Store a PDF
	pdfData := []byte("Integrity test PDF content")
	documentName := "integrity_test.pdf"
	uploaderID := "test_user"
	metadata := PDFMetadata{
		Title:  "Integrity Test Document",
		Author: "Test Author",
	}

	storedPDF, err := storage.StorePDF(pdfData, documentName, uploaderID, metadata)
	assert.NoError(t, err)

	// Verify content hash
	expectedHash := sha256.Sum256(pdfData)
	expectedHashStr := hex.EncodeToString(expectedHash[:])
	assert.Equal(t, expectedHashStr, storedPDF.ContentHash)

	// Verify document size
	assert.Equal(t, uint64(len(pdfData)), storedPDF.DocumentSize)

	// Verify upload timestamp is recent
	assert.True(t, time.Since(storedPDF.UploadTimestamp) < time.Minute)
}

func TestSimplePDFStorage_ListEmptyDirectory(t *testing.T) {
	// Create a test directory
	testDir := "/tmp/pdf_storage_empty_test"
	defer os.RemoveAll(testDir)

	// Create storage
	storage, err := NewSimplePDFStorage(testDir)
	require.NoError(t, err)

	// List PDFs in empty storage
	allPDFs, err := storage.ListPDFs()
	assert.NoError(t, err)
	assert.Len(t, allPDFs, 0)
}
