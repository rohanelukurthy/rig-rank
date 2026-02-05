package benchmark

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

// Client handles interaction with the Ollama API.
type Client struct {
	baseURL string
	http    *resty.Client
}

// Ensure Client satisfies BenchmarkClient interface (will be defined in runner.go or common local)
// For now implicit satisfaction is enough for Go.

// NewClient creates a new Ollama client.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    resty.New().SetBaseURL(baseURL).SetTimeout(5 * time.Minute),
	}
}

// CheckHealth verifies Ollama is reachable.
func CheckHealthImpl(c *Client) error {
	resp, err := c.http.R().Head("/")
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}
	return nil
}

// GenerateRequest matches Ollama /api/generate payload.
type GenerateRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// GenerateResponse matches Ollama /api/generate response (non-streamed).
type GenerateResponse struct {
	Model              string        `json:"model"`
	CreatedAt          time.Time     `json:"created_at"`
	Response           string        `json:"response"`
	Done               bool          `json:"done"`
	TotalDuration      time.Duration `json:"total_duration"`
	LoadDuration       time.Duration `json:"load_duration"`
	PromptEvalCount    int           `json:"prompt_eval_count"`
	PromptEvalDuration time.Duration `json:"prompt_eval_duration"`
	EvalCount          int           `json:"eval_count"`
	EvalDuration       time.Duration `json:"eval_duration"`
}

// CheckHealth verifies Ollama is running.
func (c *Client) CheckHealth() error {
	resp, err := c.http.R().Head("/")
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("ollama returned status %d", resp.StatusCode())
	}
	return nil
}

// Generate sends an inference request.
func (c *Client) Generate(req GenerateRequest) (*GenerateResponse, error) {
	var result GenerateResponse
	resp, err := c.http.R().
		SetBody(req).
		SetResult(&result).
		Post("/api/generate")

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("generate api error: %s", resp.String())
	}

	return &result, nil
}
