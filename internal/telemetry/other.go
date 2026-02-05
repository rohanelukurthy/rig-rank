//go:build !darwin

package telemetry

import "github.com/rohanelukurthy/rig-rank/internal/models"

func getMacOSRAMInfo(r *models.RAM) {
	// No-op on non-macOS
}

func getMacOSGPUInfo(g *models.GPU) error {
	// No-op on non-macOS
	return nil
}
