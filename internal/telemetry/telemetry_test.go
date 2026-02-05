package telemetry

import (
	"testing"
)

func TestGetSystemInfo(t *testing.T) {
	info, err := GetSystemInfo()
	if err != nil {
		t.Fatalf("GetSystemInfo() failed: %v", err)
	}

	// Basic validation - exact values depend on the host machine, so we check for non-zero/non-empty values where possible.
	if info.Arch == "" {
		t.Error("Expected Arch to be populated, got empty string")
	}

	// CPU checks
	if info.CPU.CoresLogical <= 0 {
		t.Errorf("Expected CPU.CoresLogical > 0, got %d", info.CPU.CoresLogical)
	}

	// RAM checks
	if info.RAM.TotalMB <= 0 {
		t.Errorf("Expected RAM.TotalMB > 0, got %d", info.RAM.TotalMB)
	}
}
