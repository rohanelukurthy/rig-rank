package telemetry

import (
	"testing"
	"time"
)

func TestWaitForQuietState_Success(t *testing.T) {
	// Lenient config so it should pass
	cfg := QuietStateConfig{
		Timeout:      2 * time.Second,
		WaitDuration: 1 * time.Second,
		CPUThreshold: 100.0, // Any CPU is fine
		RAMMinFreeMB: 0,     // Any RAM is fine
	}

	var msgs []string
	err := WaitForQuietState(cfg, func(msg string) {
		msgs = append(msgs, msg)
	})

	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if len(msgs) == 0 {
		t.Errorf("expected status updates, got none")
	}
}

func TestWaitForQuietState_Timeout(t *testing.T) {
	// Strict config so it should fail
	cfg := QuietStateConfig{
		Timeout:      2 * time.Second,
		WaitDuration: 5 * time.Second,
		CPUThreshold: -1.0,        // Impossible to meet
		RAMMinFreeMB: 1024 * 1024, // 1TB free RAM
	}

	err := WaitForQuietState(cfg, nil)

	if err == nil {
		t.Fatalf("expected timeout error due to noisy constraints, got success")
	}
}
