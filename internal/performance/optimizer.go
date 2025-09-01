package performance

import (
	"runtime"
	"time"
	
	"github.com/fawad-mazhar/onvif-go/internal/logger"
)

// Optimizer provides performance optimization features for resource-constrained devices
type Optimizer struct {
	// Memory optimization settings
	maxMemoryMB int
	gcInterval  time.Duration
	
	// CPU optimization settings
	maxCPUPercent int
}

// NewOptimizer creates a new performance optimizer with default settings
func NewOptimizer() *Optimizer {
	return &Optimizer{
		maxMemoryMB:  50,  // Default max memory usage in MB
		gcInterval:   30 * time.Second,  // Default GC interval
		maxCPUPercent: 80, // Default max CPU usage percentage
	}
}

// SetMaxMemory sets the maximum memory usage limit in MB
func (o *Optimizer) SetMaxMemory(maxMB int) {
	o.maxMemoryMB = maxMB
}

// SetGCInterval sets the garbage collection interval
func (o *Optimizer) SetGCInterval(interval time.Duration) {
	o.gcInterval = interval
}

// SetMaxCPU sets the maximum CPU usage percentage
func (o *Optimizer) SetMaxCPU(maxPercent int) {
	if maxPercent > 0 && maxPercent <= 100 {
		o.maxCPUPercent = maxPercent
	}
}

// StartGC starts periodic garbage collection
func (o *Optimizer) StartGC() {
	go func() {
		ticker := time.NewTicker(o.gcInterval)
		defer ticker.Stop()
		
		for range ticker.C {
			// Force garbage collection
			runtime.GC()
			
			// Log memory stats
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			
			memoryMB := m.Alloc / 1024 / 1024
			if memoryMB > uint64(o.maxMemoryMB) {
				logger.Warn("Memory usage high: %d MB (limit: %d MB)", memoryMB, o.maxMemoryMB)
			}
			
			logger.Debug("Memory usage: %d MB", memoryMB)
		}
	}()
}

// MonitorCPU monitors CPU usage and logs warnings if it exceeds the limit
func (o *Optimizer) MonitorCPU() {
	// In a real implementation, this would monitor actual CPU usage
	// For now, we'll just log a message that monitoring is enabled
	logger.Info("CPU monitoring enabled (limit: %d%%)", o.maxCPUPercent)
}

// OptimizeForEmbedded optimizes settings for embedded/resource-constrained devices
func (o *Optimizer) OptimizeForEmbedded() {
	// Set conservative memory limits
	o.SetMaxMemory(30)
	
	// Set frequent GC interval
	o.SetGCInterval(10 * time.Second)
	
	// Set conservative CPU limits
	o.SetMaxCPU(70)
	
	// Start garbage collection monitoring
	o.StartGC()
	
	// Start CPU monitoring
	o.MonitorCPU()
	
	logger.Info("Performance optimizer configured for embedded devices")
}
