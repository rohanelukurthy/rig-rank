//go:build linux

package telemetry

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/rohanelukurthy/rig-rank/internal/models"
)

// getLinuxNvidiaInfo queries nvidia-smi for VRAM, PCIe gen, and lane count.
func getLinuxNvidiaInfo(g *models.GPU) error {
	out, err := exec.Command("nvidia-smi",
		"--query-gpu=memory.total,pcie.link.gen.current,pcie.link.width.current",
		"--format=csv,noheader,nounits",
	).Output()
	if err != nil {
		return err
	}

	line := strings.TrimSpace(string(out))
	// Handle multiple GPUs: use first line
	if idx := strings.Index(line, "\n"); idx >= 0 {
		line = line[:idx]
	}
	parts := strings.Split(line, ",")
	if len(parts) < 3 {
		return fmt.Errorf("nvidia-smi: unexpected output format")
	}

	// memory.total is in MiB
	if v, err := strconv.Atoi(strings.TrimSpace(parts[0])); err == nil {
		g.VRAMTotalMB = v
	}

	// PCIe gen as integer -> "gen4", "gen3", etc.
	if v, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil && v > 0 {
		g.PCIeGen = fmt.Sprintf("gen%d", v)
	}

	// PCIe lane count
	if v, err := strconv.Atoi(strings.TrimSpace(parts[2])); err == nil {
		g.PCIeLanes = v
	}

	return nil
}
