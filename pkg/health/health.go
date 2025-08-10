package health

import (
	"encoding/json"
	"runtime"
	"sync"
	"time"
)

// Status represents the health status of a component
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
	StatusUnknown   Status = "unknown"
)

// Component represents a health checkable component
type Component struct {
	Name      string                 `json:"name"`
	Status    Status                 `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	LastCheck time.Time              `json:"last_check"`
	CheckTime time.Duration          `json:"check_time_ms"`
}

// HealthChecker defines the interface for health checkable components
type HealthChecker interface {
	Name() string
	Check() (*Component, error)
}

// SystemHealth represents the overall health of the system
type SystemHealth struct {
	mu         sync.RWMutex
	components map[string]*Component
	checkers   map[string]HealthChecker
	startTime  time.Time
	version    string
}

// NewSystemHealth creates a new system health checker
func NewSystemHealth(version string) *SystemHealth {
	return &SystemHealth{
		components: make(map[string]*Component),
		checkers:   make(map[string]HealthChecker),
		startTime:  time.Now(),
		version:    version,
	}
}

// RegisterComponent registers a health checkable component
func (sh *SystemHealth) RegisterComponent(checker HealthChecker) {
	sh.mu.Lock()
	defer sh.mu.Unlock()
	
	// Store the checker
	sh.checkers[checker.Name()] = checker
	
	// Perform initial check
	component, _ := checker.Check()
	sh.components[checker.Name()] = component
}

// UnregisterComponent removes a health checkable component
func (sh *SystemHealth) UnregisterComponent(name string) {
	sh.mu.Lock()
	defer sh.mu.Unlock()
	
	delete(sh.checkers, name)
	delete(sh.components, name)
}

// UpdateComponent updates the health status of a component
func (sh *SystemHealth) UpdateComponent(name string, status Status, message string, details map[string]interface{}) {
	sh.mu.Lock()
	defer sh.mu.Unlock()
	
	if component, exists := sh.components[name]; exists {
		component.Status = status
		component.Message = message
		component.Details = details
		component.LastCheck = time.Now()
	}
}

// GetOverallStatus returns the overall health status
func (sh *SystemHealth) GetOverallStatus() Status {
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	
	if len(sh.components) == 0 {
		return StatusUnknown
	}
	
	unhealthyCount := 0
	degradedCount := 0
	
	for _, component := range sh.components {
		switch component.Status {
		case StatusUnhealthy:
			unhealthyCount++
		case StatusDegraded:
			degradedCount++
		}
	}
	
	if unhealthyCount > 0 {
		return StatusUnhealthy
	}
	
	if degradedCount > 0 {
		return StatusDegraded
	}
	
	return StatusHealthy
}

// GetHealthReport returns a comprehensive health report
func (sh *SystemHealth) GetHealthReport() map[string]interface{} {
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	
	uptime := time.Since(sh.startTime)
	
	// Get system metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	report := map[string]interface{}{
		"status":      sh.GetOverallStatus(),
		"version":     sh.version,
		"uptime":      uptime.String(),
		"start_time":  sh.startTime,
		"components":  make(map[string]*Component),
		"system": map[string]interface{}{
			"go_version":    runtime.Version(),
			"go_os":         runtime.GOOS,
			"go_arch":       runtime.GOARCH,
			"num_goroutines": runtime.NumGoroutine(),
			"memory": map[string]interface{}{
				"alloc":      m.Alloc,
				"total_alloc": m.TotalAlloc,
				"sys":         m.Sys,
				"num_gc":      m.NumGC,
			},
		},
	}
	
	// Copy components
	for name, component := range sh.components {
		report["components"].(map[string]*Component)[name] = component
	}
	
	return report
}

// RunHealthChecks runs health checks for all registered components
func (sh *SystemHealth) RunHealthChecks() {
	sh.mu.RLock()
	checkers := make([]HealthChecker, 0, len(sh.checkers))
	for _, checker := range sh.checkers {
		checkers = append(checkers, checker)
	}
	sh.mu.RUnlock()
	
	// Run checks in parallel
	var wg sync.WaitGroup
	for _, checker := range checkers {
		wg.Add(1)
		go func(c HealthChecker) {
			defer wg.Done()
			sh.runComponentCheck(c)
		}(checker)
	}
	wg.Wait()
}

// runComponentCheck runs a health check for a single component
func (sh *SystemHealth) runComponentCheck(checker HealthChecker) {
	start := time.Now()
	component, err := checker.Check()
	checkTime := time.Since(start)
	
	if err != nil {
		component.Status = StatusUnhealthy
		component.Message = err.Error()
	}
	
	component.LastCheck = time.Now()
	component.CheckTime = checkTime
	
	sh.mu.Lock()
	sh.components[checker.Name()] = component
	sh.mu.Unlock()
}

// GetHealthJSON returns the health report as JSON
func (sh *SystemHealth) GetHealthJSON() ([]byte, error) {
	report := sh.GetHealthReport()
	return json.MarshalIndent(report, "", "  ")
}

// IsHealthy returns true if the overall status is healthy
func (sh *SystemHealth) IsHealthy() bool {
	return sh.GetOverallStatus() == StatusHealthy
}

// GetComponentStatus returns the status of a specific component
func (sh *SystemHealth) GetComponentStatus(name string) (*Component, bool) {
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	
	component, exists := sh.components[name]
	return component, exists
}

// GetRegisteredComponents returns a list of all registered component names
func (sh *SystemHealth) GetRegisteredComponents() []string {
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	
	components := make([]string, 0, len(sh.checkers))
	for name := range sh.checkers {
		components = append(components, name)
	}
	return components
}

// GetComponentCount returns the number of registered components
func (sh *SystemHealth) GetComponentCount() int {
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	
	return len(sh.checkers)
}
