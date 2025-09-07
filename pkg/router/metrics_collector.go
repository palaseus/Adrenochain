package router

import (
	"sync"
	"time"
)

// MetricsCollector collects and stores performance metrics
type MetricsCollector struct {
	mu          sync.RWMutex
	metrics     []*ClusterMetrics
	maxHistory  int
	lastCleanup time.Time
}

// ClusterMetrics represents performance metrics for the cluster router
type ClusterMetrics struct {
	Timestamp          time.Time             `json:"timestamp"`
	TotalRequests      int64                 `json:"total_requests"`
	SuccessfulRequests int64                 `json:"successful_requests"`
	FailedRequests     int64                 `json:"failed_requests"`
	AvgResponseTime    time.Duration         `json:"avg_response_time"`
	ClusterCount       int                   `json:"cluster_count"`
	NodeCount          int                   `json:"node_count"`
	ActiveClusters     int                   `json:"active_clusters"`
	ActiveNodes        int                   `json:"active_nodes"`
	UnhealthyNodes     int                   `json:"unhealthy_nodes"`
	LoadBalancerStats  *LoadBalancerMetrics  `json:"load_balancer_stats"`
	HealthMonitorStats *HealthMonitorMetrics `json:"health_monitor_stats"`
	RoutingTableStats  *RoutingTableMetrics  `json:"routing_table_stats"`
}

// LoadBalancerMetrics represents metrics for the load balancer
type LoadBalancerMetrics struct {
	Strategy           RoutingStrategy    `json:"strategy"`
	TotalConnections   int64              `json:"total_connections"`
	ActiveConnections  int64              `json:"active_connections"`
	ConnectionFailures int64              `json:"connection_failures"`
	AvgSelectionTime   time.Duration      `json:"avg_selection_time"`
	NodeWeights        map[NodeID]float64 `json:"node_weights"`
}

// HealthMonitorMetrics represents metrics for the health monitor
type HealthMonitorMetrics struct {
	TotalChecks      int64         `json:"total_checks"`
	SuccessfulChecks int64         `json:"successful_checks"`
	FailedChecks     int64         `json:"failed_checks"`
	AvgCheckLatency  time.Duration `json:"avg_check_latency"`
	CheckInterval    time.Duration `json:"check_interval"`
	Timeout          time.Duration `json:"timeout"`
}

// RoutingTableMetrics represents metrics for the routing table
type RoutingTableMetrics struct {
	TotalClusters  int   `json:"total_clusters"`
	TotalNodes     int   `json:"total_nodes"`
	ActiveClusters int   `json:"active_clusters"`
	ActiveNodes    int   `json:"active_nodes"`
	IndexCount     int   `json:"index_count"`
	LookupCount    int64 `json:"lookup_count"`
	UpdateCount    int64 `json:"update_count"`
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics:     make([]*ClusterMetrics, 0),
		maxHistory:  1000, // Keep last 1000 metrics
		lastCleanup: time.Now(),
	}
}

// RecordMetrics records new metrics
func (mc *MetricsCollector) RecordMetrics(metrics *ClusterMetrics) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Add timestamp if not set
	if metrics.Timestamp.IsZero() {
		metrics.Timestamp = time.Now()
	}

	// Add to history
	mc.metrics = append(mc.metrics, metrics)

	// Cleanup old metrics if needed
	if len(mc.metrics) > mc.maxHistory {
		mc.metrics = mc.metrics[len(mc.metrics)-mc.maxHistory:]
	}

	// Periodic cleanup
	if time.Since(mc.lastCleanup) > 1*time.Hour {
		mc.cleanupOldMetrics()
		mc.lastCleanup = time.Now()
	}
}

// GetLatestMetrics returns the most recent metrics
func (mc *MetricsCollector) GetLatestMetrics() *ClusterMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if len(mc.metrics) == 0 {
		return nil
	}

	return mc.metrics[len(mc.metrics)-1]
}

// GetMetricsHistory returns metrics history within a time range
func (mc *MetricsCollector) GetMetricsHistory(start, end time.Time) []*ClusterMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	var result []*ClusterMetrics
	for _, metrics := range mc.metrics {
		if metrics.Timestamp.After(start) && metrics.Timestamp.Before(end) {
			result = append(result, metrics)
		}
	}

	return result
}

// GetMetricsByTimeRange returns metrics within a time range with limit
func (mc *MetricsCollector) GetMetricsByTimeRange(start, end time.Time, limit int) []*ClusterMetrics {
	history := mc.GetMetricsHistory(start, end)

	if limit <= 0 || limit >= len(history) {
		return history
	}

	// Return the most recent metrics within the limit
	startIndex := len(history) - limit
	if startIndex < 0 {
		startIndex = 0
	}

	return history[startIndex:]
}

// GetAverageMetrics returns average metrics over a time period
func (mc *MetricsCollector) GetAverageMetrics(duration time.Duration) *ClusterMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	cutoff := time.Now().Add(-duration)
	var relevantMetrics []*ClusterMetrics

	for _, metrics := range mc.metrics {
		if metrics.Timestamp.After(cutoff) {
			relevantMetrics = append(relevantMetrics, metrics)
		}
	}

	if len(relevantMetrics) == 0 {
		return nil
	}

	// Calculate averages
	avgMetrics := &ClusterMetrics{
		Timestamp: time.Now(),
	}

	totalRequests := int64(0)
	totalSuccessful := int64(0)
	totalFailed := int64(0)
	totalResponseTime := time.Duration(0)

	for _, metrics := range relevantMetrics {
		totalRequests += metrics.TotalRequests
		totalSuccessful += metrics.SuccessfulRequests
		totalFailed += metrics.FailedRequests
		totalResponseTime += metrics.AvgResponseTime
	}

	count := int64(len(relevantMetrics))
	avgMetrics.TotalRequests = totalRequests / count
	avgMetrics.SuccessfulRequests = totalSuccessful / count
	avgMetrics.FailedRequests = totalFailed / count
	avgMetrics.AvgResponseTime = totalResponseTime / time.Duration(count)

	// Calculate other averages
	avgClusterCount := 0
	avgNodeCount := 0
	avgActiveClusters := 0
	avgActiveNodes := 0
	avgUnhealthyNodes := 0

	for _, metrics := range relevantMetrics {
		avgClusterCount += metrics.ClusterCount
		avgNodeCount += metrics.NodeCount
		avgActiveClusters += metrics.ActiveClusters
		avgActiveNodes += metrics.ActiveNodes
		avgUnhealthyNodes += metrics.UnhealthyNodes
	}

	avgMetrics.ClusterCount = avgClusterCount / len(relevantMetrics)
	avgMetrics.NodeCount = avgNodeCount / len(relevantMetrics)
	avgMetrics.ActiveClusters = avgActiveClusters / len(relevantMetrics)
	avgMetrics.ActiveNodes = avgActiveNodes / len(relevantMetrics)
	avgMetrics.UnhealthyNodes = avgUnhealthyNodes / len(relevantMetrics)

	return avgMetrics
}

// GetSuccessRate returns the success rate over a time period
func (mc *MetricsCollector) GetSuccessRate(duration time.Duration) float64 {
	avgMetrics := mc.GetAverageMetrics(duration)
	if avgMetrics == nil || avgMetrics.TotalRequests == 0 {
		return 0.0
	}

	return float64(avgMetrics.SuccessfulRequests) / float64(avgMetrics.TotalRequests)
}

// GetErrorRate returns the error rate over a time period
func (mc *MetricsCollector) GetErrorRate(duration time.Duration) float64 {
	avgMetrics := mc.GetAverageMetrics(duration)
	if avgMetrics == nil || avgMetrics.TotalRequests == 0 {
		return 0.0
	}

	return float64(avgMetrics.FailedRequests) / float64(avgMetrics.TotalRequests)
}

// GetThroughput returns the throughput (requests per second) over a time period
func (mc *MetricsCollector) GetThroughput(duration time.Duration) float64 {
	avgMetrics := mc.GetAverageMetrics(duration)
	if avgMetrics == nil {
		return 0.0
	}

	return float64(avgMetrics.TotalRequests) / duration.Seconds()
}

// GetLatencyPercentile returns the latency at a specific percentile
func (mc *MetricsCollector) GetLatencyPercentile(duration time.Duration, percentile float64) time.Duration {
	history := mc.GetMetricsHistory(time.Now().Add(-duration), time.Now())
	if len(history) == 0 {
		return 0
	}

	// Collect all response times
	var responseTimes []time.Duration
	for _, metrics := range history {
		responseTimes = append(responseTimes, metrics.AvgResponseTime)
	}

	// Sort response times
	for i := 0; i < len(responseTimes)-1; i++ {
		for j := i + 1; j < len(responseTimes); j++ {
			if responseTimes[i] > responseTimes[j] {
				responseTimes[i], responseTimes[j] = responseTimes[j], responseTimes[i]
			}
		}
	}

	// Calculate percentile
	index := int(float64(len(responseTimes)) * percentile / 100.0)
	if index >= len(responseTimes) {
		index = len(responseTimes) - 1
	}

	return responseTimes[index]
}

// GetTopNodesByLoad returns the top N nodes by load
func (mc *MetricsCollector) GetTopNodesByLoad(limit int) []NodeLoadInfo {
	// This would require additional data structures to track per-node metrics
	// For now, return empty slice
	return []NodeLoadInfo{}
}

// GetTopNodesByLatency returns the top N nodes by latency
func (mc *MetricsCollector) GetTopNodesByLatency(limit int) []NodeLatencyInfo {
	// This would require additional data structures to track per-node metrics
	// For now, return empty slice
	return []NodeLatencyInfo{}
}

// cleanupOldMetrics removes metrics older than the retention period
func (mc *MetricsCollector) cleanupOldMetrics() {
	cutoff := time.Now().Add(-24 * time.Hour) // Keep 24 hours of data
	var filtered []*ClusterMetrics

	for _, metrics := range mc.metrics {
		if metrics.Timestamp.After(cutoff) {
			filtered = append(filtered, metrics)
		}
	}

	mc.metrics = filtered
}

// SetMaxHistory sets the maximum number of metrics to keep in history
func (mc *MetricsCollector) SetMaxHistory(maxHistory int) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.maxHistory = maxHistory

	// Trim if necessary
	if len(mc.metrics) > maxHistory {
		mc.metrics = mc.metrics[len(mc.metrics)-maxHistory:]
	}
}

// GetStats returns statistics about the metrics collector
func (mc *MetricsCollector) GetStats() *MetricsCollectorStats {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	stats := &MetricsCollectorStats{
		TotalMetrics: len(mc.metrics),
		MaxHistory:   mc.maxHistory,
		LastCleanup:  mc.lastCleanup,
		OldestMetric: nil,
		NewestMetric: nil,
	}

	if len(mc.metrics) > 0 {
		stats.OldestMetric = mc.metrics[0]
		stats.NewestMetric = mc.metrics[len(mc.metrics)-1]
	}

	return stats
}

// Clear clears all metrics
func (mc *MetricsCollector) Clear() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metrics = make([]*ClusterMetrics, 0)
}

// MetricsCollectorStats contains statistics about the metrics collector
type MetricsCollectorStats struct {
	TotalMetrics int
	MaxHistory   int
	LastCleanup  time.Time
	OldestMetric *ClusterMetrics
	NewestMetric *ClusterMetrics
}

// NodeLoadInfo contains load information for a node
type NodeLoadInfo struct {
	NodeID NodeID
	Load   float64
}

// NodeLatencyInfo contains latency information for a node
type NodeLatencyInfo struct {
	NodeID  NodeID
	Latency time.Duration
}
