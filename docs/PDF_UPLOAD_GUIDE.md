# PDF Upload Guide for adrenochain

## Overview

This guide explains how to use the PDF upload functionality in adrenochain, which allows you to store PDF documents on the blockchain with **immutability**, **cryptographic hashing**, and **blockchain timestamping**.

## 🎯 Key Features

- **🔒 Immutability**: Once uploaded, PDFs cannot be modified or deleted from the blockchain
- **🔐 Cryptographic Hashing**: SHA256 hashes ensure document integrity and prevent tampering
- **⏰ Blockchain Timestamping**: Each document gets a permanent timestamp from the blockchain
- **📊 Rich Metadata**: Store title, author, subject, keywords, and custom fields
- **🔍 Search & Discovery**: Find documents by content, metadata, or uploader
- **📱 REST API**: Simple HTTP endpoints for all operations
- **💾 Efficient Storage**: Smart caching and file system organization

## 🏗️ Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   PDF Upload    │───▶│  PDF Transaction │───▶│   Blockchain    │
│     API         │    │     (Block)      │    │     Storage     │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   PDF Storage   │    │   Content Hash   │    │   Metadata DB   │
│   (File System) │    │   (SHA256)       │    │   (JSON)        │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## 🚀 Quick Start

### 1. Start the adrenochain Node

```bash
cd cmd/gochain
go run main.go
```

### 2. Upload a PDF via API

```bash
curl -X POST http://localhost:8080/api/v1/pdfs/upload \
  -F "pdf_file=@document.pdf" \
  -F "title=My Important Document" \
  -F "author=John Doe" \
  -F "subject=Business Proposal" \
  -F "keywords=business,proposal,2024" \
  -F "description=Q1 business proposal for 2024"
```

### 3. Retrieve the PDF

```bash
# Get PDF info
curl http://localhost:8080/api/v1/pdfs/{document_id}/info

# Download the PDF
curl http://localhost:8080/api/v1/pdfs/{document_id}/download -o downloaded.pdf
```

## 📋 API Endpoints

### Upload PDF
```http
POST /api/v1/pdfs/upload
Content-Type: multipart/form-data

Form Fields:
- pdf_file: The PDF file (required)
- title: Document title (optional)
- author: Document author (optional)
- subject: Document subject (optional)
- keywords: Comma-separated keywords (optional)
- description: Document description (optional)
- tags: Comma-separated tags (optional)
- custom_field_name: Custom field value (optional)
```

**Response:**
```json
{
  "success": true,
  "document_id": "a1b2c3d4e5f6...",
  "document_name": "document.pdf",
  "document_size": 1024000,
  "content_hash": "sha256_hash_here",
  "block_hash": "block_hash_here",
  "transaction_hash": "tx_hash_here",
  "upload_timestamp": "2024-01-01T12:00:00Z",
  "message": "PDF uploaded successfully and stored on blockchain",
  "metadata": {
    "title": "My Important Document",
    "author": "John Doe",
    "subject": "Business Proposal",
    "keywords": ["business", "proposal", "2024"],
    "description": "Q1 business proposal for 2024"
  }
}
```

### Get PDF Info
```http
GET /api/v1/pdfs/{document_id}/info
```

**Response:**
```json
{
  "success": true,
  "document": {
    "document_id": "a1b2c3d4e5f6...",
    "document_name": "document.pdf",
    "document_size": 1024000,
    "document_type": "application/pdf",
    "upload_timestamp": "2024-01-01T12:00:00Z",
    "content_hash": "sha256_hash_here",
    "block_hash": "block_hash_here",
    "transaction_hash": "tx_hash_here",
    "uploader_id": "user123",
    "title": "My Important Document",
    "author": "John Doe",
    "subject": "Business Proposal",
    "keywords": ["business", "proposal", "2024"],
    "description": "Q1 business proposal for 2024",
    "access_count": 5,
    "last_accessed": "2024-01-01T15:30:00Z",
    "is_public": true
  },
  "message": "PDF metadata retrieved successfully"
}
```

### Download PDF
```http
GET /api/v1/pdfs/{document_id}/download
```

**Response Headers:**
```
Content-Type: application/pdf
Content-Disposition: attachment; filename="document.pdf"
Content-Length: 1024000
X-Document-ID: a1b2c3d4e5f6...
X-Content-Hash: sha256_hash_here
```

### List PDFs
```http
GET /api/v1/pdfs?limit=20&offset=0
```

**Response:**
```json
{
  "success": true,
  "documents": [...],
  "total": 25,
  "limit": 20,
  "offset": 0,
  "message": "PDFs retrieved successfully"
}
```

### Search PDFs
```http
GET /api/v1/pdfs/search?q=business&uploader_id=user123&min_size=1000000
```

**Query Parameters:**
- `q`: Search query text
- `uploader_id`: Filter by uploader
- `is_public`: Filter by public/private status
- `min_size`: Minimum file size in bytes
- `max_size`: Maximum file size in bytes

### Update PDF Metadata
```http
PUT /api/v1/pdfs/{document_id}
Content-Type: application/json

{
  "title": "Updated Title",
  "description": "Updated description",
  "tags": ["updated", "tags"]
}
```

### Delete PDF
```http
DELETE /api/v1/pdfs/{document_id}
```

### Get Storage Statistics
```http
GET /api/v1/pdfs/stats
```

**Response:**
```json
{
  "success": true,
  "stats": {
    "total_documents": 150,
    "total_size_mb": 45.67,
    "cached_documents": 25,
    "cache_size_mb": 12.34,
    "max_cache_size_mb": 100.0,
    "cache_hit_rate": 0.75
  },
  "message": "Storage statistics retrieved successfully"
}
```

## 🔐 Security Features

### Document Integrity Verification

Every PDF is cryptographically hashed using SHA256. You can verify document integrity:

```go
// Verify document hasn't been tampered with
content, metadata, err := pdfStorage.GetPDF(documentID)
if err != nil {
    log.Fatal(err)
}

// Calculate hash of retrieved content
calculatedHash := sha256.Sum256(content)
calculatedHashStr := hex.EncodeToString(calculatedHash[:])

// Compare with stored hash
if calculatedHashStr == metadata.ContentHash {
    fmt.Println("✅ Document integrity verified!")
} else {
    fmt.Println("❌ Document has been modified!")
}
```

### Blockchain Immutability

Once a PDF is stored on the blockchain:
- **Cannot be modified**: The content hash is permanently recorded
- **Cannot be deleted**: The transaction is immutable
- **Timestamped**: Permanent proof of when it was uploaded
- **Audit trail**: Complete history of all operations

## 💻 Go SDK Usage

### Initialize PDF Storage

```go
import (
    "github.com/palaseus/adrenochain/pkg/storage"
    "github.com/palaseus/adrenochain/pkg/block"
)

// Create PDF storage
config := storage.DefaultPDFStorageConfig()
pdfStorage, err := storage.NewPDFStorage(config)
if err != nil {
    log.Fatal(err)
}
```

### Upload a PDF

```go
// Read PDF file
pdfContent, err := os.ReadFile("document.pdf")
if err != nil {
    log.Fatal(err)
}

// Create metadata
metadata := block.PDFMetadata{
    Title:       "Important Document",
    Author:      "John Doe",
    Subject:     "Business Proposal",
    Keywords:    []string{"business", "proposal", "2024"},
    Description: "Q1 business proposal for 2024",
    Tags:        []string{"business", "proposal"},
    CustomFields: map[string]string{
        "department": "sales",
        "priority":   "high",
    },
}

// Mock blockchain data (in real usage, this comes from your blockchain)
blockHash := []byte("block_hash_here")
transactionHash := []byte("tx_hash_here")

// Store PDF
pdfMetadata, err := pdfStorage.StorePDF(
    pdfContent,
    "document.pdf",
    "user123",
    blockHash,
    transactionHash,
    metadata,
)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("PDF stored with ID: %s\n", pdfMetadata.DocumentID)
```

### Retrieve and Verify PDF

```go
// Get PDF content and metadata
content, metadata, err := pdfStorage.GetPDF(documentID)
if err != nil {
    log.Fatal(err)
}

// Verify integrity
calculatedHash := sha256.Sum256(content)
if hex.EncodeToString(calculatedHash[:]) == metadata.ContentHash {
    fmt.Println("Document integrity verified!")
} else {
    fmt.Println("Document integrity check failed!")
}

// Access metadata
fmt.Printf("Title: %s\n", metadata.Title)
fmt.Printf("Author: %s\n", metadata.Author)
fmt.Printf("Uploaded: %s\n", metadata.UploadTimestamp.Format(time.RFC3339))
fmt.Printf("Block Hash: %s\n", metadata.BlockHash)
```

### Search PDFs

```go
// Search by text
results, err := pdfStorage.SearchPDFs("business", map[string]interface{}{
    "uploader_id": "user123",
    "min_size":    uint64(1000000), // 1MB
})
if err != nil {
    log.Fatal(err)
}

for _, doc := range results {
    fmt.Printf("Found: %s by %s\n", doc.Title, doc.Author)
}
```

## 🧪 Testing

### Run PDF Tests

```bash
# Test PDF transaction functionality
go test ./pkg/block -v -run TestPDF

# Test PDF storage functionality
go test ./pkg/storage -v -run TestPDF

# Run all tests
./scripts/test_suite.sh
```

### Example Application

```bash
# Run the example application
cd examples
go run pdf_upload_example.go
```

## 🔧 Configuration

### PDF Storage Configuration

```go
config := &storage.PDFStorageConfig{
    BaseDir:      "./data/pdfs",           // Storage directory
    MaxCacheSize: 100 * 1024 * 1024,      // 100MB cache limit
}

pdfStorage, err := storage.NewPDFStorage(config)
```

### Environment Variables

```bash
# PDF storage configuration
PDF_STORAGE_DIR=./data/pdfs
PDF_MAX_CACHE_SIZE=100MB
PDF_MAX_FILE_SIZE=50MB

# Blockchain configuration
BLOCKCHAIN_NETWORK=mainnet
BLOCKCHAIN_RPC_URL=http://localhost:8545
```

## 📁 File Structure

```
data/pdfs/
├── content/           # PDF file content (by document ID)
│   ├── a1b2c3d4e5f6...
│   └── f6e5d4c3b2a1...
├── metadata/          # PDF metadata (JSON files)
│   ├── a1b2c3d4e5f6.json
│   └── f6e5d4c3b2a1.json
└── temp/              # Temporary upload directory
```

## 🚨 Limitations & Considerations

### File Size Limits
- **Maximum file size**: 50MB (configurable)
- **Recommended size**: < 10MB for optimal performance
- **Large files**: May take longer to process and store

### Storage Considerations
- **Disk space**: Ensure sufficient storage for all PDFs
- **Backup strategy**: Implement regular backups of PDF storage
- **Performance**: Large numbers of PDFs may impact search performance

### Blockchain Considerations
- **Transaction fees**: Each PDF upload requires a blockchain transaction
- **Block confirmation**: PDFs are immutable after block confirmation
- **Network congestion**: High network usage may delay confirmations

## 🔮 Future Enhancements

### Planned Features
- **Compression**: Automatic PDF compression to reduce storage
- **Encryption**: End-to-end encryption for private documents
- **Versioning**: Support for document versioning
- **Collaboration**: Multi-user document editing and sharing
- **OCR**: Automatic text extraction and indexing
- **Digital signatures**: Built-in digital signature verification

### Integration Possibilities
- **IPFS**: Integration with IPFS for distributed storage
- **Cloud storage**: Support for cloud storage backends
- **CDN**: Content delivery network integration
- **Analytics**: Document usage analytics and insights

## 🆘 Troubleshooting

### Common Issues

#### Upload Fails
```bash
# Check file size
ls -lh document.pdf

# Verify file is valid PDF
file document.pdf

# Check storage permissions
ls -la data/pdfs/
```

#### Integrity Check Fails
```bash
# Re-download and verify
curl -o test.pdf http://localhost:8080/api/v1/pdfs/{id}/download
sha256sum test.pdf

# Compare with stored hash
curl http://localhost:8080/api/v1/pdfs/{id}/info | jq '.document.content_hash'
```

#### Performance Issues
```bash
# Check storage statistics
curl http://localhost:8080/api/v1/pdfs/stats

# Monitor disk usage
df -h data/pdfs/

# Check cache hit rate
curl http://localhost:8080/api/v1/pdfs/stats | jq '.stats.cache_hit_rate'
```

### Debug Mode

```bash
# Enable debug logging
export PDF_DEBUG=true
export PDF_LOG_LEVEL=debug

# Run with verbose output
go run main.go -v
```

## 📞 Support

### Getting Help
- **Documentation**: [docs.adrenochain.io](https://docs.adrenochain.io)
- **GitHub Issues**: [github.com/adrenochain/adrenochain/issues](https://github.com/adrenochain/adrenochain/issues)
- **Discord**: [discord.gg/adrenochain](https://discord.gg/adrenochain)
- **Email**: pdf-support@adrenochain.io

### Contributing
We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Happy PDF uploading on the blockchain! 🚀📄⛓️**
