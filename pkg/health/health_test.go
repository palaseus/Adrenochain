package health

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSystemHealth(t *testing.T) {
	version := "1.0.0"
	sh := NewSystemHealth(version)

	assert.NotNil(t, sh)
	assert.Equal(t, version, sh.version)
	assert.Equal(t, 0, sh.GetComponentCount())
	assert.Equal(t, StatusUnknown, sh.GetOverallStatus())
}

func TestRegisterComponent(t *testing.T) {
	sh := NewSystemHealth("1.0.0")

	// Create a mock health checker
	mockChecker := &MockHealthChecker{
		name:    "test-component",
		status:  StatusHealthy,
		message: "Test component is healthy",
	}

	sh.RegisterComponent(mockChecker)

	assert.Equal(t, 1, sh.GetComponentCount())
	assert.True(t, sh.IsHealthy())

	components := sh.GetRegisteredComponents()
	assert.Contains(t, components, "test-component")
}

func TestUnregisterComponent(t *testing.T) {
	sh := NewSystemHealth("1.0.0")

	mockChecker := &MockHealthChecker{
		name:    "test-component",
		status:  StatusHealthy,
		message: "Test component is healthy",
	}

	sh.RegisterComponent(mockChecker)
	assert.Equal(t, 1, sh.GetComponentCount())

	sh.UnregisterComponent("test-component")
	assert.Equal(t, 0, sh.GetComponentCount())
	assert.Equal(t, StatusUnknown, sh.GetOverallStatus())
}

func TestUpdateComponent(t *testing.T) {
	sh := NewSystemHealth("1.0.0")

	mockChecker := &MockHealthChecker{
		name:    "test-component",
		status:  StatusHealthy,
		message: "Test component is healthy",
	}

	sh.RegisterComponent(mockChecker)

	// Update component status
	details := map[string]interface{}{
		"key":    "value",
		"number": 42,
	}
	sh.UpdateComponent("test-component", StatusDegraded, "Component is degraded", details)

	component, exists := sh.GetComponentStatus("test-component")
	require.True(t, exists)
	assert.Equal(t, StatusDegraded, component.Status)
	assert.Equal(t, "Component is degraded", component.Message)
	assert.Equal(t, details, component.Details)
}

func TestGetOverallStatus(t *testing.T) {
	sh := NewSystemHealth("1.0.0")

	// No components - should be unknown
	assert.Equal(t, StatusUnknown, sh.GetOverallStatus())

	// Add healthy component
	healthyChecker := &MockHealthChecker{
		name:    "healthy",
		status:  StatusHealthy,
		message: "Healthy",
	}
	sh.RegisterComponent(healthyChecker)
	assert.Equal(t, StatusHealthy, sh.GetOverallStatus())

	// Add degraded component
	degradedChecker := &MockHealthChecker{
		name:    "degraded",
		status:  StatusDegraded,
		message: "Degraded",
	}
	sh.RegisterComponent(degradedChecker)
	assert.Equal(t, StatusDegraded, sh.GetOverallStatus())

	// Add unhealthy component
	unhealthyChecker := &MockHealthChecker{
		name:    "unhealthy",
		status:  StatusUnhealthy,
		message: "Unhealthy",
	}
	sh.RegisterComponent(unhealthyChecker)
	assert.Equal(t, StatusUnhealthy, sh.GetOverallStatus())
}

func TestRunHealthChecks(t *testing.T) {
	sh := NewSystemHealth("1.0.0")

	// Add a mock checker that takes time
	slowChecker := &MockHealthChecker{
		name:      "slow-component",
		status:    StatusHealthy,
		message:   "Slow but healthy",
		checkTime: 100 * time.Millisecond,
	}

	sh.RegisterComponent(slowChecker)

	start := time.Now()
	sh.RunHealthChecks()
	duration := time.Since(start)

	// Should complete quickly (parallel execution)
	assert.Less(t, duration, 200*time.Millisecond)

	// Component should be updated
	component, exists := sh.GetComponentStatus("slow-component")
	require.True(t, exists)
	assert.True(t, component.CheckTime > 0)
}

func TestGetHealthReport(t *testing.T) {
	sh := NewSystemHealth("1.0.0")

	mockChecker := &MockHealthChecker{
		name:    "test-component",
		status:  StatusHealthy,
		message: "Test component is healthy",
	}

	sh.RegisterComponent(mockChecker)

	report := sh.GetHealthReport()

	// Verify report structure
	assert.Contains(t, report, "status")
	assert.Contains(t, report, "version")
	assert.Contains(t, report, "uptime")
	assert.Contains(t, report, "start_time")
	assert.Contains(t, report, "components")
	assert.Contains(t, report, "system")

	// Verify system info
	system := report["system"].(map[string]interface{})
	assert.Contains(t, system, "go_version")
	assert.Contains(t, system, "go_os")
	assert.Contains(t, system, "go_arch")
	assert.Contains(t, system, "num_goroutines")
	assert.Contains(t, system, "memory")

	// Verify components
	components := report["components"].(map[string]*Component)
	assert.Contains(t, components, "test-component")
}

func TestGetHealthJSON(t *testing.T) {
	sh := NewSystemHealth("1.0.0")

	mockChecker := &MockHealthChecker{
		name:    "test-component",
		status:  StatusHealthy,
		message: "Test component is healthy",
	}

	sh.RegisterComponent(mockChecker)

	jsonData, err := sh.GetHealthJSON()
	require.NoError(t, err)

	// Verify it's valid JSON
	assert.True(t, len(jsonData) > 0)
	assert.Contains(t, string(jsonData), "test-component")
}

func TestComponentStatus(t *testing.T) {
	sh := NewSystemHealth("1.0.0")

	mockChecker := &MockHealthChecker{
		name:    "test-component",
		status:  StatusHealthy,
		message: "Test component is healthy",
	}

	sh.RegisterComponent(mockChecker)

	// Get component status
	component, exists := sh.GetComponentStatus("test-component")
	require.True(t, exists)
	assert.Equal(t, "test-component", component.Name)
	assert.Equal(t, StatusHealthy, component.Status)
	assert.Equal(t, "Test component is healthy", component.Message)

	// Get non-existent component
	_, exists = sh.GetComponentStatus("non-existent")
	assert.False(t, exists)
}

// MockHealthChecker is a mock implementation for testing
type MockHealthChecker struct {
	name      string
	status    Status
	message   string
	details   map[string]interface{}
	checkTime time.Duration
}

func (m *MockHealthChecker) Name() string {
	return m.name
}

func (m *MockHealthChecker) Check() (*Component, error) {
	if m.checkTime > 0 {
		time.Sleep(m.checkTime)
	}

	return &Component{
		Name:      m.name,
		Status:    m.status,
		Message:   m.message,
		Details:   m.details,
		LastCheck: time.Now(),
		CheckTime: m.checkTime,
	}, nil
}

func TestStatusConstants(t *testing.T) {
	// Test that status constants are properly defined
	assert.Equal(t, Status("healthy"), StatusHealthy)
	assert.Equal(t, Status("degraded"), StatusDegraded)
	assert.Equal(t, Status("unhealthy"), StatusUnhealthy)
	assert.Equal(t, Status("unknown"), StatusUnknown)
}

func TestComponentFields(t *testing.T) {
	details := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}

	component := &Component{
		Name:      "test",
		Status:    StatusHealthy,
		Message:   "Test message",
		Details:   details,
		LastCheck: time.Now(),
		CheckTime: 100 * time.Millisecond,
	}

	assert.Equal(t, "test", component.Name)
	assert.Equal(t, StatusHealthy, component.Status)
	assert.Equal(t, "Test message", component.Message)
	assert.Equal(t, details, component.Details)
	assert.True(t, component.LastCheck.After(time.Time{}))
	assert.Equal(t, 100*time.Millisecond, component.CheckTime)
}
