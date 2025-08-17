// adrenochain Explorer Mempool Visualization Module

class MempoolVisualizer {
    constructor() {
        this.mempoolData = [];
        this.canvas = null;
        this.ctx = null;
        this.init();
    }

    init() {
        this.setupMempoolContainer();
        this.loadMempoolData();
        this.startAutoRefresh();
    }

    setupMempoolContainer() {
        // Find or create mempool container
        let mempoolContainer = document.querySelector('.mempool-container');
        if (!mempoolContainer) {
            mempoolContainer = document.createElement('div');
            mempoolContainer.className = 'mempool-container';
            mempoolContainer.innerHTML = `
                <div class="mempool-header">
                    <h3>Mempool (Pending Transactions)</h3>
                    <div class="mempool-stats">
                        <span class="tx-count">0 transactions</span>
                        <span class="total-fees">0 fees</span>
                    </div>
                </div>
                <div class="mempool-visualization">
                    <canvas id="mempoolCanvas" width="800" height="400"></canvas>
                </div>
                <div class="mempool-transactions">
                    <h4>Recent Pending Transactions</h4>
                    <div class="pending-tx-list"></div>
                </div>
            `;
            
            // Insert into dashboard if available
            const dashboard = document.querySelector('.dashboard');
            if (dashboard) {
                dashboard.appendChild(mempoolContainer);
            }
        }

        // Setup canvas
        this.canvas = document.getElementById('mempoolCanvas');
        if (this.canvas) {
            this.ctx = this.canvas.getContext('2d');
            this.setupCanvas();
        }
    }

    setupCanvas() {
        if (!this.ctx) return;

        // Set canvas size
        this.canvas.width = this.canvas.offsetWidth;
        this.canvas.height = this.canvas.offsetHeight;

        // Initial draw
        this.drawMempool();
    }

    async loadMempoolData() {
        try {
            // In a real implementation, this would fetch from the API
            // For now, we'll generate sample data
            this.generateSampleMempoolData();
            this.updateMempoolDisplay();
            this.drawMempool();
        } catch (error) {
            console.error('Failed to load mempool data:', error);
        }
    }

    generateSampleMempoolData() {
        this.mempoolData = [];
        const txCount = Math.floor(Math.random() * 50) + 10; // 10-60 transactions
        
        for (let i = 0; i < txCount; i++) {
            this.mempoolData.push({
                hash: this.generateRandomHash(),
                fee: Math.floor(Math.random() * 10000) + 100, // 100-10100 satoshis
                size: Math.floor(Math.random() * 1000) + 200, // 200-1200 bytes
                timestamp: Date.now() - Math.random() * 3600000, // Last hour
                priority: Math.random() // 0-1 priority score
            });
        }

        // Sort by fee (highest first)
        this.mempoolData.sort((a, b) => b.fee - a.fee);
    }

    generateRandomHash() {
        const chars = '0123456789abcdef';
        let hash = '';
        for (let i = 0; i < 64; i++) {
            hash += chars[Math.floor(Math.random() * chars.length)];
        }
        return hash;
    }

    updateMempoolDisplay() {
        const txCount = document.querySelector('.tx-count');
        const totalFees = document.querySelector('.total-fees');
        const pendingTxList = document.querySelector('.pending-tx-list');

        if (txCount) {
            txCount.textContent = `${this.mempoolData.length} transactions`;
        }

        if (totalFees) {
            const total = this.mempoolData.reduce((sum, tx) => sum + tx.fee, 0);
            totalFees.textContent = `${(total / 100000000).toFixed(8)} fees`;
        }

        if (pendingTxList) {
            pendingTxList.innerHTML = this.mempoolData.slice(0, 10).map(tx => `
                <div class="pending-tx-item">
                    <div class="tx-hash">
                        <a href="/transactions/${tx.hash}">${this.formatHash(tx.hash)}</a>
                    </div>
                    <div class="tx-fee">${tx.fee} sat</div>
                    <div class="tx-size">${tx.size} bytes</div>
                    <div class="tx-priority">
                        <div class="priority-bar" style="width: ${tx.priority * 100}%"></div>
                    </div>
                </div>
            `).join('');
        }
    }

    drawMempool() {
        if (!this.ctx || !this.canvas) return;

        const width = this.canvas.width;
        const height = this.canvas.height;

        // Clear canvas
        this.ctx.clearRect(0, 0, width, height);

        if (this.mempoolData.length === 0) {
            this.drawEmptyMempool(width, height);
            return;
        }

        // Draw mempool visualization
        this.drawFeeDistribution(width, height);
        this.drawSizeDistribution(width, height);
        this.drawPriorityHeatmap(width, height);
    }

    drawEmptyMempool(width, height) {
        this.ctx.fillStyle = '#f8f9fa';
        this.ctx.fillRect(0, 0, width, height);
        
        this.ctx.fillStyle = '#6c757d';
        this.ctx.font = '16px Arial';
        this.ctx.textAlign = 'center';
        this.ctx.fillText('Mempool is empty', width / 2, height / 2);
    }

    drawFeeDistribution(width, height) {
        const maxFee = Math.max(...this.mempoolData.map(tx => tx.fee));
        const minFee = Math.min(...this.mempoolData.map(tx => tx.fee));
        
        this.ctx.fillStyle = '#007bff';
        this.ctx.globalAlpha = 0.7;
        
        this.mempoolData.forEach((tx, index) => {
            const x = (index / this.mempoolData.length) * width;
            const feeHeight = ((tx.fee - minFee) / (maxFee - minFee)) * (height * 0.4);
            const y = height - feeHeight;
            
            this.ctx.fillRect(x, y, width / this.mempoolData.length, feeHeight);
        });
        
        this.ctx.globalAlpha = 1.0;
    }

    drawSizeDistribution(width, height) {
        const maxSize = Math.max(...this.mempoolData.map(tx => tx.size));
        const minSize = Math.min(...this.mempoolData.map(tx => tx.size));
        
        this.ctx.fillStyle = '#28a745';
        this.ctx.globalAlpha = 0.5;
        
        this.mempoolData.forEach((tx, index) => {
            const x = (index / this.mempoolData.length) * width;
            const sizeHeight = ((tx.size - minSize) / (maxSize - minSize)) * (height * 0.3);
            const y = height * 0.5 - sizeHeight;
            
            this.ctx.fillRect(x, y, width / this.mempoolData.length, sizeHeight);
        });
        
        this.ctx.globalAlpha = 1.0;
    }

    drawPriorityHeatmap(width, height) {
        this.ctx.fillStyle = '#ffc107';
        this.ctx.globalAlpha = 0.6;
        
        this.mempoolData.forEach((tx, index) => {
            const x = (index / this.mempoolData.length) * width;
            const priorityHeight = tx.priority * (height * 0.2);
            const y = height * 0.2 - priorityHeight;
            
            this.ctx.fillRect(x, y, width / this.mempoolData.length, priorityHeight);
        });
        
        this.ctx.globalAlpha = 1.0;
    }

    startAutoRefresh() {
        // Refresh mempool data every 30 seconds
        setInterval(() => {
            this.loadMempoolData();
        }, 30000);
    }

    formatHash(hash) {
        if (hash.length <= 8) return hash;
        return `${hash.substring(0, 4)}...${hash.substring(hash.length - 4)}`;
    }

    // Public methods for external access
    getMempoolStats() {
        return {
            transactionCount: this.mempoolData.length,
            totalFees: this.mempoolData.reduce((sum, tx) => sum + tx.fee, 0),
            averageFee: this.mempoolData.reduce((sum, tx) => sum + tx.fee, 0) / this.mempoolData.length,
            averageSize: this.mempoolData.reduce((sum, tx) => sum + tx.size, 0) / this.mempoolData.length,
            oldestTransaction: Math.min(...this.mempoolData.map(tx => tx.timestamp)),
            newestTransaction: Math.max(...this.mempoolData.map(tx => tx.timestamp))
        };
    }

    exportMempoolData() {
        const data = {
            timestamp: Date.now(),
            transactions: this.mempoolData,
            stats: this.getMempoolStats()
        };
        
        const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `mempool-${Date.now()}.json`;
        a.click();
        URL.revokeObjectURL(url);
    }
}

// Initialize mempool visualizer
window.mempoolVisualizer = new MempoolVisualizer();

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = MempoolVisualizer;
}
