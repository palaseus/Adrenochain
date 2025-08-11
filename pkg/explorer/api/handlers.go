package api

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gochain/gochain/pkg/explorer/service"
	"github.com/gorilla/mux"
)

// ExplorerHandler handles HTTP requests for the blockchain explorer
type ExplorerHandler struct {
	explorerService service.ExplorerService
}

// NewExplorerHandler creates a new explorer handler
func NewExplorerHandler(explorerService service.ExplorerService) *ExplorerHandler {
	return &ExplorerHandler{
		explorerService: explorerService,
	}
}

// DashboardHandler handles requests for the main dashboard
func (h *ExplorerHandler) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dashboard, err := h.explorerService.GetDashboard(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get dashboard: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSONResponse(w, dashboard)
}

// BlockDetailsHandler handles requests for block details
func (h *ExplorerHandler) BlockDetailsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hashStr := vars["hash"]

	// Decode hex hash
	hash, err := hex.DecodeString(hashStr)
	if err != nil {
		http.Error(w, "Invalid block hash format", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	blockDetails, err := h.explorerService.GetBlockDetails(ctx, hash)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get block details: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSONResponse(w, blockDetails)
}

// TransactionDetailsHandler handles requests for transaction details
func (h *ExplorerHandler) TransactionDetailsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hashStr := vars["hash"]

	// Decode hex hash
	hash, err := hex.DecodeString(hashStr)
	if err != nil {
		http.Error(w, "Invalid transaction hash format", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	txDetails, err := h.explorerService.GetTransactionDetails(ctx, hash)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get transaction details: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSONResponse(w, txDetails)
}

// AddressDetailsHandler handles requests for address details
func (h *ExplorerHandler) AddressDetailsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	// Validate address format (basic check)
	if len(address) < 26 || len(address) > 35 {
		http.Error(w, "Invalid address format", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	addressDetails, err := h.explorerService.GetAddressDetails(ctx, address)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get address details: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSONResponse(w, addressDetails)
}

// BlocksHandler handles requests for block lists
func (h *ExplorerHandler) BlocksHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limit, offset := h.parsePaginationParams(r)

	ctx := r.Context()
	blocks, err := h.explorerService.GetBlocks(ctx, limit, offset)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get blocks: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSONResponse(w, map[string]interface{}{
		"blocks": blocks,
		"limit":  limit,
		"offset": offset,
		"count":  len(blocks),
	})
}

// TransactionsHandler handles requests for transaction lists
func (h *ExplorerHandler) TransactionsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limit, offset := h.parsePaginationParams(r)

	ctx := r.Context()
	transactions, err := h.explorerService.GetTransactions(ctx, limit, offset)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get transactions: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSONResponse(w, map[string]interface{}{
		"transactions": transactions,
		"limit":        limit,
		"offset":       offset,
		"count":        len(transactions),
	})
}

// SearchHandler handles search requests
func (h *ExplorerHandler) SearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Missing search query parameter 'q'", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	searchResult, err := h.explorerService.Search(ctx, query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSONResponse(w, searchResult)
}

// StatisticsHandler handles requests for blockchain statistics
func (h *ExplorerHandler) StatisticsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	statistics, err := h.explorerService.GetStatistics(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get statistics: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSONResponse(w, statistics)
}

// HealthHandler handles health check requests
func (h *ExplorerHandler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"service":   "blockchain-explorer",
		"timestamp": "2024-01-01T00:00:00Z", // Would be dynamic in production
	}

	h.writeJSONResponse(w, response)
}

// OptionsHandler handles OPTIONS requests for CORS preflight
func (h *ExplorerHandler) OptionsHandler(w http.ResponseWriter, r *http.Request) {
	// CORS headers are set by middleware
	w.WriteHeader(http.StatusOK)
}

// Helper methods

// writeJSONResponse writes a JSON response to the HTTP response writer
func (h *ExplorerHandler) writeJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// parsePaginationParams parses limit and offset from query parameters
func (h *ExplorerHandler) parsePaginationParams(r *http.Request) (limit, offset int) {
	// Default values
	limit = 20
	offset = 0

	// Parse limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			if parsed > 100 { // Cap at 100
				limit = 100
			} else {
				limit = parsed
			}
		}
	}

	// Parse offset
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return limit, offset
}

// validateHexHash validates that a string is a valid hex hash
func (h *ExplorerHandler) validateHexHash(hashStr string) ([]byte, error) {
	// Remove any "0x" prefix
	hashStr = strings.TrimPrefix(hashStr, "0x")

	// Check length (should be 64 characters for SHA256)
	if len(hashStr) != 64 {
		return nil, fmt.Errorf("hash must be 64 characters long")
	}

	// Decode hex
	hash, err := hex.DecodeString(hashStr)
	if err != nil {
		return nil, fmt.Errorf("invalid hex format: %v", err)
	}

	return hash, nil
}

// writeErrorResponse writes an error response in JSON format
func (h *ExplorerHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	errorResponse := map[string]interface{}{
		"error":   message,
		"status":  statusCode,
		"success": false,
	}

	jsonData, err := json.Marshal(errorResponse)
	if err != nil {
		http.Error(w, "Failed to marshal error response", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(statusCode)
	w.Write(jsonData)
}
