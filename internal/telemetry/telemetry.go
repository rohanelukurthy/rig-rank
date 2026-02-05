package telemetry

import (
	"runtime"

	"github.com/jaypipes/ghw"
	"github.com/rohanelukurthy/rig-rank/internal/models"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// GetSystemInfo gathers hardware telemetry from the host system.
func GetSystemInfo() (*models.SystemInfo, error) {
	info := &models.SystemInfo{
		Arch: runtime.GOARCH,
	}

	// 1. CPU
	if err := getCPUInfo(&info.CPU); err != nil {
		// Log error but continue? For now, return error.
		return nil, err
	}

	// 2. RAM
	if err := getRAMInfo(&info.RAM); err != nil {
		return nil, err
	}

	// 3. GPU (Using ghw for better cross-platform support)
	if err := getGPUInfo(&info.GPU); err != nil {
		// GPU might not be present or accessible, strict failure might be too harsh.
		// For now we just log/ignore or leave empty.
		// Let's return nil to indicate we tried.
	}

	return info, nil
}

func getCPUInfo(c *models.CPU) error {
	// gopsutil for basic info
	stats, err := cpu.Info()
	if err != nil {
		return err
	}
	if len(stats) > 0 {
		c.Model = stats[0].ModelName
		c.FrequencyMaxMHz = stats[0].Mhz
	}

	// Cores
	physical, err := cpu.Counts(false)
	if err == nil {
		c.CoresPhysical = physical
	}
	logical, err := cpu.Counts(true)
	if err == nil {
		c.CoresLogical = logical
	}
	return nil
}

func getRAMInfo(r *models.RAM) error {
	v, err := mem.VirtualMemory()
	if err != nil {
		return err
	}
	r.TotalMB = v.Total / 1024 / 1024
	r.Type = "Unknown"

	if runtime.GOOS == "darwin" {
		getMacOSRAMInfo(r)
	}

	return nil
}

func getGPUInfo(g *models.GPU) error {
	if runtime.GOOS == "darwin" {
		return getMacOSGPUInfo(g)
	}

	// Fallback to ghw for Linux/Windows
	gpu, err := ghw.GPU()
	if err != nil {
		return err
	}
	for _, card := range gpu.GraphicsCards {
		g.Model = card.DeviceInfo.Product.Name
		break
	}
	return nil
}
