// adrenochain Explorer Charts Module

class BlockchainCharts {
    constructor() {
        this.charts = new Map();
        this.chartColors = {
            primary: '#007bff',
            secondary: '#6c757d',
            success: '#28a745',
            warning: '#ffc107',
            danger: '#dc3545',
            info: '#17a2b8'
        };
        this.init();
    }

    init() {
        this.setupChartContainers();
        this.loadChartData();
    }

    setupChartContainers() {
        // Create chart containers if they don't exist
        const dashboard = document.querySelector('.dashboard');
        if (dashboard) {
            this.createChartSection(dashboard);
        }
    }

    createChartSection(dashboard) {
        const chartsSection = document.createElement('div');
        chartsSection.className = 'charts-section';
        chartsSection.innerHTML = `
            <div class="charts-header">
                <h2>Blockchain Analytics</h2>
                <p>Visual insights into blockchain activity</p>
            </div>
            <div class="charts-grid">
                <div class="chart-container">
                    <h3>Block Time Distribution</h3>
                    <canvas id="blockTimeChart" width="400" height="200"></canvas>
                </div>
                <div class="chart-container">
                    <h3>Transaction Volume</h3>
                    <canvas id="txVolumeChart" width="400" height="200"></canvas>
                </div>
                <div class="chart-container">
                    <h3>Difficulty Trend</h3>
                    <canvas id="difficultyChart" width="400" height="200"></canvas>
                </div>
                <div class="chart-container">
                    <h3>Address Growth</h3>
                    <canvas id="addressChart" width="400" height="200"></canvas>
                </div>
            </div>
        `;
        
        // Insert after stats grid
        const statsGrid = dashboard.querySelector('.stats-grid');
        if (statsGrid) {
            statsGrid.parentNode.insertBefore(chartsSection, statsGrid.nextSibling);
        }
    }

    async loadChartData() {
        try {
            // Load historical data for charts
            const response = await fetch('/api/v1/statistics');
            const data = await response.json();
            
            if (data && data.blockchain) {
                this.renderCharts(data);
            }
        } catch (error) {
            console.error('Failed to load chart data:', error);
            this.renderSampleCharts();
        }
    }

    renderCharts(data) {
        // Block time chart
        this.createBlockTimeChart(data);
        
        // Transaction volume chart
        this.createTransactionVolumeChart(data);
        
        // Difficulty chart
        this.createDifficultyChart(data);
        
        // Address growth chart
        this.createAddressGrowthChart(data);
    }

    renderSampleCharts() {
        // Render sample charts with mock data for demonstration
        this.createBlockTimeChart({
            block_time_data: this.generateSampleBlockTimeData()
        });
        
        this.createTransactionVolumeChart({
            tx_volume_data: this.generateSampleTxVolumeData()
        });
        
        this.createDifficultyChart({
            difficulty_data: this.generateSampleDifficultyData()
        });
        
        this.createAddressGrowthChart({
            address_data: this.generateSampleAddressData()
        });
    }

    createBlockTimeChart(data) {
        const ctx = document.getElementById('blockTimeChart');
        if (!ctx) return;

        const chartData = data.block_time_data || this.generateSampleBlockTimeData();
        
        this.charts.set('blockTime', new Chart(ctx, {
            type: 'line',
            data: {
                labels: chartData.labels,
                datasets: [{
                    label: 'Block Time (seconds)',
                    data: chartData.values,
                    borderColor: this.chartColors.primary,
                    backgroundColor: this.chartColors.primary + '20',
                    borderWidth: 2,
                    fill: true,
                    tension: 0.4
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: false
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        title: {
                            display: true,
                            text: 'Seconds'
                        }
                    },
                    x: {
                        title: {
                            display: true,
                            text: 'Block Height'
                        }
                    }
                }
            }
        }));
    }

    createTransactionVolumeChart(data) {
        const ctx = document.getElementById('txVolumeChart');
        if (!ctx) return;

        const chartData = data.tx_volume_data || this.generateSampleTxVolumeData();
        
        this.charts.set('txVolume', new Chart(ctx, {
            type: 'bar',
            data: {
                labels: chartData.labels,
                datasets: [{
                    label: 'Transactions per Block',
                    data: chartData.values,
                    backgroundColor: this.chartColors.success,
                    borderColor: this.chartColors.success,
                    borderWidth: 1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: false
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        title: {
                            display: true,
                            text: 'Transaction Count'
                        }
                    },
                    x: {
                        title: {
                            display: true,
                            text: 'Block Height'
                        }
                    }
                }
            }
        }));
    }

    createDifficultyChart(data) {
        const ctx = document.getElementById('difficultyChart');
        if (!ctx) return;

        const chartData = data.difficulty_data || this.generateSampleDifficultyData();
        
        this.charts.set('difficulty', new Chart(ctx, {
            type: 'line',
            data: {
                labels: chartData.labels,
                datasets: [{
                    label: 'Difficulty',
                    data: chartData.values,
                    borderColor: this.chartColors.warning,
                    backgroundColor: this.chartColors.warning + '20',
                    borderWidth: 2,
                    fill: true,
                    tension: 0.4
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: false
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        title: {
                            display: true,
                            text: 'Difficulty'
                        }
                    },
                    x: {
                        title: {
                            display: true,
                            text: 'Block Height'
                        }
                    }
                }
            }
        }));
    }

    createAddressGrowthChart(data) {
        const ctx = document.getElementById('addressChart');
        if (!ctx) return;

        const chartData = data.address_data || this.generateSampleAddressData();
        
        this.charts.set('addressGrowth', new Chart(ctx, {
            type: 'line',
            data: {
                labels: chartData.labels,
                datasets: [{
                    label: 'Total Addresses',
                    data: chartData.values,
                    borderColor: this.chartColors.info,
                    backgroundColor: this.chartColors.info + '20',
                    borderWidth: 2,
                    fill: true,
                    tension: 0.4
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: false
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        title: {
                            display: true,
                            text: 'Address Count'
                        }
                    },
                    x: {
                        title: {
                            display: true,
                            text: 'Time'
                        }
                    }
                }
            }
        }));
    }

    // Sample data generators for demonstration
    generateSampleBlockTimeData() {
        const labels = [];
        const values = [];
        for (let i = 0; i < 20; i++) {
            labels.push(`Block ${1000 + i * 50}`);
            values.push(Math.random() * 600 + 300); // 300-900 seconds
        }
        return { labels, values };
    }

    generateSampleTxVolumeData() {
        const labels = [];
        const values = [];
        for (let i = 0; i < 20; i++) {
            labels.push(`Block ${1000 + i * 50}`);
            values.push(Math.floor(Math.random() * 1000 + 100)); // 100-1100 transactions
        }
        return { labels, values };
    }

    generateSampleDifficultyData() {
        const labels = [];
        const values = [];
        let difficulty = 1000000;
        for (let i = 0; i < 20; i++) {
            labels.push(`Block ${1000 + i * 50}`);
            difficulty += Math.random() * 100000 - 50000;
            values.push(Math.max(100000, difficulty));
        }
        return { labels, values };
    }

    generateSampleAddressData() {
        const labels = [];
        const values = [];
        let addresses = 10000;
        for (let i = 0; i < 20; i++) {
            labels.push(`Day ${i + 1}`);
            addresses += Math.floor(Math.random() * 1000 + 100);
            values.push(addresses);
        }
        return { labels, values };
    }

    // Update charts with real-time data
    updateCharts(newData) {
        this.charts.forEach((chart, key) => {
            if (newData[key]) {
                chart.data.datasets[0].data = newData[key];
                chart.update('none');
            }
        });
    }

    // Resize charts on window resize
    resizeCharts() {
        this.charts.forEach(chart => {
            chart.resize();
        });
    }

    // Destroy charts
    destroyCharts() {
        this.charts.forEach(chart => {
            chart.destroy();
        });
        this.charts.clear();
    }
}

// Export for use in main app
if (typeof module !== 'undefined' && module.exports) {
    module.exports = BlockchainCharts;
}
