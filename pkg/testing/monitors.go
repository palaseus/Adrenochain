package testing

import (
	"runtime"
	"sync"
	"time"
)

// PerformanceMonitor monitors overall test performance
type PerformanceMonitor struct {
	mu sync.RWMutex

	// Performance metrics
	startTime     time.Time
	endTime       time.Time
	totalDuration time.Duration
	
	// Test execution metrics
	totalTests        uint64
	completedTests    uint64
	averageTestTime   time.Duration
	longestTestTime   time.Duration
	shortestTestTime  time.Duration
	
	// Resource usage
	peakMemoryUsage   uint64
	averageMemoryUsage uint64
	peakCPUUsage      float64
	averageCPUUsage   float64
	
	// Statistics
	lastUpdate        time.Time
}

// MemoryMonitor monitors memory usage during testing
type MemoryMonitor struct {
	mu sync.RWMutex

	// Memory metrics
	startMemory      uint64
	currentMemory    uint64
	peakMemory       uint64
	averageMemory    uint64
	memorySamples    []uint64
	
	// Memory allocation tracking
	totalAllocations uint64
	totalFrees       uint64
	activeAllocations uint64
	
	// Statistics
	lastUpdate       time.Time
	sampleCount      uint64
}

// CPUMonitor monitors CPU usage during testing
type CPUMonitor struct {
	mu sync.RWMutex

	// CPU metrics
	startCPU         float64
	currentCPU       float64
	peakCPU          float64
	averageCPU       float64
	cpuSamples       []float64
	
	// CPU time tracking
	userTime         time.Duration
	systemTime       time.Duration
	idleTime         time.Duration
	
	// Statistics
	lastUpdate       time.Time
	sampleCount      uint64
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		startTime:        time.Now(),
		totalTests:       0,
		completedTests:   0,
		peakMemoryUsage:  0,
		peakCPUUsage:     0,
		lastUpdate:       time.Now(),
	}
}

// NewMemoryMonitor creates a new memory monitor
func NewMemoryMonitor() *MemoryMonitor {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return &MemoryMonitor{
		startMemory:       m.Alloc,
		currentMemory:     m.Alloc,
		peakMemory:        m.Alloc,
		averageMemory:     m.Alloc,
		memorySamples:     make([]uint64, 0),
		totalAllocations:  m.Mallocs,
		totalFrees:        m.Frees,
		activeAllocations: m.Mallocs - m.Frees,
		lastUpdate:        time.Now(),
		sampleCount:       0,
	}
}

// NewCPUMonitor creates a new CPU monitor
func NewCPUMonitor() *CPUMonitor {
	return &CPUMonitor{
		startCPU:     0,
		currentCPU:   0,
		peakCPU:      0,
		averageCPU:   0,
		cpuSamples:   make([]float64, 0),
		lastUpdate:   time.Now(),
		sampleCount:  0,
	}
}

// Start begins performance monitoring
func (pm *PerformanceMonitor) Start() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	pm.startTime = time.Now()
	pm.lastUpdate = pm.startTime
}

// Stop ends performance monitoring
func (pm *PerformanceMonitor) Stop() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	pm.endTime = time.Now()
	pm.totalDuration = pm.endTime.Sub(pm.startTime)
}

// RecordTestExecution records metrics for a test execution
func (pm *PerformanceMonitor) RecordTestExecution(duration time.Duration, memoryUsage uint64, cpuUsage float64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	pm.completedTests++
	
	// Update test time metrics
	if pm.completedTests == 1 {
		pm.shortestTestTime = duration
		pm.longestTestTime = duration
		pm.averageTestTime = duration
	} else {
		if duration < pm.shortestTestTime {
			pm.shortestTestTime = duration
		}
		if duration > pm.longestTestTime {
			pm.longestTestTime = duration
		}
		
		// Update average
		totalTime := pm.averageTestTime * time.Duration(pm.completedTests-1)
		totalTime += duration
		pm.averageTestTime = totalTime / time.Duration(pm.completedTests)
	}
	
	// Update memory metrics
	if memoryUsage > pm.peakMemoryUsage {
		pm.peakMemoryUsage = memoryUsage
	}
	
	// Update CPU metrics
	if cpuUsage > pm.peakCPUUsage {
		pm.peakCPUUsage = cpuUsage
	}
	
	pm.lastUpdate = time.Now()
}

// GetPerformanceMetrics returns current performance metrics
func (pm *PerformanceMonitor) GetPerformanceMetrics() *PerformanceMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	return &PerformanceMetrics{
		StartTime:        pm.startTime,
		EndTime:          pm.endTime,
		TotalDuration:    pm.totalDuration,
		TotalTests:       pm.totalTests,
		CompletedTests:   pm.completedTests,
		AverageTestTime:  pm.averageTestTime,
		LongestTestTime:  pm.longestTestTime,
		ShortestTestTime: pm.shortestTestTime,
		PeakMemoryUsage:  pm.peakMemoryUsage,
		PeakCPUUsage:     pm.peakCPUUsage,
		LastUpdate:       pm.lastUpdate,
	}
}

// Start begins memory monitoring
func (mm *MemoryMonitor) Start() {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	mm.startMemory = m.Alloc
	mm.currentMemory = m.Alloc
	mm.peakMemory = m.Alloc
	mm.averageMemory = m.Alloc
	mm.totalAllocations = m.Mallocs
	mm.totalFrees = m.Frees
	mm.activeAllocations = m.Mallocs - m.Frees
	mm.lastUpdate = time.Now()
}

// Sample records a memory usage sample
func (mm *MemoryMonitor) Sample() {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	mm.currentMemory = m.Alloc
	mm.totalAllocations = m.Mallocs
	mm.totalFrees = m.Frees
	mm.activeAllocations = m.Mallocs - m.Frees
	
	// Update peak memory
	if mm.currentMemory > mm.peakMemory {
		mm.peakMemory = mm.currentMemory
	}
	
	// Add to samples
	mm.memorySamples = append(mm.memorySamples, mm.currentMemory)
	mm.sampleCount++
	
	// Update average
	totalMemory := mm.averageMemory * uint64(mm.sampleCount-1)
	totalMemory += mm.currentMemory
	mm.averageMemory = totalMemory / uint64(mm.sampleCount)
	
	mm.lastUpdate = time.Now()
}

// GetMemoryMetrics returns current memory metrics
func (mm *MemoryMonitor) GetMemoryMetrics() *MemoryMetrics {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	
	return &MemoryMetrics{
		StartMemory:       mm.startMemory,
		CurrentMemory:     mm.currentMemory,
		PeakMemory:        mm.peakMemory,
		AverageMemory:     mm.averageMemory,
		TotalAllocations:  mm.totalAllocations,
		TotalFrees:        mm.totalFrees,
		ActiveAllocations: mm.activeAllocations,
		SampleCount:       mm.sampleCount,
		LastUpdate:        mm.lastUpdate,
	}
}

// Start begins CPU monitoring
func (cm *CPUMonitor) Start() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	cm.startCPU = 0
	cm.currentCPU = 0
	cm.peakCPU = 0
	cm.averageCPU = 0
	cm.lastUpdate = time.Now()
}

// Sample records a CPU usage sample
func (cm *CPUMonitor) Sample() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	// In a real implementation, this would measure actual CPU usage
	// For now, simulate CPU usage
	cm.currentCPU = float64(time.Now().UnixNano() % 100) / 100.0
	
	// Update peak CPU
	if cm.currentCPU > cm.peakCPU {
		cm.peakCPU = cm.currentCPU
	}
	
	// Add to samples
	cm.cpuSamples = append(cm.cpuSamples, cm.currentCPU)
	cm.sampleCount++
	
	// Update average
	totalCPU := cm.averageCPU * float64(cm.sampleCount-1)
	totalCPU += cm.currentCPU
	cm.averageCPU = totalCPU / float64(cm.sampleCount)
	
	cm.lastUpdate = time.Now()
}

// GetCPUMetrics returns current CPU metrics
func (cm *CPUMonitor) GetCPUMetrics() *CPUMetrics {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	return &CPUMetrics{
		StartCPU:     cm.startCPU,
		CurrentCPU:   cm.currentCPU,
		PeakCPU:      cm.peakCPU,
		AverageCPU:   cm.averageCPU,
		UserTime:     cm.userTime,
		SystemTime:   cm.systemTime,
		IdleTime:     cm.idleTime,
		SampleCount:  cm.sampleCount,
		LastUpdate:   cm.lastUpdate,
	}
}

// PerformanceMetrics contains performance monitoring data
type PerformanceMetrics struct {
	StartTime        time.Time
	EndTime          time.Time
	TotalDuration    time.Duration
	TotalTests       uint64
	CompletedTests   uint64
	AverageTestTime  time.Duration
	LongestTestTime  time.Duration
	ShortestTestTime time.Duration
	PeakMemoryUsage  uint64
	PeakCPUUsage     float64
	LastUpdate       time.Time
}

// MemoryMetrics contains memory monitoring data
type MemoryMetrics struct {
	StartMemory       uint64
	CurrentMemory     uint64
	PeakMemory        uint64
	AverageMemory     uint64
	TotalAllocations  uint64
	TotalFrees        uint64
	ActiveAllocations uint64
	SampleCount       uint64
	LastUpdate        time.Time
}

// CPUMetrics contains CPU monitoring data
type CPUMetrics struct {
	StartCPU    float64
	CurrentCPU  float64
	PeakCPU     float64
	AverageCPU  float64
	UserTime    time.Duration
	SystemTime  time.Duration
	IdleTime    time.Duration
	SampleCount uint64
	LastUpdate  time.Time
}
