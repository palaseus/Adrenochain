package web

import (
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gochain/gochain/pkg/explorer/service"
	"github.com/gorilla/mux"
)

// Rate limiter for API endpoints
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// IsAllowed checks if a request is allowed
func (rl *RateLimiter) IsAllowed(clientIP string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Clean old requests
	if times, exists := rl.requests[clientIP]; exists {
		var validTimes []time.Time
		for _, t := range times {
			if t.After(windowStart) {
				validTimes = append(validTimes, t)
			}
		}
		rl.requests[clientIP] = validTimes
	}

	// Check if limit exceeded
	if len(rl.requests[clientIP]) >= rl.limit {
		return false
	}

	// Add current request
	rl.requests[clientIP] = append(rl.requests[clientIP], now)
	return true
}

// WebHandler handles web page requests for the blockchain explorer
type WebHandler struct {
	explorerService service.ExplorerService
	templates       *Templates
	rateLimiter     *RateLimiter
}

// NewWebHandler creates a new web handler
func NewWebHandler(explorerService service.ExplorerService, templates *Templates) *WebHandler {
	return &WebHandler{
		explorerService: explorerService,
		templates:       templates,
		rateLimiter:     NewRateLimiter(100, time.Minute), // 100 requests per minute
	}
}

// HomeHandler handles the main homepage
func (h *WebHandler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dashboard, err := h.explorerService.GetDashboard(ctx)
	if err != nil {
		h.renderError(w, "Failed to load dashboard", err)
		return
	}

	data := map[string]interface{}{
		"Title":     "GoChain Explorer",
		"Dashboard": dashboard,
	}

	h.templates.Render(w, "home.html", data)
}

// BlockListHandler handles the blocks list page
func (h *WebHandler) BlockListHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Parse pagination parameters
	limit, offset := h.parsePaginationParams(r)
	
	blocks, err := h.explorerService.GetBlocks(ctx, limit, offset)
	if err != nil {
		h.renderError(w, "Failed to load blocks", err)
		return
	}

	// Get total count for pagination
	stats, err := h.explorerService.GetStatistics(ctx)
	if err != nil {
		h.renderError(w, "Failed to load statistics", err)
		return
	}

	data := map[string]interface{}{
		"Title":      "Blocks - GoChain Explorer",
		"Blocks":     blocks,
		"Pagination": h.createPagination(limit, offset, int(stats.Blockchain.TotalBlocks)),
	}

	h.templates.Render(w, "blocks.html", data)
}

// BlockDetailHandler handles individual block detail pages
func (h *WebHandler) BlockDetailHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hashStr := vars["hash"]

	// Decode hex hash
	hash, err := hex.DecodeString(hashStr)
	if err != nil {
		h.renderError(w, "Invalid block hash format", err)
		return
	}

	ctx := r.Context()
	blockDetails, err := h.explorerService.GetBlockDetails(ctx, hash)
	if err != nil {
		h.renderError(w, "Failed to load block details", err)
		return
	}

	data := map[string]interface{}{
		"Title": "Block Details - GoChain Explorer",
		"Block": blockDetails,
	}

	h.templates.Render(w, "block_detail.html", data)
}

// TransactionListHandler handles the transactions list page
func (h *WebHandler) TransactionListHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Parse pagination parameters
	limit, offset := h.parsePaginationParams(r)
	
	transactions, err := h.explorerService.GetTransactions(ctx, limit, offset)
	if err != nil {
		h.renderError(w, "Failed to load transactions", err)
		return
	}

	// Get total count for pagination
	stats, err := h.explorerService.GetStatistics(ctx)
	if err != nil {
		h.renderError(w, "Failed to load statistics", err)
		return
	}

	data := map[string]interface{}{
		"Title":        "Transactions - GoChain Explorer",
		"Transactions": transactions,
		"Pagination":   h.createPagination(limit, offset, int(stats.Blockchain.TotalTransactions)),
	}

	h.templates.Render(w, "transactions.html", data)
}

// TransactionDetailHandler handles individual transaction detail pages
func (h *WebHandler) TransactionDetailHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hashStr := vars["hash"]

	// Decode hex hash
	hash, err := hex.DecodeString(hashStr)
	if err != nil {
		h.renderError(w, "Invalid transaction hash format", err)
		return
	}

	ctx := r.Context()
	txDetails, err := h.explorerService.GetTransactionDetails(ctx, hash)
	if err != nil {
		h.renderError(w, "Failed to load transaction details", err)
		return
	}

	data := map[string]interface{}{
		"Title":       "Transaction Details - GoChain Explorer",
		"Transaction": txDetails,
	}

	h.templates.Render(w, "transaction_detail.html", data)
}

// AddressDetailHandler handles individual address detail pages
func (h *WebHandler) AddressDetailHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	// Validate address format (basic check)
	if len(address) < 26 || len(address) > 35 {
		h.renderError(w, "Invalid address format", nil)
		return
	}

	ctx := r.Context()
	addressDetails, err := h.explorerService.GetAddressDetails(ctx, address)
	if err != nil {
		h.renderError(w, "Failed to load address details", err)
		return
	}

	data := map[string]interface{}{
		"Title":   "Address Details - GoChain Explorer",
		"Address": addressDetails,
	}

	h.templates.Render(w, "address_detail.html", data)
}

// SearchHandler handles search requests
func (h *WebHandler) SearchHandler(w http.ResponseWriter, r *http.Request) {
	// Apply rate limiting for search
	clientIP := h.getClientIP(r)
	if !h.rateLimiter.IsAllowed(clientIP) {
		http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		// Show search form
		data := map[string]interface{}{
			"Title": "Search - GoChain Explorer",
		}
		h.templates.Render(w, "search.html", data)
		return
	}

	// Validate search query
	if !h.isValidSearchQuery(query) {
		h.renderError(w, "Invalid search query format", nil)
		return
	}

	ctx := r.Context()
	searchResult, err := h.explorerService.Search(ctx, query)
	if err != nil {
		h.renderError(w, "Search failed", err)
		return
	}

	data := map[string]interface{}{
		"Title":       "Search Results - GoChain Explorer",
		"Query":       query,
		"SearchResult": searchResult,
	}

	h.templates.Render(w, "search_results.html", data)
}

// APIHandler redirects API requests to the API endpoints
func (h *WebHandler) APIHandler(w http.ResponseWriter, r *http.Request) {
	// Redirect API requests to the API endpoints
	http.Redirect(w, r, "/api/v1"+r.URL.Path, http.StatusMovedPermanently)
}

// StaticFileHandler serves static files
func (h *WebHandler) StaticFileHandler(w http.ResponseWriter, r *http.Request) {
	// Serve static files from the static directory
	http.ServeFile(w, r, "pkg/explorer/web/static"+r.URL.Path)
}

// Helper methods

func (h *WebHandler) renderError(w http.ResponseWriter, message string, err error) {
	data := map[string]interface{}{
		"Title":   "Error - GoChain Explorer",
		"Message": message,
		"Error":   err,
	}
	
	w.WriteHeader(http.StatusInternalServerError)
	h.templates.Render(w, "error.html", data)
}

func (h *WebHandler) parsePaginationParams(r *http.Request) (limit, offset int) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	
	limit = 20 // Default limit
	offset = 0 // Default offset
	
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	
	return limit, offset
}

func (h *WebHandler) createPagination(limit, offset, total int) map[string]interface{} {
	totalPages := (total + limit - 1) / limit
	currentPage := (offset / limit) + 1
	
	hasPrev := currentPage > 1
	hasNext := currentPage < totalPages
	
	prevOffset := offset - limit
	if prevOffset < 0 {
		prevOffset = 0
	}
	
	nextOffset := offset + limit
	if nextOffset >= total {
		nextOffset = total - limit
		if nextOffset < 0 {
			nextOffset = 0
		}
	}
	
	return map[string]interface{}{
		"CurrentPage":  currentPage,
		"TotalPages":   totalPages,
		"TotalItems":   total,
		"Limit":        limit,
		"Offset":       offset,
		"HasPrev":      hasPrev,
		"HasNext":      hasNext,
		"PrevOffset":   prevOffset,
		"NextOffset":   nextOffset,
		"PrevPage":     currentPage - 1,
		"NextPage":     currentPage + 1,
	}
}

func (h *WebHandler) getClientIP(r *http.Request) string {
	// Get client IP from various headers
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Client-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

func (h *WebHandler) isValidSearchQuery(query string) bool {
	// Block hash validation (64 hex characters)
	blockHashRegex := `^[a-fA-F0-9]{64}$`
	
	// Transaction hash validation (64 hex characters)
	txHashRegex := `^[a-fA-F0-9]{64}$`
	
	// Address validation (26-35 characters, alphanumeric)
	addressRegex := `^[13][a-km-zA-HJ-NP-Z1-9]{25,34}$`
	
	// Check if query matches any of the patterns
	return strings.Contains(query, blockHashRegex) || 
		   strings.Contains(query, txHashRegex) || 
		   strings.Contains(query, addressRegex) ||
		   len(query) >= 10 // Allow longer text searches
}
