package web

// Template content constants

const baseTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <link rel="stylesheet" href="/static/css/style.css">
    <link rel="icon" type="image/x-icon" href="/static/favicon.ico">
</head>
<body>
    <header class="header">
        <nav class="nav">
            <div class="nav-brand">
                <a href="/" class="brand-link">adrenochain Explorer</a>
            </div>
            <div class="nav-menu">
                <a href="/" class="nav-link">Home</a>
                <a href="/blocks" class="nav-link">Blocks</a>
                <a href="/transactions" class="nav-link">Transactions</a>
                <a href="/search" class="nav-link">Search</a>
            </div>
            <div class="nav-search">
                <form action="/search" method="GET" class="search-form">
                    <input type="text" name="q" placeholder="Search blocks, transactions, addresses..." class="search-input">
                    <button type="submit" class="search-button">Search</button>
                </form>
            </div>
        </nav>
    </header>

    <main class="main">
        {{template "content" .}}
    </main>

    <footer class="footer">
        <div class="footer-content">
            <p>&copy; 2024 adrenochain Explorer. Built with Go.</p>
        </div>
    </footer>

    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <script src="/static/js/charts.js"></script>
    <script src="/static/js/performance.js"></script>
    <script src="/static/js/mempool.js"></script>
    <script src="/static/js/app.js"></script>
</body>
</html>`

const homeTemplate = `{{define "content"}}
<div class="dashboard">
    <div class="dashboard-header">
        <h1>adrenochain Blockchain Explorer</h1>
        <p>Explore the adrenochain blockchain in real-time</p>
    </div>

    {{if .Dashboard}}
    <div class="stats-grid">
        <div class="stat-card">
            <h3>Total Blocks</h3>
            <div class="stat-value">{{.Dashboard.Stats.TotalBlocks}}</div>
        </div>
        <div class="stat-card">
            <h3>Total Transactions</h3>
            <div class="stat-value">{{.Dashboard.Stats.TotalTransactions}}</div>
        </div>
        <div class="stat-card">
            <h3>Total Addresses</h3>
            <div class="stat-value">{{.Dashboard.Stats.TotalAddresses}}</div>
        </div>
        <div class="stat-card">
            <h3>Difficulty</h3>
            <div class="stat-value">{{formatDifficulty .Dashboard.Stats.Difficulty}}</div>
        </div>
    </div>

    <div class="recent-activity">
        <div class="recent-blocks">
            <h2>Recent Blocks</h2>
            {{if .Dashboard.RecentBlocks}}
            <div class="block-list">
                {{range .Dashboard.RecentBlocks}}
                <div class="block-item">
                    <div class="block-hash">
                        <a href="/blocks/{{formatHash .Hash}}">{{formatHash .Hash}}</a>
                    </div>
                    <div class="block-info">
                        <span class="block-height">#{{.Height}}</span>
                        <span class="block-time">{{formatTime .Timestamp}}</span>
                        <span class="block-tx-count">{{.TxCount}} tx</span>
                    </div>
                </div>
                {{end}}
            </div>
            {{else}}
            <p>No recent blocks available</p>
            {{end}}
        </div>

        <div class="recent-transactions">
            <h2>Recent Transactions</h2>
            {{if .Dashboard.RecentTxs}}
            <div class="transaction-list">
                {{range .Dashboard.RecentTxs}}
                <div class="transaction-item">
                    <div class="tx-hash">
                        <a href="/transactions/{{formatHash .Hash}}">{{formatHash .Hash}}</a>
                    </div>
                    <div class="tx-info">
                        <span class="tx-amount">{{formatAmount .Amount}}</span>
                        <span class="tx-time">{{formatTime .Timestamp}}</span>
                    </div>
                </div>
                {{end}}
            </div>
            {{else}}
            <p>No recent transactions available</p>
            {{end}}
        </div>
    </div>
    {{else}}
    <div class="error-message">
        <p>Failed to load dashboard data</p>
    </div>
    {{end}}
</div>
{{end}}`

const blocksTemplate = `{{define "content"}}
<div class="blocks-page">
    <div class="page-header">
        <h1>Blocks</h1>
        <p>Browse all blocks in the adrenochain blockchain</p>
    </div>

    {{if .Blocks}}
    <div class="blocks-list">
        {{range .Blocks}}
        <div class="block-card">
            <div class="block-header">
                <h3 class="block-height">Block #{{.Height}}</h3>
                <span class="block-hash">
                    <a href="/blocks/{{formatHash .Hash}}">{{formatHash .Hash}}</a>
                </span>
            </div>
            <div class="block-details">
                <div class="block-info">
                    <span class="info-label">Time:</span>
                    <span class="info-value">{{formatTime .Timestamp}}</span>
                </div>
                <div class="block-info">
                    <span class="info-label">Transactions:</span>
                    <span class="info-value">{{.TxCount}}</span>
                </div>
                <div class="block-info">
                    <span class="info-label">Size:</span>
                    <span class="info-value">{{.Size}} bytes</span>
                </div>
                <div class="block-info">
                    <span class="info-label">Difficulty:</span>
                    <span class="info-value">{{formatDifficulty .Difficulty}}</span>
                </div>
                <div class="block-info">
                    <span class="info-label">Confirmations:</span>
                    <span class="info-value">{{.Confirmations}}</span>
                </div>
            </div>
        </div>
        {{end}}
    </div>

    {{if .Pagination}}
    <div class="pagination">
        {{if .Pagination.HasPrev}}
        <a href="?limit={{.Pagination.Limit}}&offset={{.Pagination.PrevOffset}}" class="pagination-link">Previous</a>
        {{end}}
        
        <span class="pagination-info">
            Page {{.Pagination.CurrentPage}} of {{.Pagination.TotalPages}}
        </span>
        
        {{if .Pagination.HasNext}}
        <a href="?limit={{.Pagination.Limit}}&offset={{.Pagination.NextOffset}}" class="pagination-link">Next</a>
        {{end}}
    </div>
    {{end}}
    {{else}}
    <div class="empty-state">
        <p>No blocks found</p>
    </div>
    {{end}}
</div>
{{end}}`

const blockDetailTemplate = `{{define "content"}}
<div class="block-detail-page">
    {{if .Block}}
    <div class="page-header">
        <h1>Block #{{.Block.Height}}</h1>
        <p class="block-hash">{{formatHash .Block.Hash}}</p>
    </div>

    <div class="block-details-grid">
        <div class="detail-section">
            <h3>Block Information</h3>
            <div class="detail-item">
                <span class="detail-label">Hash:</span>
                <span class="detail-value">{{formatHash .Block.Hash}}</span>
            </div>
            <div class="detail-item">
                <span class="detail-label">Height:</span>
                <span class="detail-value">{{.Block.Height}}</span>
            </div>
            <div class="detail-item">
                <span class="detail-label">Timestamp:</span>
                <span class="detail-value">{{formatTime .Block.Timestamp}}</span>
            </div>
            <div class="detail-item">
                <span class="detail-label">Difficulty:</span>
                <span class="detail-value">{{formatDifficulty .Block.Difficulty}}</span>
            </div>
            <div class="detail-item">
                <span class="detail-label">Nonce:</span>
                <span class="detail-value">{{.Block.Nonce}}</span>
            </div>
            <div class="detail-item">
                <span class="detail-label">Version:</span>
                <span class="detail-value">{{.Block.Version}}</span>
            </div>
        </div>

        <div class="detail-section">
            <h3>Block Links</h3>
            {{if .Block.PrevHash}}
            <div class="detail-item">
                <span class="detail-label">Previous Block:</span>
                <span class="detail-value">
                    <a href="/blocks/{{formatHash .Block.PrevHash}}">{{formatHash .Block.PrevHash}}</a>
                </span>
            </div>
            {{end}}
            {{if .Block.NextHash}}
            <div class="detail-item">
                <span class="detail-label">Next Block:</span>
                <span class="detail-value">
                    <a href="/blocks/{{formatHash .Block.NextHash}}">{{formatHash .Block.NextHash}}</a>
                </span>
            </div>
            {{end}}
        </div>

        {{if .Block.Validation}}
        <div class="detail-section">
            <h3>Validation</h3>
            <div class="detail-item">
                <span class="detail-label">Status:</span>
                <span class="detail-value {{if .Block.Validation.IsValid}}valid{{else}}invalid{{end}}">
                    {{if .Block.Validation.IsValid}}Valid{{else}}Invalid{{end}}
                </span>
            </div>
            <div class="detail-item">
                <span class="detail-label">Confirmations:</span>
                <span class="detail-value">{{.Block.Validation.Confirmations}}</span>
            </div>
            <div class="detail-item">
                <span class="detail-label">Finality:</span>
                <span class="detail-value">{{.Block.Validation.Finality}}</span>
            </div>
            {{if .Block.Validation.Error}}
            <div class="detail-item">
                <span class="detail-label">Error:</span>
                <span class="detail-value error">{{.Block.Validation.Error}}</span>
            </div>
            {{end}}
        </div>
        {{end}}
    </div>

    <div class="transactions-section">
        <h3>Transactions ({{len .Block.Transactions}})</h3>
        {{if .Block.Transactions}}
        <div class="transaction-list">
            {{range .Block.Transactions}}
            <div class="transaction-item">
                <div class="tx-hash">
                    <a href="/transactions/{{formatHash .Hash}}">{{formatHash .Hash}}</a>
                </div>
                <div class="tx-info">
                    <span class="tx-amount">{{formatAmount .Amount}}</span>
                    <span class="tx-fee">Fee: {{formatAmount .Fee}}</span>
                    <span class="tx-inputs">{{.Inputs}} inputs</span>
                    <span class="tx-outputs">{{.Outputs}} outputs</span>
                </div>
            </div>
            {{end}}
        </div>
        {{else}}
        <p>No transactions in this block</p>
        {{end}}
    </div>
    {{else}}
    <div class="error-message">
        <p>Block not found</p>
    </div>
    {{end}}
</div>
{{end}}`

const transactionsTemplate = `{{define "content"}}
<div class="transactions-page">
    <div class="page-header">
        <h1>Transactions</h1>
        <p>Browse all transactions in the adrenochain blockchain</p>
    </div>

    {{if .Transactions}}
    <div class="transactions-list">
        {{range .Transactions}}
        <div class="transaction-card">
            <div class="transaction-header">
                <h3 class="transaction-hash">
                    <a href="/transactions/{{formatHash .Hash}}">{{formatHash .Hash}}</a>
                </h3>
                <span class="transaction-status {{.Status}}">{{.Status}}</span>
            </div>
            <div class="transaction-details">
                <div class="transaction-info">
                    <span class="info-label">Block:</span>
                    <span class="info-value">
                        <a href="/blocks/{{formatHash .BlockHash}}">#{{.Height}}</a>
                    </span>
                </div>
                <div class="transaction-info">
                    <span class="info-label">Time:</span>
                    <span class="info-value">{{formatTime .Timestamp}}</span>
                </div>
                <div class="transaction-info">
                    <span class="info-label">Amount:</span>
                    <span class="info-value">{{formatAmount .Amount}}</span>
                </div>
                <div class="transaction-info">
                    <span class="info-label">Fee:</span>
                    <span class="info-value">{{formatAmount .Fee}}</span>
                </div>
                <div class="transaction-info">
                    <span class="info-label">Inputs:</span>
                    <span class="info-value">{{.Inputs}}</span>
                </div>
                <div class="transaction-info">
                    <span class="info-label">Outputs:</span>
                    <span class="info-value">{{.Outputs}}</span>
                </div>
            </div>
        </div>
        {{end}}
    </div>

    {{if .Pagination}}
    <div class="pagination">
        {{if .Pagination.HasPrev}}
        <a href="?limit={{.Pagination.Limit}}&offset={{.Pagination.PrevOffset}}" class="pagination-link">Previous</a>
        {{end}}
        
        <span class="pagination-info">
            Page {{.Pagination.CurrentPage}} of {{.Pagination.TotalPages}}
        </span>
        
        {{if .Pagination.HasNext}}
        <a href="?limit={{.Pagination.Limit}}&offset={{.Pagination.NextOffset}}" class="pagination-link">Next</a>
        {{end}}
    </div>
    {{end}}
    {{else}}
    <div class="empty-state">
        <p>No transactions found</p>
    </div>
    {{end}}
</div>
{{end}}`

const transactionDetailTemplate = `{{define "content"}}
<div class="transaction-detail-page">
    {{if .Transaction}}
    <div class="page-header">
        <h1>Transaction</h1>
        <p class="transaction-hash">{{formatHash .Transaction.Hash}}</p>
    </div>

    <div class="transaction-details-grid">
        <div class="detail-section">
            <h3>Transaction Information</h3>
            <div class="detail-item">
                <span class="detail-label">Hash:</span>
                <span class="detail-value">{{formatHash .Transaction.Hash}}</span>
            </div>
            <div class="detail-item">
                <span class="detail-label">Status:</span>
                <span class="detail-value {{.Transaction.Status}}">{{.Transaction.Status}}</span>
            </div>
            <div class="detail-item">
                <span class="detail-label">Amount:</span>
                <span class="detail-value">{{formatAmount .Transaction.Amount}}</span>
            </div>
            <div class="detail-item">
                <span class="detail-label">Fee:</span>
                <span class="detail-value">{{formatAmount .Transaction.Fee}}</span>
            </div>
            <div class="detail-item">
                <span class="detail-label">Inputs:</span>
                <span class="detail-value">{{.Transaction.Inputs}}</span>
            </div>
            <div class="detail-item">
                <span class="detail-label">Outputs:</span>
                <span class="detail-value">{{.Transaction.Outputs}}</span>
            </div>
        </div>

        {{if .Transaction.BlockInfo}}
        <div class="detail-section">
            <h3>Block Information</h3>
            <div class="detail-item">
                <span class="detail-label">Block:</span>
                <span class="detail-value">
                    <a href="/blocks/{{formatHash .Transaction.BlockInfo.Hash}}">#{{.Transaction.BlockInfo.Height}}</a>
                </span>
            </div>
            <div class="detail-item">
                <span class="detail-label">Block Hash:</span>
                <span class="detail-value">{{formatHash .Transaction.BlockInfo.Hash}}</span>
            </div>
            <div class="detail-item">
                <span class="detail-label">Time:</span>
                <span class="detail-value">{{formatTime .Transaction.BlockInfo.Timestamp}}</span>
            </div>
            <div class="detail-item">
                <span class="detail-label">Confirmations:</span>
                <span class="detail-value">{{.Transaction.BlockInfo.Confirmations}}</span>
            </div>
        </div>
        {{end}}
    </div>

    {{if .Transaction.InputDetails}}
    <div class="inputs-section">
        <h3>Inputs ({{len .Transaction.InputDetails}})</h3>
        <div class="input-list">
            {{range .Transaction.InputDetails}}
            <div class="input-item">
                <div class="input-info">
                    <span class="input-tx">Tx: {{formatHash .TxHash}}</span>
                    <span class="input-index">Index: {{.TxIndex}}</span>
                </div>
                <div class="input-details">
                    <span class="input-address">{{formatAddress .Address}}</span>
                    <span class="input-amount">{{formatAmount .Amount}}</span>
                </div>
            </div>
            {{end}}
        </div>
    </div>
    {{end}}

    {{if .Transaction.OutputDetails}}
    <div class="outputs-section">
        <h3>Outputs ({{len .Transaction.OutputDetails}})</h3>
        <div class="output-list">
            {{range .Transaction.OutputDetails}}
            <div class="output-item">
                <div class="output-info">
                    <span class="output-index">Index: {{.Index}}</span>
                    <span class="output-spent {{if .Spent}}spent{{else}}unspent{{end}}">
                        {{if .Spent}}Spent{{else}}Unspent{{end}}
                    </span>
                </div>
                <div class="output-details">
                    <span class="output-address">{{formatAddress .Address}}</span>
                    <span class="output-amount">{{formatAmount .Amount}}</span>
                </div>
                {{if .SpentBy}}
                <div class="spent-by">
                    <span class="spent-by-label">Spent by:</span>
                    <span class="spent-by-tx">{{formatHash .SpentBy}}</span>
                </div>
                {{end}}
            </div>
            {{end}}
        </div>
    </div>
    {{end}}
    {{else}}
    <div class="error-message">
        <p>Transaction not found</p>
    </div>
    {{end}}
</div>
{{end}}`

const addressDetailTemplate = `{{define "content"}}
<div class="address-detail-page">
    {{if .Address}}
    <div class="page-header">
        <h1>Address</h1>
        <p class="address-hash">{{.Address.Address}}</p>
    </div>

    <div class="address-details-grid">
        <div class="detail-section">
            <h3>Address Information</h3>
            <div class="detail-item">
                <span class="detail-label">Address:</span>
                <span class="detail-value">{{.Address.Address}}</span>
            </div>
            <div class="detail-item">
                <span class="detail-label">Balance:</span>
                <span class="detail-value balance">{{formatAmount .Address.Balance}}</span>
            </div>
            <div class="detail-item">
                <span class="detail-label">Transaction Count:</span>
                <span class="detail-value">{{.Address.TxCount}}</span>
            </div>
            <div class="detail-item">
                <span class="detail-label">First Seen:</span>
                <span class="detail-value">{{formatTime .Address.FirstSeen}}</span>
            </div>
            <div class="detail-item">
                <span class="detail-label">Last Seen:</span>
                <span class="detail-value">{{formatTime .Address.LastSeen}}</span>
            </div>
        </div>

        <div class="detail-section">
            <h3>Summary</h3>
            <div class="detail-item">
                <span class="detail-label">Total Received:</span>
                <span class="detail-value received">{{formatAmount .Address.TotalReceived}}</span>
            </div>
            <div class="detail-item">
                <span class="detail-label">Total Sent:</span>
                <span class="detail-value sent">{{formatAmount .Address.TotalSent}}</span>
            </div>
        </div>
    </div>

    {{if .Address.UTXOs}}
    <div class="utxos-section">
        <h3>Unspent Transaction Outputs ({{len .Address.UTXOs}})</h3>
        <div class="utxo-list">
            {{range .Address.UTXOs}}
            <div class="utxo-item">
                <div class="utxo-info">
                    <span class="utxo-tx">Tx: {{formatHash .TxHash}}</span>
                    <span class="utxo-index">Index: {{.TxIndex}}</span>
                </div>
                <div class="utxo-details">
                    <span class="utxo-amount">{{formatAmount .Value}}</span>
                    <span class="utxo-height">Block: {{.Height}}</span>
                </div>
            </div>
            {{end}}
        </div>
    </div>
    {{end}}

    {{if .Address.Transactions}}
    <div class="transactions-section">
        <h3>Recent Transactions ({{len .Address.Transactions}})</h3>
        <div class="transaction-list">
            {{range .Address.Transactions}}
            <div class="transaction-item">
                <div class="tx-hash">
                    <a href="/transactions/{{formatHash .Hash}}">{{formatHash .Hash}}</a>
                </div>
                <div class="tx-info">
                    <span class="tx-amount">{{formatAmount .Amount}}</span>
                    <span class="tx-time">{{formatTime .Timestamp}}</span>
                    <span class="tx-block">
                        <a href="/blocks/{{formatHash .BlockHash}}">#{{.Height}}</a>
                    </span>
                </div>
            </div>
            {{end}}
        </div>
    </div>
    {{end}}
    {{else}}
    <div class="error-message">
        <p>Address not found</p>
    </div>
    {{end}}
</div>
{{end}}`

const searchTemplate = `{{define "content"}}
<div class="search-page">
    <div class="page-header">
        <h1>Search</h1>
        <p>Search for blocks, transactions, and addresses</p>
    </div>

    <div class="search-form-container">
        <form action="/search" method="GET" class="search-form">
            <div class="search-input-group">
                <input type="text" name="q" placeholder="Enter block hash, transaction hash, or address..." 
                       class="search-input" required>
                <button type="submit" class="search-button">Search</button>
            </div>
        </form>
    </div>

    <div class="search-help">
        <h3>What can you search for?</h3>
        <div class="search-examples">
            <div class="example-item">
                <h4>Block Hash</h4>
                <p>Search for a specific block by its hash</p>
                <code>0000000000000000000...</code>
            </div>
            <div class="example-item">
                <h4>Transaction Hash</h4>
                <p>Search for a specific transaction by its hash</p>
                <code>a1b2c3d4e5f6...</code>
            </div>
            <div class="example-item">
                <h4>Address</h4>
                <p>Search for an address to see its balance and transactions</p>
                <code>1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa</code>
            </div>
        </div>
    </div>
</div>
{{end}}`

const searchResultsTemplate = `{{define "content"}}
<div class="search-results-page">
    <div class="page-header">
        <h1>Search Results</h1>
        <p>Results for: "{{.Query}}"</p>
    </div>

    {{if .SearchResult}}
        {{if .SearchResult.Error}}
        <div class="search-error">
            <p>Search error: {{.SearchResult.Error}}</p>
        </div>
        {{else}}
            {{if .SearchResult.Block}}
            <div class="search-result-block">
                <h3>Block Found</h3>
                <div class="result-item">
                    <span class="result-label">Height:</span>
                    <span class="result-value">{{.SearchResult.Block.Height}}</span>
                </div>
                <div class="result-item">
                    <span class="result-label">Hash:</span>
                    <span class="result-value">
                        <a href="/blocks/{{formatHash .SearchResult.Block.Hash}}">{{formatHash .SearchResult.Block.Hash}}</a>
                    </span>
                </div>
                <div class="result-item">
                    <span class="result-label">Time:</span>
                    <span class="result-value">{{formatTime .SearchResult.Block.Timestamp}}</span>
                </div>
                <div class="result-item">
                    <span class="result-label">Transactions:</span>
                    <span class="result-value">{{.SearchResult.Block.TxCount}}</span>
                </div>
            </div>
            {{end}}

            {{if .SearchResult.Transaction}}
            <div class="search-result-transaction">
                <h3>Transaction Found</h3>
                <div class="result-item">
                    <span class="result-label">Hash:</span>
                    <span class="result-value">
                        <a href="/transactions/{{formatHash .SearchResult.Transaction.Hash}}">{{formatHash .SearchResult.Transaction.Hash}}</a>
                    </span>
                </div>
                <div class="result-item">
                    <span class="result-label">Block:</span>
                    <span class="result-value">
                        <a href="/blocks/{{formatHash .SearchResult.Transaction.BlockHash}}">#{{.SearchResult.Transaction.Height}}</a>
                    </span>
                </div>
                <div class="result-item">
                    <span class="result-label">Amount:</span>
                    <span class="result-value">{{formatAmount .SearchResult.Transaction.Amount}}</span>
                </div>
                <div class="result-item">
                    <span class="result-label">Time:</span>
                    <span class="result-value">{{formatTime .SearchResult.Transaction.Timestamp}}</span>
                </div>
            </div>
            {{end}}

            {{if .SearchResult.Address}}
            <div class="search-result-address">
                <h3>Address Found</h3>
                <div class="result-item">
                    <span class="result-label">Address:</span>
                    <span class="result-value">{{.SearchResult.Address.Address}}</span>
                </div>
                <div class="result-item">
                    <span class="result-label">Balance:</span>
                    <span class="result-value">{{formatAmount .SearchResult.Address.Balance}}</span>
                </div>
                <div class="result-item">
                    <span class="result-label">Transactions:</span>
                    <span class="result-value">{{.SearchResult.Address.TxCount}}</span>
                </div>
                <div class="result-item">
                    <span class="result-label">First Seen:</span>
                    <span class="result-value">{{formatTime .SearchResult.Address.FirstSeen}}</span>
                </div>
            </div>
            {{end}}

            {{if .SearchResult.Suggestions}}
            <div class="search-suggestions">
                <h3>Suggestions</h3>
                <ul class="suggestion-list">
                    {{range .SearchResult.Suggestions}}
                    <li><a href="/search?q={{.}}">{{.}}</a></li>
                    {{end}}
                </ul>
            </div>
            {{end}}
        {{end}}
    {{else}}
        <div class="no-results">
            <p>No results found for "{{.Query}}"</p>
            <p>Try searching for:</p>
            <ul>
                <li>A valid block hash</li>
                <li>A valid transaction hash</li>
                <li>A valid address</li>
            </ul>
        </div>
    {{end}}

    <div class="search-again">
        <a href="/search" class="search-again-link">Search Again</a>
    </div>
</div>
{{end}}`

const errorTemplate = `{{define "content"}}
<div class="error-page">
    <div class="error-content">
        <h1>Error</h1>
        <p class="error-message">{{.Message}}</p>
        {{if .Error}}
        <div class="error-details">
            <p class="error-detail">{{.Error}}</p>
        </div>
        {{end}}
        <div class="error-actions">
            <a href="/" class="btn btn-primary">Go Home</a>
            <a href="javascript:history.back()" class="btn btn-secondary">Go Back</a>
        </div>
    </div>
</div>
{{end}}`
