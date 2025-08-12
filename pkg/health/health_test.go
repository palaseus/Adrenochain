package health

import (
	"encoding/json"
	"fmt"
	"sync"
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

// Test edge cases and error scenarios
func TestHealthEdgeCases(t *testing.T) {
	sh := NewSystemHealth("1.0.0")

	// Test unregistering non-existent component
	sh.UnregisterComponent("non-existent")
	assert.Equal(t, 0, sh.GetComponentCount())

	// Test updating non-existent component
	sh.UpdateComponent("non-existent", StatusUnhealthy, "Not found", nil)
	component, exists := sh.GetComponentStatus("non-existent")
	assert.False(t, exists)
	assert.Nil(t, component)

	// Test with nil details
	sh.UpdateComponent("test", StatusHealthy, "No details", nil)
	component, exists = sh.GetComponentStatus("test")
	assert.False(t, exists) // Component wasn't registered
}

// Test concurrent access
func TestHealthConcurrency(t *testing.T) {
	sh := NewSystemHealth("1.0.0")

	// Create multiple mock checkers
	checkers := make([]*MockHealthChecker, 10)
	for i := 0; i < 10; i++ {
		checkers[i] = &MockHealthChecker{
			name:    fmt.Sprintf("component-%d", i),
			status:  StatusHealthy,
			message: fmt.Sprintf("Component %d is healthy", i),
		}
	}

	// Register components concurrently
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(checker *MockHealthChecker) {
			defer wg.Done()
			sh.RegisterComponent(checker)
		}(checkers[i])
	}
	wg.Wait()

	assert.Equal(t, 10, sh.GetComponentCount())
	assert.Equal(t, StatusHealthy, sh.GetOverallStatus())
}

// Test health report edge cases
func TestHealthReportEdgeCases(t *testing.T) {
	sh := NewSystemHealth("1.0.0")

	// Test empty health report
	report := sh.GetHealthReport()
	assert.NotNil(t, report)
	assert.Equal(t, StatusUnknown, report["status"])
	assert.Equal(t, "1.0.0", report["version"])
	assert.NotNil(t, report["system"])
	assert.NotNil(t, report["components"])

	// Test health report with components
	mockChecker := &MockHealthChecker{
		name:    "test-component",
		status:  StatusHealthy,
		message: "Test component is healthy",
	}
	sh.RegisterComponent(mockChecker)

	report = sh.GetHealthReport()
	assert.Equal(t, StatusHealthy, report["status"])
	assert.NotNil(t, report["components"].(map[string]*Component)["test-component"])
}

// Test JSON marshaling edge cases
func TestHealthJSONEdgeCases(t *testing.T) {
	sh := NewSystemHealth("1.0.0")

	// Test empty JSON
	jsonData, err := sh.GetHealthJSON()
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Test JSON with components
	mockChecker := &MockHealthChecker{
		name:    "test-component",
		status:  StatusHealthy,
		message: "Test component is healthy",
	}
	sh.RegisterComponent(mockChecker)

	jsonData, err = sh.GetHealthJSON()
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Verify JSON can be unmarshaled
	var report map[string]interface{}
	err = json.Unmarshal(jsonData, &report)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", report["status"])
}

// Test component status retrieval edge cases
func TestComponentStatusEdgeCases(t *testing.T) {
	sh := NewSystemHealth("1.0.0")

	// Test getting status of non-existent component
	component, exists := sh.GetComponentStatus("non-existent")
	assert.False(t, exists)
	assert.Nil(t, component)

	// Test getting status of registered component
	mockChecker := &MockHealthChecker{
		name:    "test-component",
		status:  StatusHealthy,
		message: "Test component is healthy",
	}
	sh.RegisterComponent(mockChecker)

	component, exists = sh.GetComponentStatus("test-component")
	assert.True(t, exists)
	assert.NotNil(t, component)
	assert.Equal(t, "test-component", component.Name)
}

// Test overall status edge cases
func TestOverallStatusEdgeCases(t *testing.T) {
	sh := NewSystemHealth("1.0.0")

	// Test with no components
	assert.Equal(t, StatusUnknown, sh.GetOverallStatus())

	// Test with one healthy component
	healthyChecker := &MockHealthChecker{
		name:    "healthy",
		status:  StatusHealthy,
		message: "Healthy",
	}
	sh.RegisterComponent(healthyChecker)
	assert.Equal(t, StatusHealthy, sh.GetOverallStatus())

	// Test with one degraded component
	degradedChecker := &MockHealthChecker{
		name:    "degraded",
		status:  StatusDegraded,
		message: "Degraded",
	}
	sh.RegisterComponent(degradedChecker)
	assert.Equal(t, StatusDegraded, sh.GetOverallStatus())

	// Test with one unhealthy component
	unhealthyChecker := &MockHealthChecker{
		name:    "unhealthy",
		status:  StatusUnhealthy,
		message: "Unhealthy",
	}
	sh.RegisterComponent(unhealthyChecker)
	assert.Equal(t, StatusUnhealthy, sh.GetOverallStatus())

	// Test with mixed statuses (unhealthy should take precedence)
	sh.UnregisterComponent("unhealthy")
	sh.RegisterComponent(&MockHealthChecker{
		name:    "healthy2",
		status:  StatusHealthy,
		message: "Healthy 2",
	})
	assert.Equal(t, StatusDegraded, sh.GetOverallStatus())
}

// Test blockchain health checker
func TestChainHealthChecker(t *testing.T) {
	// Test with mock chain
	mockChain := &MockChain{
		height:       100,
		bestBlock:    &MockBlock{height: 100, timestamp: time.Now(), difficulty: 1000, hash: []byte("mock_hash_12345")},
		genesisBlock: &MockBlock{height: 0, timestamp: time.Now().Add(-24 * time.Hour)},
	}

	checker := NewMockChainHealthChecker(mockChain)
	assert.NotNil(t, checker)
	assert.Equal(t, "blockchain", checker.Name())

	// Test health check
	component, err := checker.Check()
	assert.NoError(t, err)
	assert.NotNil(t, component)
	assert.Equal(t, "blockchain", component.Name)
	assert.Equal(t, StatusHealthy, component.Status)
	assert.Contains(t, component.Message, "healthy")
	assert.NotNil(t, component.Details)
	assert.Contains(t, component.Details, "height")
}

// Test blockchain health checker edge cases
func TestChainHealthCheckerEdgeCases(t *testing.T) {
	// Test with chain that has no best block
	mockChain := &MockChain{
		height:       0,
		bestBlock:    nil,
		genesisBlock: nil,
	}

	checker := NewMockChainHealthChecker(mockChain)
	component, err := checker.Check()
	assert.NoError(t, err)
	assert.Equal(t, StatusUnhealthy, component.Status)
	assert.Contains(t, component.Message, "No best block available")

	// Test with chain that has nil hash
	mockChain = &MockChain{
		height:       100,
		bestBlock:    &MockBlock{height: 100, timestamp: time.Now(), hash: nil, difficulty: 1000},
		genesisBlock: &MockBlock{height: 0, timestamp: time.Now().Add(-24 * time.Hour)},
	}

	checker = NewMockChainHealthChecker(mockChain)
	component, err = checker.Check()
	assert.NoError(t, err)
	assert.Equal(t, StatusUnhealthy, component.Status)
	assert.Contains(t, component.Message, "Failed to calculate best block hash")

	// Test with chain that has old block
	mockChain = &MockChain{
		height:       100,
		bestBlock:    &MockBlock{height: 100, timestamp: time.Now().Add(-2 * time.Hour), difficulty: 1000, hash: []byte("mock_hash_old")},
		genesisBlock: &MockBlock{height: 0, timestamp: time.Now().Add(-24 * time.Hour)},
	}

	checker = NewMockChainHealthChecker(mockChain)
	component, err = checker.Check()
	assert.NoError(t, err)
	assert.Equal(t, StatusDegraded, component.Status)
	assert.Contains(t, component.Message, "Last block is")

	// Test with chain that has zero difficulty
	mockChain = &MockChain{
		height:       100,
		bestBlock:    &MockBlock{height: 100, timestamp: time.Now(), difficulty: 0, hash: []byte("mock_hash_zero_diff")},
		genesisBlock: &MockBlock{height: 0, timestamp: time.Now().Add(-24 * time.Hour)},
	}

	checker = NewMockChainHealthChecker(mockChain)
	component, err = checker.Check()
	assert.NoError(t, err)
	assert.Equal(t, StatusDegraded, component.Status)
	assert.Contains(t, component.Message, "Block difficulty is zero or negative")
}

// Mock implementations for testing
type MockChain struct {
	height       uint64
	bestBlock    *MockBlock
	genesisBlock *MockBlock
}

func (m *MockChain) GetHeight() uint64 {
	return m.height
}

func (m *MockChain) GetBestBlock() *MockBlock {
	return m.bestBlock
}

func (m *MockChain) GetGenesisBlock() *MockBlock {
	return m.genesisBlock
}

type MockBlock struct {
	height       uint64
	timestamp    time.Time
	hash         []byte
	difficulty   uint64
	transactions []interface{}
}

func (m *MockBlock) CalculateHash() []byte {
	if m.hash == nil {
		// Return nil to simulate hash calculation failure
		return nil
	}
	return m.hash
}

func (m *MockBlock) Header() *MockHeader {
	return &MockHeader{
		Height:     m.height,
		Timestamp:  m.timestamp,
		Difficulty: m.difficulty,
	}
}

type MockHeader struct {
	Height     uint64
	Timestamp  time.Time
	Difficulty uint64
}

// Create a mock chain health checker that doesn't require the actual chain package
type MockChainHealthChecker struct {
	name  string
	chain *MockChain
}

func NewMockChainHealthChecker(chain *MockChain) *MockChainHealthChecker {
	return &MockChainHealthChecker{
		name:  "blockchain",
		chain: chain,
	}
}

func (m *MockChainHealthChecker) Name() string {
	return m.name
}

func (m *MockChainHealthChecker) Check() (*Component, error) {
	start := time.Now()

	// Get current chain state
	height := m.chain.GetHeight()
	bestBlock := m.chain.GetBestBlock()

	if bestBlock == nil {
		return &Component{
			Name:      m.Name(),
			Status:    StatusUnhealthy,
			Message:   "No best block available",
			LastCheck: time.Now(),
			CheckTime: time.Since(start),
			Details: map[string]interface{}{
				"height": height,
				"error":  "best block is nil",
			},
		}, nil
	}

	// Check if the best block hash matches the expected hash
	expectedHash := bestBlock.CalculateHash()
	if expectedHash == nil {
		return &Component{
			Name:      m.Name(),
			Status:    StatusUnhealthy,
			Message:   "Failed to calculate best block hash",
			LastCheck: time.Now(),
			CheckTime: time.Since(start),
			Details: map[string]interface{}{
				"height": height,
				"error":  "hash calculation failed",
			},
		}, nil
	}

	// Check if the chain has reasonable height (not stuck at 0)
	if height == 0 && bestBlock.Header().Height == 0 {
		// This might be normal for a new chain, but let's check if it's the genesis block
		genesisBlock := m.chain.GetGenesisBlock()
		if genesisBlock == nil {
			return &Component{
				Name:      m.Name(),
				Status:    StatusUnhealthy,
				Message:   "No genesis block available",
				LastCheck: time.Now(),
				CheckTime: time.Since(start),
				Details: map[string]interface{}{
					"height": height,
					"error":  "genesis block missing",
				},
			}, nil
		}
	}

	// Check if the last block is recent (within reasonable time)
	now := time.Now()
	blockAge := now.Sub(bestBlock.Header().Timestamp)

	// Consider block unhealthy if it's older than 1 hour (for a 10-second block time chain)
	maxBlockAge := time.Hour
	if blockAge > maxBlockAge {
		return &Component{
			Name:      m.Name(),
			Status:    StatusDegraded,
			Message:   fmt.Sprintf("Last block is %v old", blockAge),
			LastCheck: time.Now(),
			CheckTime: time.Since(start),
			Details: map[string]interface{}{
				"height":          height,
				"last_block_time": bestBlock.Header().Timestamp,
				"block_age":       blockAge.String(),
				"max_block_age":   maxBlockAge.String(),
				"best_block_hash": fmt.Sprintf("%x", expectedHash),
				"difficulty":      bestBlock.Header().Difficulty,
			},
		}, nil
	}

	// Check if difficulty is reasonable (not 0 or extremely high)
	if bestBlock.Header().Difficulty <= 0 {
		return &Component{
			Name:      m.Name(),
			Status:    StatusDegraded,
			Message:   "Block difficulty is zero or negative",
			LastCheck: time.Now(),
			CheckTime: time.Since(start),
			Details: map[string]interface{}{
				"height":          height,
				"last_block_time": bestBlock.Header().Timestamp,
				"block_age":       blockAge.String(),
				"best_block_hash": fmt.Sprintf("%x", expectedHash),
				"difficulty":      bestBlock.Header().Difficulty,
				"warning":         "difficulty should be positive",
			},
		}, nil
	}

	// Chain appears healthy
	return &Component{
		Name:      m.Name(),
		Status:    StatusHealthy,
		Message:   "Blockchain is healthy",
		LastCheck: time.Now(),
		CheckTime: time.Since(start),
		Details: map[string]interface{}{
			"height":          height,
			"last_block_time": bestBlock.Header().Timestamp,
			"block_age":       blockAge.String(),
			"best_block_hash": fmt.Sprintf("%x", expectedHash),
			"difficulty":      bestBlock.Header().Difficulty,
			"transactions":    len(bestBlock.transactions),
		},
	}, nil
}
