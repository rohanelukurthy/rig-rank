package telemetry

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// QuietStateConfig holds the parameters for validating the system is idle.
type QuietStateConfig struct {
	Timeout       time.Duration
	WaitDuration  time.Duration
	CPUThreshold  float64 // Percentage (e.g., 15.0)
	RAMMinFreeMB  uint64  // Minimum free RAM required in MB
}

// WaitForQuietState monitors the system over a specified duration to ensure
// CPU and RAM usage are below thresholds, indicating a "quiet" system suitable for benchmarking.
// The callback can be used to update a UI with progress or status messages.
func WaitForQuietState(cfg QuietStateConfig, statusCallback func(string)) error {
	deadline := time.Now().Add(cfg.Timeout)
	pollInterval := 1 * time.Second

	var currentQuietDuration time.Duration

	for time.Now().Before(deadline) {
		// 1. Check CPU
		cpuPercents, err := cpu.Percent(pollInterval, false)
		if err != nil {
			return fmt.Errorf("failed to read CPU usage: %w", err)
		}

		cpuUsage := 0.0
		if len(cpuPercents) > 0 {
			cpuUsage = cpuPercents[0]
		}

		// 2. Check RAM
		v, err := mem.VirtualMemory()
		if err != nil {
			return fmt.Errorf("failed to read memory usage: %w", err)
		}
		
		freeRAMMB := v.Available / 1024 / 1024

		noisy := false
		var noisyReason string

		if cpuUsage > cfg.CPUThreshold {
			noisy = true
			noisyReason = fmt.Sprintf("CPU usage (%.1f%%) > %.1f%%", cpuUsage, cfg.CPUThreshold)
		} else if freeRAMMB < cfg.RAMMinFreeMB {
			noisy = true
			noisyReason = fmt.Sprintf("Free RAM (%d MB) < %d MB", freeRAMMB, cfg.RAMMinFreeMB)
		}

		if noisy {
			currentQuietDuration = 0 // Reset contiguous quiet time
			if statusCallback != nil {
				statusCallback(fmt.Sprintf("Waiting for quiet state... Noisy: %s", noisyReason))
			}
		} else {
			currentQuietDuration += pollInterval
			if statusCallback != nil {
				statusCallback(fmt.Sprintf("Monitoring quiet state... %v / %v", currentQuietDuration, cfg.WaitDuration))
			}

			if currentQuietDuration >= cfg.WaitDuration {
				return nil // Success! System is sufficiently quiet.
			}
		}
	}

	return fmt.Errorf("system did not reach a quiet state within %v (CPU Threshold: %.1f%%, Min Free RAM: %d MB)", cfg.Timeout, cfg.CPUThreshold, cfg.RAMMinFreeMB)
}
