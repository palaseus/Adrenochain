package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/palaseus/adrenochain/pkg/block"
	"github.com/palaseus/adrenochain/pkg/storage"
)

func main() {
	fmt.Println("🚀 adrenochain PDF Upload Example")
	fmt.Println("==================================")

	// Create PDF storage
	pdfStorage, err := storage.NewPDFStorage(storage.DefaultPDFStorageConfig())
	if err != nil {
		log.Fatalf("Failed to create PDF storage: %v", err)
	}

	// Example 1: Upload a simple PDF document
	fmt.Println("\n📄 Example 1: Uploading a simple PDF document")
	example1(pdfStorage)

	// Example 2: Upload with rich metadata
	fmt.Println("\n📄 Example 2: Uploading with rich metadata")
	example2(pdfStorage)

	// Example 3: Verify document integrity
	fmt.Println("\n🔍 Example 3: Verifying document integrity")
	example3(pdfStorage)

	// Example 4: List and search documents
	fmt.Println("\n📋 Example 4: Listing and searching documents")
	example4(pdfStorage)

	// Example 5: Get storage statistics
	fmt.Println("\n📊 Example 5: Storage statistics")
	example5(pdfStorage)

	fmt.Println("\n✅ PDF upload examples completed successfully!")
}

func example1(pdfStorage *storage.PDFStorage) {
	// Create a simple PDF content (in real usage, this would be an actual PDF file)
	pdfContent := []byte("%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n/Pages 2 0 R\n>>\nendobj\n2 0 obj\n<<\n/Type /Pages\n/Kids [3 0 R]\n/Count 1\n>>\nendobj\n3 0 obj\n<<\n/Type /Page\n/Parent 2 0 R\n/MediaBox [0 0 612 792]\n/Contents 4 0 R\n>>\nendobj\n4 0 obj\n<<\n/Length 44\n>>\nstream\nBT\n/F1 12 Tf\n72 720 Td\n(Hello, Blockchain!) Tj\nET\nendstream\nendobj\nxref\n0 5\n0000000000 65535 f \n0000000009 00000 n \n0000000058 00000 n \n0000000115 00000 n \n0000000204 00000 n \ntrailer\n<<\n/Size 5\n/Root 1 0 R\n>>\nstartxref\n297\n%%EOF")

	// Create metadata
	metadata := block.PDFMetadata{
		Title:       "Simple Test Document",
		Author:      "Blockchain User",
		Subject:     "PDF Upload Test",
		Description: "A simple PDF document for testing blockchain storage",
		Keywords:    []string{"test", "pdf", "blockchain", "simple"},
		Tags:        []string{"test", "document"},
		CustomFields: map[string]string{
			"category": "example",
			"version":  "1.0",
		},
	}

	// Create mock blockchain data
	blockHash := []byte("example_block_hash_001")
	transactionHash := []byte("example_tx_hash_001")

	// Store PDF
	pdfMetadata, err := pdfStorage.StorePDF(
		pdfContent,
		"simple_test.pdf",
		"user123",
		blockHash,
		transactionHash,
		metadata,
	)
	if err != nil {
		log.Printf("Failed to store PDF: %v", err)
		return
	}

	fmt.Printf("✅ PDF stored successfully!\n")
	fmt.Printf("   Document ID: %s\n", pdfMetadata.DocumentID)
	fmt.Printf("   Document Name: %s\n", pdfMetadata.DocumentName)
	fmt.Printf("   Document Size: %d bytes\n", pdfMetadata.DocumentSize)
	fmt.Printf("   Content Hash: %s\n", pdfMetadata.ContentHash)
	fmt.Printf("   Block Hash: %s\n", pdfMetadata.BlockHash)
	fmt.Printf("   Transaction Hash: %s\n", pdfMetadata.TransactionHash)
	fmt.Printf("   Upload Timestamp: %s\n", pdfMetadata.UploadTimestamp.Format(time.RFC3339))
}

func example2(pdfStorage *storage.PDFStorage) {
	// Create a more complex PDF content
	pdfContent := []byte("%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n/Pages 2 0 R\n/Info 5 0 R\n>>\nendobj\n2 0 obj\n<<\n/Type /Pages\n/Kids [3 0 R]\n/Count 1\n>>\nendobj\n3 0 obj\n<<\n/Type /Page\n/Parent 2 0 R\n/MediaBox [0 0 612 792]\n/Contents 4 0 R\n>>\nendobj\n4 0 obj\n<<\n/Length 120\n>>\nstream\nBT\n/F1 16 Tf\n72 720 Td\n(Rich Metadata Example) Tj\n/F1 12 Tf\n72 680 Td\n(Author: John Doe) Tj\n72 660 Td\n(Subject: Advanced PDF Testing) Tj\n72 640 Td\n(Keywords: blockchain, pdf, metadata, advanced) Tj\nET\nendstream\nendobj\n5 0 obj\n<<\n/Title (Rich Metadata PDF)\n/Author (John Doe)\n/Subject (Advanced PDF Testing)\n/Keywords (blockchain pdf metadata advanced)\n/Creator (adrenochain)\n/Producer (adrenochain v1.0)\n/CreationDate (D:20240101000000)\n>>\nendobj\nxref\n0 6\n0000000000 65535 f \n0000000009 00000 n \n0000000058 00000 n \n0000000115 00000 n \n0000000204 00000 n \n0000000354 00000 n \ntrailer\n<<\n/Size 6\n/Root 1 0 R\n/Info 5 0 R\n>>\nstartxref\n474\n%%EOF")

	// Create rich metadata
	metadata := block.PDFMetadata{
		Title:       "Rich Metadata Example",
		Author:      "John Doe",
		Subject:     "Advanced PDF Testing with Blockchain",
		Description: "This document demonstrates rich metadata capabilities for PDF storage on the blockchain",
		Keywords:    []string{"blockchain", "pdf", "metadata", "advanced", "testing", "example"},
		Tags:        []string{"advanced", "metadata", "example", "blockchain"},
		PageCount:   1,
		CustomFields: map[string]string{
			"category":     "advanced",
			"version":      "2.0",
			"department":   "engineering",
			"project":      "blockchain-pdf",
			"reviewer":     "Jane Smith",
			"approval_date": "2024-01-01",
		},
	}

	// Create mock blockchain data
	blockHash := []byte("example_block_hash_002")
	transactionHash := []byte("example_tx_hash_002")

	// Store PDF
	pdfMetadata, err := pdfStorage.StorePDF(
		pdfContent,
		"rich_metadata_example.pdf",
		"john_doe",
		blockHash,
		transactionHash,
		metadata,
	)
	if err != nil {
		log.Printf("Failed to store PDF: %v", err)
		return
	}

	fmt.Printf("✅ Rich metadata PDF stored successfully!\n")
	fmt.Printf("   Document ID: %s\n", pdfMetadata.DocumentID)
	fmt.Printf("   Title: %s\n", pdfMetadata.Title)
	fmt.Printf("   Author: %s\n", pdfMetadata.Author)
	fmt.Printf("   Subject: %s\n", pdfMetadata.Subject)
	fmt.Printf("   Keywords: %v\n", pdfMetadata.Keywords)
	fmt.Printf("   Custom Fields: %v\n", pdfMetadata.CustomFields)
}

func example3(pdfStorage *storage.PDFStorage) {
	// Get the first document to verify integrity
	documents, err := pdfStorage.ListPDFs(1, 0)
	if err != nil || len(documents) == 0 {
		log.Printf("No documents found for integrity test")
		return
	}

	documentID := documents[0].DocumentID
	fmt.Printf("🔍 Testing integrity for document: %s\n", documentID)

	// Retrieve the document
	content, metadata, err := pdfStorage.GetPDF(documentID)
	if err != nil {
		log.Printf("Failed to retrieve PDF: %v", err)
		return
	}

	// Verify content hash
	calculatedHash := sha256.Sum256(content)
	calculatedHashStr := hex.EncodeToString(calculatedHash[:])
	storedHashStr := metadata.ContentHash

	if calculatedHashStr == storedHashStr {
		fmt.Printf("✅ Document integrity verified!\n")
		fmt.Printf("   Calculated Hash: %s\n", calculatedHashStr)
		fmt.Printf("   Stored Hash: %s\n", storedHashStr)
	} else {
		fmt.Printf("❌ Document integrity check failed!\n")
		fmt.Printf("   Calculated Hash: %s\n", calculatedHashStr)
		fmt.Printf("   Stored Hash: %s\n", storedHashStr)
	}

	// Test with modified content
	modifiedContent := append(content, []byte("modified")...)
	modifiedHash := sha256.Sum256(modifiedContent)
	modifiedHashStr := hex.EncodeToString(modifiedHash[:])

	if modifiedHashStr != storedHashStr {
		fmt.Printf("✅ Integrity check correctly detects modifications!\n")
		fmt.Printf("   Modified Hash: %s\n", modifiedHashStr)
		fmt.Printf("   Original Hash: %s\n", storedHashStr)
	} else {
		fmt.Printf("❌ Integrity check failed to detect modifications!\n")
	}
}

func example4(pdfStorage *storage.PDFStorage) {
	// List all documents
	fmt.Println("📋 Listing all documents:")
	documents, err := pdfStorage.ListPDFs(10, 0)
	if err != nil {
		log.Printf("Failed to list documents: %v", err)
		return
	}

	for i, doc := range documents {
		fmt.Printf("   %d. %s (ID: %s)\n", i+1, doc.DocumentName, doc.DocumentID)
		fmt.Printf("      Size: %d bytes, Uploaded: %s\n", doc.DocumentSize, doc.UploadTimestamp.Format(time.RFC3339))
	}

	// Search for documents
	fmt.Println("\n🔍 Searching for documents with 'test' in title:")
	searchResults, err := pdfStorage.SearchPDFs("test", map[string]interface{}{
		"min_size": uint64(100), // Only documents larger than 100 bytes
	})
	if err != nil {
		log.Printf("Failed to search documents: %v", err)
		return
	}

	for i, doc := range searchResults {
		fmt.Printf("   %d. %s - %s\n", i+1, doc.Title, doc.DocumentName)
	}

	// Search by uploader
	fmt.Println("\n🔍 Searching for documents by uploader 'user123':")
	uploaderResults, err := pdfStorage.SearchPDFs("", map[string]interface{}{
		"uploader_id": "user123",
	})
	if err != nil {
		log.Printf("Failed to search by uploader: %v", err)
		return
	}

	for i, doc := range uploaderResults {
		fmt.Printf("   %d. %s by %s\n", i+1, doc.DocumentName, doc.UploaderID)
	}
}

func example5(pdfStorage *storage.PDFStorage) {
	// Get storage statistics
	stats, err := pdfStorage.GetStorageStats()
	if err != nil {
		log.Printf("Failed to get storage stats: %v", err)
		return
	}

	fmt.Printf("📊 Storage Statistics:\n")
	fmt.Printf("   Total Documents: %v\n", stats["total_documents"])
	fmt.Printf("   Total Size: %.2f MB\n", stats["total_size_mb"])
	fmt.Printf("   Cached Documents: %v\n", stats["cached_documents"])
	fmt.Printf("   Cache Size: %.2f MB\n", stats["cache_size_mb"])
	fmt.Printf("   Max Cache Size: %.2f MB\n", stats["max_cache_size_mb"])
	fmt.Printf("   Cache Hit Rate: %.2f%%\n", stats["cache_hit_rate"].(float64)*100)
}

// Helper function to create a mock PDF file for testing
func createMockPDFFile(filename string, content string) error {
	// Create a simple PDF structure
	pdfContent := fmt.Sprintf(`%%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj
2 0 obj
<<
/Type /Pages
/Kids [3 0 R]
/Count 1
>>
endobj
3 0 obj
<<
/Type /Page
/Parent 2 0 R
/MediaBox [0 0 612 792]
/Contents 4 0 R
>>
endobj
4 0 obj
<<
/Length %d
>>
stream
BT
/F1 12 Tf
72 720 Td
(%s) Tj
ET
endstream
endobj
xref
0 5
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
0000000115 00000 n 
0000000204 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
>>
startxref
297
%%EOF`, len(content)+20, content)

	return os.WriteFile(filename, []byte(pdfContent), 0644)
}
