package benchmark

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_CheckHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			t.Errorf("Expected path /, got %s", r.URL.Path)
		}
		if r.Method != http.MethodHead {
			t.Errorf("Expected method HEAD, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	if err := client.CheckHealth(); err != nil {
		t.Errorf("CheckHealth() failed: %v", err)
	}
}

func TestClient_Generate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			t.Errorf("Expected path /api/generate, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("Expected method POST, got %s", r.Method)
		}

		// Mock response
		w.Header().Set("Content-Type", "application/json")
		resp := GenerateResponse{
			Model:              "llama3",
			CreatedAt:          time.Now(),
			Response:           "Paris",
			Done:               true,
			TotalDuration:      100 * time.Millisecond,
			LoadDuration:       10 * time.Millisecond,
			PromptEvalCount:    32,
			PromptEvalDuration: 40 * time.Millisecond,
			EvalCount:          16,
			EvalDuration:       50 * time.Millisecond,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	req := GenerateRequest{
		Model:  "llama3",
		Prompt: "Capital of France?",
		Stream: false,
		Options: map[string]interface{}{
			"num_predict": 16,
		},
	}

	stats, err := client.Generate(req)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	// Validate mappings to models.StatsMetric (simplified for single run, actually Generate returns raw data usually,
	// but let's assume client returns a struct we can interpret or raw response)
	// For this test, let's say Generate returns the parsed GenerateResponse or similar struct.
	// Since we are decoupling, let's return the raw response wrapper for now effectively.

	if stats.PromptEvalCount != 32 {
		t.Errorf("Expected 32 prompt tokens, got %d", stats.PromptEvalCount)
	}
}
