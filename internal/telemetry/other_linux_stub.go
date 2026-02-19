//go:build !darwin && !linux

package telemetry

import "github.com/rohanelukurthy/rig-rank/internal/models"

func getLinuxNvidiaInfo(g *models.GPU) error {
	// No-op on non-Linux platforms (Windows, FreeBSD, etc.)
	return nil
}
