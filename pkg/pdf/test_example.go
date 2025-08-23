package pdf

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/palaseus/adrenochain/pkg/block"
)

// TestPDFFunctionality demonstrates the PDF transaction functionality
func TestPDFFunctionality() {
	fmt.Println("🚀 Testing PDF Transaction Functionality")
	fmt.Println("========================================")

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

	// Create inputs and outputs (empty for this test)
	inputs := []*block.TxInput{}
	outputs := []*block.TxOutput{}
	fee := uint64(1000)

	// Create PDF transaction
	pdfTx := NewPDFTransaction(documentContent, documentName, uploaderID, metadata, inputs, outputs, fee)

	// Display results
	fmt.Printf("✅ PDF Transaction created successfully!\n")
	fmt.Printf("   Document ID: %s\n", pdfTx.GetDocumentID())
	fmt.Printf("   Document Name: %s\n", pdfTx.DocumentName)
	fmt.Printf("   Document Size: %d bytes\n", pdfTx.DocumentSize)
	fmt.Printf("   Content Hash: %s\n", hex.EncodeToString(pdfTx.ContentHash))
	fmt.Printf("   Upload Timestamp: %s\n", pdfTx.GetUploadTimestamp())

	// Test integrity verification
	fmt.Println("\n🔍 Testing Document Integrity:")
	originalHash := sha256.Sum256(documentContent)
	originalHashStr := hex.EncodeToString(originalHash[:])
	
	if pdfTx.VerifyDocumentIntegrity(documentContent) {
		fmt.Printf("✅ Document integrity verified!\n")
		fmt.Printf("   Original Hash: %s\n", originalHashStr)
		fmt.Printf("   Stored Hash: %s\n", hex.EncodeToString(pdfTx.ContentHash))
	} else {
		fmt.Printf("❌ Document integrity check failed!\n")
	}

	// Test with modified content
	modifiedContent := append(documentContent, []byte("modified")...)
	if !pdfTx.VerifyDocumentIntegrity(modifiedContent) {
		fmt.Printf("✅ Integrity check correctly detects modifications!\n")
	} else {
		fmt.Printf("❌ Integrity check failed to detect modifications!\n")
	}

	// Test validation
	fmt.Println("\n🔒 Testing Transaction Validation:")
	if err := pdfTx.IsValid(); err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
	} else {
		fmt.Printf("✅ Transaction validation passed!\n")
	}

	// Display document info
	fmt.Println("\n📊 Document Information:")
	info := pdfTx.GetDocumentInfo()
	for key, value := range info {
		fmt.Printf("   %s: %v\n", key, value)
	}

	fmt.Println("\n🎉 PDF functionality test completed successfully!")
}
