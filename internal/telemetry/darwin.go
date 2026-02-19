//go:build darwin

package telemetry

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/rohanelukurthy/rig-rank/internal/models"
)

func getMacOSRAMInfo(r *models.RAM) {
	// system_profiler SPHardwareDataType | grep "Memory:"
	out, err := exec.Command("system_profiler", "SPHardwareDataType").Output()
	if err == nil {
		output := string(out)
		// Try to find Type and Speed
		// Example: "Memory: 32 GB LPDDR5" or separate lines
		// On Apple Silicon it often just says "Memory: 16 GB" and type is implied unified.

		if strings.Contains(output, "LPDDR") {
			r.Type = "LPDDR"
		} else if strings.Contains(output, "DDR5") {
			r.Type = "DDR5"
		} else if strings.Contains(output, "DDR4") {
			r.Type = "DDR4"
		}

		// Speed is harder on simple output, often need SPMemoryDataType
	}

	// Try Memory Data Type for details
	outMem, err := exec.Command("system_profiler", "SPMemoryDataType").Output()
	if err == nil {
		output := string(outMem)
		// look for "Speed: 6400 MHz"
		reSpeed := regexp.MustCompile(`Speed: (\d+) MHz`)
		matches := reSpeed.FindStringSubmatch(output)
		if len(matches) > 1 {
			speed, _ := strconv.Atoi(matches[1])
			r.SpeedMts = speed
		}

		// Type might appear here too
		if r.Type == "Unknown" {
			reType := regexp.MustCompile(`Type: (.+)`)
			typeMatches := reType.FindStringSubmatch(output)
			if len(typeMatches) > 1 {
				r.Type = strings.TrimSpace(typeMatches[1])
			}
		}
	}
}

func getMacOSGPUInfo(g *models.GPU) error {
	out, err := exec.Command("system_profiler", "SPDisplaysDataType").Output()
	if err != nil {
		return err
	}
	output := string(out)

	// Chipset Model: Apple M2
	reModel := regexp.MustCompile(`Chipset Model: (.+)`)
	modelMatch := reModel.FindStringSubmatch(output)
	if len(modelMatch) > 1 {
		g.Model = strings.TrimSpace(modelMatch[1])
	}

	// This is shared memory for Apple Silicon, usually matches RAM total or close to it
	// but system_profiler might not show "VRAM" explicitly for unified memory in the same way.
	// However, usually "Total Number of Cores" or similar helps identify performance.
	// For VRAM on Apple Silicon, it's unified. We can default to total RAM if we detect "Apple M"

	if strings.Contains(g.Model, "Apple M") {
		// Unified memory architecture
		// We can't easily get "allocated" VRAM, but total system mem is the limit.
		// Let's leave VRAM as 0 or set to Total RAM?
		// Plan says "VRAM Total MB"
		// Let's rely on what we grabbed for RAM.
		// Or try to find "VRAM (Total):" line if it exists (for external GPUs on Intel Macs)
	}

	return nil
}

func getLinuxNvidiaInfo(g *models.GPU) error {
	// No-op on darwin (not Linux)
	return nil
}
