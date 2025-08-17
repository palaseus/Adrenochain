// adrenochain Explorer Performance Monitoring Module

class PerformanceMonitor {
    constructor() {
        this.metrics = {
            responseTimes: {},
            throughput: {},
            errors: {},
            usage: {},
            system: {}
        };
        this.startTime = performance.now();
        this.init();
    }

    init() {
        this.setupPerformanceObserver();
        this.setupMetricsCollection();
        this.setupBenchmarks();
    }

    setupPerformanceObserver() {
        // Monitor navigation timing
        if ('PerformanceObserver' in window) {
            const observer = new PerformanceObserver((list) => {
                for (const entry of list.getEntries()) {
                    this.recordResponseTime(entry.name, entry.duration);
                }
            });
            observer.observe({ entryTypes: ['navigation', 'resource'] });
        }

        // Monitor long tasks
        if ('PerformanceObserver' in window) {
            const longTaskObserver = new PerformanceObserver((list) => {
                for (const entry of list.getEntries()) {
                    if (entry.duration > 50) { // 50ms threshold
                        this.recordLongTask(entry.duration);
                    }
                }
            });
            longTaskObserver.observe({ entryTypes: ['longtask'] });
        }
    }

    setupMetricsCollection() {
        // Collect performance metrics every 5 seconds
        setInterval(() => {
            this.collectSystemMetrics();
            this.collectUsageMetrics();
        }, 5000);

        // Monitor memory usage
        if ('memory' in performance) {
            setInterval(() => {
                this.recordMemoryUsage();
            }, 10000);
        }
    }

    setupBenchmarks() {
        // Run performance benchmarks
        this.runBenchmarks();
    }

    recordResponseTime(endpoint, duration) {
        if (!this.metrics.responseTimes[endpoint]) {
            this.metrics.responseTimes[endpoint] = [];
        }
        this.metrics.responseTimes[endpoint].push(duration);

        // Keep only last 100 measurements
        if (this.metrics.responseTimes[endpoint].length > 100) {
            this.metrics.responseTimes[endpoint].shift();
        }

        // Check performance targets
        this.checkPerformanceTargets(endpoint, duration);
    }

    checkPerformanceTargets(endpoint, duration) {
        const targets = {
            '/': 200,
            '/blocks': 300,
            '/transactions': 400,
            '/addresses': 500,
            '/search': 300
        };

        if (targets[endpoint] && duration > targets[endpoint]) {
            console.warn(`Performance target exceeded for ${endpoint}: ${duration}ms > ${targets[endpoint]}ms`);
            this.recordPerformanceViolation(endpoint, duration, targets[endpoint]);
        }
    }

    recordPerformanceViolation(endpoint, actual, target) {
        if (!this.metrics.errors.performanceViolations) {
            this.metrics.errors.performanceViolations = [];
        }
        this.metrics.errors.performanceViolations.push({
            endpoint,
            actual,
            target,
            timestamp: Date.now()
        });
    }

    recordLongTask(duration) {
        if (!this.metrics.errors.longTasks) {
            this.metrics.errors.longTasks = [];
        }
        this.metrics.errors.longTasks.push({
            duration,
            timestamp: Date.now()
        });
    }

    recordMemoryUsage() {
        const memory = performance.memory;
        this.metrics.system.memory = {
            used: memory.usedJSHeapSize,
            total: memory.totalJSHeapSize,
            limit: memory.jsHeapSizeLimit,
            timestamp: Date.now()
        };
    }

    collectSystemMetrics() {
        // CPU usage estimation
        const now = performance.now();
        const timeSinceStart = now - this.startTime;
        
        // Simple CPU estimation based on time
        this.metrics.system.cpu = {
            load: this.estimateCPULoad(),
            timestamp: now
        };

        // Network status
        if ('connection' in navigator) {
            this.metrics.system.network = {
                effectiveType: navigator.connection.effectiveType,
                downlink: navigator.connection.downlink,
                rtt: navigator.connection.rtt,
                timestamp: now
            };
        }
    }

    estimateCPULoad() {
        // Simple CPU load estimation
        const longTasks = this.metrics.errors.longTasks || [];
        const recentLongTasks = longTasks.filter(task => 
            Date.now() - task.timestamp < 60000
        );
        return Math.min(1, recentLongTasks.length / 10);
    }

    collectUsageMetrics() {
        // Page views
        if (!this.metrics.usage.pageViews) {
            this.metrics.usage.pageViews = 0;
        }
        this.metrics.usage.pageViews++;

        // User interactions
        this.metrics.usage.lastInteraction = Date.now();
    }

    runBenchmarks() {
        // Benchmark critical operations
        this.benchmarkTemplateRendering();
        this.benchmarkDataProcessing();
        this.benchmarkChartRendering();
    }

    async benchmarkTemplateRendering() {
        const start = performance.now();
        
        // Simulate template rendering
        const template = document.createElement('div');
        for (let i = 0; i < 1000; i++) {
            template.innerHTML = `<div>Block #${i}</div>`;
        }
        
        const duration = performance.now() - start;
        this.recordBenchmark('templateRendering', duration);
    }

    async benchmarkDataProcessing() {
        const start = performance.now();
        
        // Simulate data processing
        const data = Array.from({length: 10000}, (_, i) => ({
            hash: `hash${i}`,
            height: i,
            timestamp: Date.now()
        }));
        
        const processed = data.map(item => ({
            ...item,
            formatted: `Block #${item.height} at ${new Date(item.timestamp).toLocaleString()}`
        }));
        
        const duration = performance.now() - start;
        this.recordBenchmark('dataProcessing', duration);
    }

    async benchmarkChartRendering() {
        const start = performance.now();
        
        // Simulate chart rendering
        const canvas = document.createElement('canvas');
        const ctx = canvas.getContext('2d');
        if (ctx) {
            for (let i = 0; i < 100; i++) {
                ctx.fillStyle = `hsl(${i * 3.6}, 70%, 50%)`;
                ctx.fillRect(i * 4, 0, 4, 100);
            }
        }
        
        const duration = performance.now() - start;
        this.recordBenchmark('chartRendering', duration);
    }

    recordBenchmark(name, duration) {
        if (!this.metrics.throughput.benchmarks) {
            this.metrics.throughput.benchmarks = {};
        }
        if (!this.metrics.throughput.benchmarks[name]) {
            this.metrics.throughput.benchmarks[name] = [];
        }
        this.metrics.throughput.benchmarks[name].push(duration);
    }

    getMetrics() {
        return {
            ...this.metrics,
            uptime: performance.now() - this.startTime,
            timestamp: Date.now()
        };
    }

    getPerformanceReport() {
        const report = {
            summary: {},
            details: this.metrics,
            recommendations: []
        };

        // Calculate averages
        Object.keys(this.metrics.responseTimes).forEach(endpoint => {
            const times = this.metrics.responseTimes[endpoint];
            const avg = times.reduce((a, b) => a + b, 0) / times.length;
            const max = Math.max(...times);
            const min = Math.min(...times);
            
            report.summary[endpoint] = { avg, max, min, samples: times.length };
        });

        // Generate recommendations
        if (this.metrics.errors.performanceViolations?.length > 0) {
            report.recommendations.push('Performance targets exceeded - consider optimization');
        }
        if (this.metrics.errors.longTasks?.length > 0) {
            report.recommendations.push('Long tasks detected - consider breaking up heavy operations');
        }
        if (this.metrics.system.memory?.used > this.metrics.system.memory?.limit * 0.8) {
            report.recommendations.push('High memory usage - consider memory optimization');
        }

        return report;
    }

    exportMetrics() {
        const data = this.getMetrics();
        const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `explorer-performance-${Date.now()}.json`;
        a.click();
        URL.revokeObjectURL(url);
    }
}

// Initialize performance monitoring
window.performanceMonitor = new PerformanceMonitor();

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = PerformanceMonitor;
}
