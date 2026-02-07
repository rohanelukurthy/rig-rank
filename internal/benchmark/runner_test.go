package benchmark

import (
	"testing"
	"time"
)

// MockClient satisfies the interaction needed by Runner (we might need an interface later, but for now we can wrap or mock)
// To keep it simple without large refactor, let's allow injecting a mock function or interface.
// Refactoring Client to Interface would be cleaner TDD.

type MockBenchmarkClient struct {
	GenerateFunc func(req GenerateRequest) (*GenerateResponse, error)
}

func (m *MockBenchmarkClient) Generate(req GenerateRequest) (*GenerateResponse, error) {
	return m.GenerateFunc(req)
}

func (m *MockBenchmarkClient) CheckHealth() error {
	return nil
}

func TestRunSuite(t *testing.T) {
	// Mock client that returns success
	mockClient := &MockBenchmarkClient{
		GenerateFunc: func(req GenerateRequest) (*GenerateResponse, error) {
			return &GenerateResponse{
				TotalDuration:      150 * time.Millisecond,
				PromptEvalDuration: 50 * time.Millisecond,
				EvalDuration:       50 * time.Millisecond,
				LoadDuration:       50 * time.Millisecond,
				PromptEvalCount:    10,
				EvalCount:          10,
			}, nil
		},
	}

	// This is slightly tricky because RunSuite relies on the concrete Client struct currently.
	// We need to refactor Runner to accept an interface.

	// Let's create the Runner
	runner := NewRunner(mockClient, 4096)

	// Run the suite
	results, err := runner.RunSuite("llama3")
	if err != nil {
		t.Fatalf("RunSuite failed: %v", err)
	}

	// Check if we got results for all 5 profiles
	if results.Benchmarks.Atomic.Stats.TTFTMs == nil {
		t.Error("Atomic profile missing stats")
	}
	if results.InitialLoadMs != 50 {
		t.Errorf("Expected InitialLoadMs to be 50ms, got %v", results.InitialLoadMs)
	}
	if results.SteadyStateLoadMs != 50 {
		t.Errorf("Expected SteadyStateLoadMs to be 50ms, got %v", results.SteadyStateLoadMs)
	}
	if results.Benchmarks.Atomic.Stats.LoadDurationMs.Mean != 50 {
		t.Errorf("Expected Atomic Load Duration Mean to be 50ms, got %v", results.Benchmarks.Atomic.Stats.LoadDurationMs.Mean)
	}
	// ... validation for others
}
