// GoChain Explorer JavaScript

// Main application class
class GoChainExplorer {
    constructor() {
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.setupSearchSuggestions();
        this.setupCopyToClipboard();
        this.setupResponsiveNavigation();
        this.setupAutoRefresh();
        this.setupAdvancedFeatures();
        this.setupPerformanceMonitoring();
    }

    // Event listeners setup
    setupEventListeners() {
        // Search form submission
        const searchForm = document.querySelector('.search-form');
        if (searchForm) {
            searchForm.addEventListener('submit', this.handleSearch.bind(this));
        }

        // Search input suggestions
        const searchInput = document.querySelector('.search-input');
        if (searchInput) {
            searchInput.addEventListener('input', this.handleSearchInput.bind(this));
            searchInput.addEventListener('focus', this.showSearchSuggestions.bind(this));
            searchInput.addEventListener('blur', this.hideSearchSuggestions.bind(this));
        }

        // Mobile navigation toggle
        const mobileNavToggle = document.querySelector('.mobile-nav-toggle');
        if (mobileNavToggle) {
            mobileNavToggle.addEventListener('click', this.toggleMobileNavigation.bind(this));
        }

        // Copy buttons
        document.addEventListener('click', this.handleCopyClick.bind(this));

        // Advanced search filters
        this.setupAdvancedSearchFilters();
    }

    // Advanced search filters
    setupAdvancedSearchFilters() {
        const searchContainer = document.querySelector('.search-form-container');
        if (searchContainer) {
            const advancedFilters = document.createElement('div');
            advancedFilters.className = 'advanced-filters';
            advancedFilters.innerHTML = `
                <div class="filters-toggle">
                    <button type="button" class="btn btn-outline-secondary" id="toggleFilters">
                        Advanced Filters
                    </button>
                </div>
                <div class="filters-panel" id="filtersPanel" style="display: none;">
                    <div class="filter-group">
                        <label for="filterType">Type:</label>
                        <select id="filterType" class="form-control">
                            <option value="">All Types</option>
                            <option value="block">Block</option>
                            <option value="transaction">Transaction</option>
                            <option value="address">Address</option>
                        </select>
                    </div>
                    <div class="filter-group">
                        <label for="filterDate">Date Range:</label>
                        <input type="date" id="filterDateFrom" class="form-control" placeholder="From">
                        <input type="date" id="filterDateTo" class="form-control" placeholder="To">
                    </div>
                    <div class="filter-group">
                        <label for="filterAmount">Amount Range:</label>
                        <input type="number" id="filterAmountMin" class="form-control" placeholder="Min">
                        <input type="number" id="filterAmountMax" class="form-control" placeholder="Max">
                    </div>
                    <div class="filter-group">
                        <label for="filterBlockHeight">Block Height:</label>
                        <input type="number" id="filterBlockHeight" class="form-control" placeholder="Block #">
                    </div>
                </div>
            `;
            
            searchContainer.appendChild(advancedFilters);
            
            // Toggle filters panel
            const toggleBtn = document.getElementById('toggleFilters');
            const filtersPanel = document.getElementById('filtersPanel');
            
            if (toggleBtn && filtersPanel) {
                toggleBtn.addEventListener('click', () => {
                    const isVisible = filtersPanel.style.display !== 'none';
                    filtersPanel.style.display = isVisible ? 'none' : 'block';
                    toggleBtn.textContent = isVisible ? 'Advanced Filters' : 'Hide Filters';
                });
            }
        }
    }

    // Search functionality
    handleSearch(event) {
        const searchInput = event.target.querySelector('.search-input');
        const query = searchInput.value.trim();
        
        if (!query) {
            event.preventDefault();
            this.showError('Please enter a search term');
            return;
        }

        // Validate search input format
        if (!this.isValidSearchQuery(query)) {
            event.preventDefault();
            this.showError('Please enter a valid block hash, transaction hash, or address');
            return;
        }

        // Apply advanced filters
        this.applyAdvancedFilters(event);
    }

    applyAdvancedFilters(event) {
        const filterType = document.getElementById('filterType')?.value;
        const filterDateFrom = document.getElementById('filterDateFrom')?.value;
        const filterDateTo = document.getElementById('filterDateTo')?.value;
        const filterAmountMin = document.getElementById('filterAmountMin')?.value;
        const filterAmountMax = document.getElementById('filterAmountMax')?.value;
        const filterBlockHeight = document.getElementById('filterBlockHeight')?.value;

        // Build query string with filters
        let queryString = event.target.action;
        const params = new URLSearchParams();
        
        const searchInput = event.target.querySelector('.search-input');
        params.append('q', searchInput.value.trim());
        
        if (filterType) params.append('type', filterType);
        if (filterDateFrom) params.append('date_from', filterDateFrom);
        if (filterDateTo) params.append('date_to', filterDateTo);
        if (filterAmountMin) params.append('amount_min', filterAmountMin);
        if (filterAmountMax) params.append('amount_max', filterAmountMax);
        if (filterBlockHeight) params.append('block_height', filterBlockHeight);

        // Redirect to search with filters
        event.preventDefault();
        window.location.href = `${queryString}?${params.toString()}`;
    }

    handleSearchInput(event) {
        const query = event.target.value.trim();
        
        if (query.length >= 3) {
            this.fetchSearchSuggestions(query);
        } else {
            this.hideSearchSuggestions();
        }
    }

    async fetchSearchSuggestions(query) {
        try {
            const response = await fetch(`/api/v1/search?q=${encodeURIComponent(query)}`);
            const data = await response.json();
            
            if (data.suggestions && data.suggestions.length > 0) {
                this.showSearchSuggestions(data.suggestions);
            }
        } catch (error) {
            console.error('Failed to fetch search suggestions:', error);
        }
    }

    showSearchSuggestions(suggestions = []) {
        let suggestionsContainer = document.querySelector('.search-suggestions');
        
        if (!suggestionsContainer) {
            suggestionsContainer = document.createElement('div');
            suggestionsContainer.className = 'search-suggestions';
            suggestionsContainer.style.cssText = `
                position: absolute;
                top: 100%;
                left: 0;
                right: 0;
                background: white;
                border: 1px solid #e9ecef;
                border-radius: 4px;
                box-shadow: 0 4px 8px rgba(0,0,0,0.1);
                z-index: 1000;
                max-height: 200px;
                overflow-y: auto;
            `;
            
            const searchContainer = document.querySelector('.search-form');
            if (searchContainer) {
                searchContainer.style.position = 'relative';
                searchContainer.appendChild(suggestionsContainer);
            }
        }

        if (suggestions.length === 0) {
            suggestionsContainer.innerHTML = '<div class="suggestion-item">No suggestions found</div>';
        } else {
            suggestionsContainer.innerHTML = suggestions.map(suggestion => 
                `<div class="suggestion-item" data-suggestion="${suggestion}">${suggestion}</div>`
            ).join('');
            
            // Add click handlers for suggestions
            suggestionsContainer.querySelectorAll('.suggestion-item').forEach(item => {
                item.addEventListener('click', () => {
                    const searchInput = document.querySelector('.search-input');
                    if (searchInput) {
                        searchInput.value = item.dataset.suggestion;
                        document.querySelector('.search-form').submit();
                    }
                });
            });
        }
        
        suggestionsContainer.style.display = 'block';
    }

    hideSearchSuggestions() {
        const suggestionsContainer = document.querySelector('.search-suggestions');
        if (suggestionsContainer) {
            suggestionsContainer.style.display = 'none';
        }
    }

    // Search validation
    isValidSearchQuery(query) {
        // Block hash validation (64 hex characters)
        const blockHashRegex = /^[a-fA-F0-9]{64}$/;
        
        // Transaction hash validation (64 hex characters)
        const txHashRegex = /^[a-fA-F0-9]{64}$/;
        
        // Address validation (26-35 characters, alphanumeric)
        const addressRegex = /^[13][a-km-zA-HJ-NP-Z1-9]{25,34}$/;
        
        return blockHashRegex.test(query) || txHashRegex.test(query) || addressRegex.test(query);
    }

    // Copy to clipboard functionality
    setupCopyToClipboard() {
        // Add copy buttons to hash and address elements
        const copyableElements = document.querySelectorAll('.block-hash, .transaction-hash, .address-hash, .tx-hash');
        
        copyableElements.forEach(element => {
            const copyButton = document.createElement('button');
            copyButton.className = 'copy-button';
            copyButton.innerHTML = '📋';
            copyButton.title = 'Copy to clipboard';
            copyButton.style.cssText = `
                background: none;
                border: none;
                cursor: pointer;
                font-size: 1rem;
                margin-left: 0.5rem;
                opacity: 0.7;
                transition: opacity 0.2s ease;
            `;
            
            copyButton.addEventListener('mouseenter', () => {
                copyButton.style.opacity = '1';
            });
            
            copyButton.addEventListener('mouseleave', () => {
                copyButton.style.opacity = '0.7';
            });
            
            element.appendChild(copyButton);
        });
    }

    handleCopyClick(event) {
        if (event.target.classList.contains('copy-button')) {
            const container = event.target.parentElement;
            const textToCopy = container.textContent.replace('📋', '').trim();
            
            this.copyToClipboard(textToCopy);
            this.showCopySuccess(event.target);
        }
    }

    async copyToClipboard(text) {
        try {
            if (navigator.clipboard && window.isSecureContext) {
                await navigator.clipboard.writeText(text);
            } else {
                // Fallback for older browsers
                const textArea = document.createElement('textarea');
                textArea.value = text;
                textArea.style.position = 'fixed';
                textArea.style.left = '-999999px';
                textArea.style.top = '-999999px';
                document.body.appendChild(textArea);
                textArea.focus();
                textArea.select();
                document.execCommand('copy');
                textArea.remove();
            }
        } catch (error) {
            console.error('Failed to copy to clipboard:', error);
        }
    }

    showCopySuccess(button) {
        const originalText = button.innerHTML;
        button.innerHTML = '✅';
        button.style.color = '#28a745';
        
        setTimeout(() => {
            button.innerHTML = originalText;
            button.style.color = '';
        }, 2000);
    }

    // Mobile navigation
    setupResponsiveNavigation() {
        const nav = document.querySelector('.nav');
        if (nav) {
            // Add mobile navigation toggle button
            const mobileToggle = document.createElement('button');
            mobileToggle.className = 'mobile-nav-toggle';
            mobileToggle.innerHTML = '☰';
            mobileToggle.style.cssText = `
                display: none;
                background: none;
                border: none;
                color: white;
                font-size: 1.5rem;
                cursor: pointer;
            `;
            
            nav.insertBefore(mobileToggle, nav.firstChild);
            
            // Add mobile styles
            const style = document.createElement('style');
            style.textContent = `
                @media (max-width: 768px) {
                    .nav-menu {
                        display: none;
                        width: 100%;
                        flex-direction: column;
                        gap: 0;
                    }
                    
                    .nav-menu.active {
                        display: flex;
                    }
                    
                    .nav-menu .nav-link {
                        padding: 1rem;
                        border-bottom: 1px solid rgba(255,255,255,0.1);
                    }
                    
                    .mobile-nav-toggle {
                        display: block !important;
                    }
                }
            `;
            document.head.appendChild(style);
        }
    }

    toggleMobileNavigation() {
        const navMenu = document.querySelector('.nav-menu');
        if (navMenu) {
            navMenu.classList.toggle('active');
        }
    }

    // Auto-refresh functionality for dashboard
    setupAutoRefresh() {
        // Only auto-refresh on dashboard page
        if (window.location.pathname === '/') {
            this.startAutoRefresh();
        }
    }

    startAutoRefresh() {
        // Refresh dashboard data every 30 seconds
        setInterval(() => {
            this.refreshDashboard();
        }, 30000);
    }

    async refreshDashboard() {
        try {
            const response = await fetch('/api/v1/dashboard');
            const data = await response.json();
            
            if (data) {
                this.updateDashboard(data);
            }
        } catch (error) {
            console.error('Failed to refresh dashboard:', error);
        }
    }

    updateDashboard(data) {
        // Update statistics
        if (data.stats) {
            this.updateElement('.stat-card:nth-child(1) .stat-value', data.stats.total_blocks);
            this.updateElement('.stat-card:nth-child(2) .stat-value', data.stats.total_transactions);
            this.updateElement('.stat-card:nth-child(3) .stat-value', data.stats.total_addresses);
            this.updateElement('.stat-card:nth-child(4) .stat-value', this.formatDifficulty(data.stats.difficulty));
        }

        // Update recent blocks
        if (data.recent_blocks) {
            this.updateRecentBlocks(data.recent_blocks);
        }

        // Update recent transactions
        if (data.recent_transactions) {
            this.updateRecentTransactions(data.recent_transactions);
        }

        // Update last update time
        this.updateElement('.dashboard .stats-grid', `Last updated: ${new Date().toLocaleTimeString()}`);
    }

    updateElement(selector, content) {
        const element = document.querySelector(selector);
        if (element) {
            element.textContent = content;
        }
    }

    updateRecentBlocks(blocks) {
        const container = document.querySelector('.recent-blocks .block-list');
        if (container && blocks.length > 0) {
            container.innerHTML = blocks.map(block => `
                <div class="block-item">
                    <div class="block-hash">
                        <a href="/blocks/${this.formatHash(block.hash)}">${this.formatHash(block.hash)}</a>
                    </div>
                    <div class="block-info">
                        <span class="block-height">#${block.height}</span>
                        <span class="block-time">${this.formatTime(block.timestamp)}</span>
                        <span class="block-tx-count">${block.tx_count} tx</span>
                    </div>
                </div>
            `).join('');
        }
    }

    updateRecentTransactions(transactions) {
        const container = document.querySelector('.recent-transactions .transaction-list');
        if (container && transactions.length > 0) {
            container.innerHTML = transactions.map(tx => `
                <div class="transaction-item">
                    <div class="tx-hash">
                        <a href="/transactions/${this.formatHash(tx.hash)}">${this.formatHash(tx.hash)}</a>
                    </div>
                    <div class="tx-info">
                        <span class="tx-amount">${this.formatAmount(tx.amount)}</span>
                        <span class="tx-time">${this.formatTime(tx.timestamp)}</span>
                    </div>
                </div>
            `).join('');
        }
    }

    // Advanced features setup
    setupAdvancedFeatures() {
        // Initialize charts if Chart.js is available
        if (typeof Chart !== 'undefined') {
            this.initializeCharts();
        } else {
            // Load Chart.js dynamically
            this.loadChartJS();
        }

        // Setup real-time updates
        this.setupRealTimeUpdates();
        
        // Setup export functionality
        this.setupExportFeatures();
    }

    async loadChartJS() {
        try {
            // Load Chart.js from CDN
            const script = document.createElement('script');
            script.src = 'https://cdn.jsdelivr.net/npm/chart.js';
            script.onload = () => {
                this.initializeCharts();
            };
            document.head.appendChild(script);
        } catch (error) {
            console.error('Failed to load Chart.js:', error);
        }
    }

    initializeCharts() {
        // Initialize charts module
        if (typeof BlockchainCharts !== 'undefined') {
            this.charts = new BlockchainCharts();
        }
    }

    setupRealTimeUpdates() {
        // Setup WebSocket connection for real-time updates
        if ('WebSocket' in window) {
            this.connectWebSocket();
        }
    }

    connectWebSocket() {
        try {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = `${protocol}//${window.location.host}/ws/explorer`;
            
            this.websocket = new WebSocket(wsUrl);
            
            this.websocket.onopen = () => {
                console.log('WebSocket connected for real-time updates');
            };
            
            this.websocket.onmessage = (event) => {
                const data = JSON.parse(event.data);
                this.handleRealTimeUpdate(data);
            };
            
            this.websocket.onclose = () => {
                console.log('WebSocket disconnected, attempting to reconnect...');
                setTimeout(() => this.connectWebSocket(), 5000);
            };
            
            this.websocket.onerror = (error) => {
                console.error('WebSocket error:', error);
            };
        } catch (error) {
            console.error('Failed to connect WebSocket:', error);
        }
    }

    handleRealTimeUpdate(data) {
        switch (data.type) {
            case 'new_block':
                this.handleNewBlock(data.block);
                break;
            case 'new_transaction':
                this.handleNewTransaction(data.transaction);
                break;
            case 'network_update':
                this.handleNetworkUpdate(data.network);
                break;
        }
    }

    handleNewBlock(block) {
        // Update dashboard with new block
        this.showNotification(`New block found: #${block.height}`, 'success');
        
        // Update charts if available
        if (this.charts) {
            this.charts.updateCharts({
                blockTime: [block.timestamp],
                txVolume: [block.transaction_count]
            });
        }
    }

    handleNewTransaction(transaction) {
        // Show notification for new transaction
        this.showNotification(`New transaction: ${this.formatHash(transaction.hash)}`, 'info');
    }

    handleNetworkUpdate(network) {
        // Update network statistics
        this.updateElement('.network-status', network.status);
        this.updateElement('.peer-count', network.peer_count);
    }

    setupExportFeatures() {
        // Add export buttons to data pages
        this.addExportButtons();
    }

    addExportButtons() {
        const exportButtons = `
            <div class="export-controls">
                <button class="btn btn-outline-primary" onclick="window.explorerApp.exportData('csv')">
                    Export CSV
                </button>
                <button class="btn btn-outline-primary" onclick="window.explorerApp.exportData('json')">
                    Export JSON
                </button>
            </div>
        `;
        
        // Add to appropriate pages
        const blockList = document.querySelector('.blocks-page .page-header');
        const txList = document.querySelector('.transactions-page .page-header');
        
        if (blockList) {
            blockList.insertAdjacentHTML('beforeend', exportButtons);
        }
        
        if (txList) {
            txList.insertAdjacentHTML('beforeend', exportButtons);
        }
    }

    exportData(format) {
        const pageType = this.getCurrentPageType();
        const data = this.getPageData();
        
        if (format === 'csv') {
            this.exportToCSV(data, pageType);
        } else if (format === 'json') {
            this.exportToJSON(data, pageType);
        }
    }

    getCurrentPageType() {
        if (window.location.pathname.includes('/blocks')) return 'blocks';
        if (window.location.pathname.includes('/transactions')) return 'transactions';
        if (window.location.pathname.includes('/addresses')) return 'addresses';
        return 'data';
    }

    getPageData() {
        // Extract data from current page
        const data = [];
        
        if (window.location.pathname.includes('/blocks')) {
            document.querySelectorAll('.block-card').forEach(card => {
                data.push({
                    height: card.querySelector('.block-height')?.textContent,
                    hash: card.querySelector('.block-hash a')?.textContent,
                    time: card.querySelector('.block-time')?.textContent,
                    transactions: card.querySelector('.block-tx-count')?.textContent
                });
            });
        } else if (window.location.pathname.includes('/transactions')) {
            document.querySelectorAll('.transaction-card').forEach(card => {
                data.push({
                    hash: card.querySelector('.transaction-hash a')?.textContent,
                    amount: card.querySelector('.tx-amount')?.textContent,
                    time: card.querySelector('.tx-time')?.textContent,
                    status: card.querySelector('.transaction-status')?.textContent
                });
            });
        }
        
        return data;
    }

    exportToCSV(data, type) {
        if (data.length === 0) {
            this.showError('No data to export');
            return;
        }
        
        const headers = Object.keys(data[0]);
        const csvContent = [
            headers.join(','),
            ...data.map(row => headers.map(header => `"${row[header] || ''}"`).join(','))
        ].join('\n');
        
        this.downloadFile(csvContent, `${type}_export.csv`, 'text/csv');
    }

    exportToJSON(data, type) {
        if (data.length === 0) {
            this.showError('No data to export');
            return;
        }
        
        const jsonContent = JSON.stringify(data, null, 2);
        this.downloadFile(jsonContent, `${type}_export.json`, 'application/json');
    }

    downloadFile(content, filename, mimeType) {
        const blob = new Blob([content], { type: mimeType });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    }

    showNotification(message, type = 'info') {
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.textContent = message;
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            padding: 1rem;
            border-radius: 4px;
            color: white;
            z-index: 10000;
            max-width: 300px;
            background: ${type === 'success' ? '#28a745' : type === 'error' ? '#dc3545' : '#17a2b8'};
        `;
        
        document.body.appendChild(notification);
        
        setTimeout(() => {
            if (notification.parentNode) {
                notification.parentNode.removeChild(notification);
            }
        }, 5000);
    }

    // Utility functions
    formatHash(hash) {
        if (!hash || hash.length <= 8) {
            return hash || 'N/A';
        }
        return `${hash.substring(0, 4)}...${hash.substring(hash.length - 4)}`;
    }

    formatAmount(amount) {
        if (amount === 0) return '0';
        return (amount / 100000000).toFixed(8);
    }

    formatTime(timestamp) {
        if (!timestamp) return 'N/A';
        return new Date(timestamp).toLocaleString();
    }

    formatDifficulty(difficulty) {
        if (difficulty === 0) return '0';
        if (difficulty >= 1000000000) {
            return `${(difficulty / 1000000000).toFixed(2)} G`;
        } else if (difficulty >= 1000000) {
            return `${(difficulty / 1000000).toFixed(2)} M`;
        } else if (difficulty >= 1000) {
            return `${(difficulty / 1000).toFixed(2)} K`;
        }
        return difficulty.toString();
    }

    // Error handling
    showError(message) {
        // Create error notification
        const notification = document.createElement('div');
        notification.className = 'error-notification';
        notification.textContent = message;
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            background: #dc3545;
            color: white;
            padding: 1rem;
            border-radius: 4px;
            box-shadow: 0 4px 8px rgba(0,0,0,0.2);
            z-index: 10000;
            max-width: 300px;
        `;
        
        document.body.appendChild(notification);
        
        // Auto-remove after 5 seconds
        setTimeout(() => {
            if (notification.parentNode) {
                notification.parentNode.removeChild(notification);
            }
        }, 5000);
    }

    // Performance monitoring
    setupPerformanceMonitoring() {
        // Monitor page load performance
        if ('performance' in window) {
            window.addEventListener('load', () => {
                const perfData = performance.getEntriesByType('navigation')[0];
                if (perfData) {
                    console.log(`Page load time: ${perfData.loadEventEnd - perfData.loadEventStart}ms`);
                }
            });
        }

        // Monitor API response times
        this.interceptFetchRequests();
    }

    interceptFetchRequests() {
        const originalFetch = window.fetch;
        window.fetch = async (...args) => {
            const startTime = performance.now();
            try {
                const response = await originalFetch(...args);
                const endTime = performance.now();
                console.log(`API request to ${args[0]} took ${endTime - startTime}ms`);
                return response;
            } catch (error) {
                const endTime = performance.now();
                console.error(`API request to ${args[0]} failed after ${endTime - startTime}ms:`, error);
                throw error;
            }
        };
    }
}

// Initialize the application when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.explorerApp = new GoChainExplorer();
});

// Export for potential use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = GoChainExplorer;
}
