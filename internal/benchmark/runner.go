package benchmark

import (
	"fmt"
	"math"
	"sort"

	"github.com/rohanelukurthy/rig-rank/internal/models"
)

// BenchmarkClient Interface to allow mocking
type BenchmarkClient interface {
	Generate(req GenerateRequest) (*GenerateResponse, error)
	CheckHealth() error
}

// Runner executes the benchmark suite.
type Runner struct {
	client        BenchmarkClient
	Debug         bool
	ContextWindow int
}

// NewRunner creates a new benchmark runner.
func NewRunner(client BenchmarkClient, contextWindow int) *Runner {
	return &Runner{client: client, Debug: false, ContextWindow: contextWindow}
}

// RunSuite executes all 5 test profiles.
func (r *Runner) RunSuite(modelName string) (*models.BenchmarkResult, error) {
	result := &models.BenchmarkResult{
		MetricsVersion: "1.0",
		ModelMetadata: models.ModelMetadata{
			Name: modelName,
		},
	}

	// Collect all load durations across all iterations
	var allLoadDurations []float64

	// 1. Atomic
	if r.Debug {
		fmt.Println("[DEBUG] Starting Atomic Check...")
	}
	atomicStats, loadDurs, err := r.RunProfile(modelName, ProfileConfig{
		Input: 32, Output: 16, Name: "Atomic Check", Iterations: 5,
		Prompt: "What is the capital of France? Answer in one word.",
	})
	if err != nil {
		return nil, fmt.Errorf("atomic profile failed: %w", err)
	}
	result.Benchmarks.Atomic = *atomicStats
	allLoadDurations = append(allLoadDurations, loadDurs...)

	// 2. Code Gen
	if r.Debug {
		fmt.Println("[DEBUG] Starting Code Generation...")
	}
	codeStats, loadDurs, err := r.RunProfile(modelName, ProfileConfig{
		Input: 80, Output: 256, Name: "Code Generation", Iterations: 5,
		Prompt: "Write a Python function to find the second largest element in a list.",
	})
	if err != nil {
		return nil, err
	}
	result.Benchmarks.CodeGen = *codeStats
	allLoadDurations = append(allLoadDurations, loadDurs...)

	// 3. Story Gen
	if r.Debug {
		fmt.Println("[DEBUG] Starting Story Generation...")
	}
	storyStats, loadDurs, err := r.RunProfile(modelName, ProfileConfig{
		Input: 50, Output: 400, Name: "Story Generation", Iterations: 5,
		Prompt: "Write a short story about a robot who discovers nature.",
	})
	if err != nil {
		return nil, err
	}
	result.Benchmarks.StoryGen = *storyStats
	allLoadDurations = append(allLoadDurations, loadDurs...)

	// 4. Summarization
	if r.Debug {
		fmt.Println("[DEBUG] Starting Summarization...")
	}
	summStats, loadDurs, err := r.RunProfile(modelName, ProfileConfig{
		Input: 2048, Output: 128, Name: "Summarization", Iterations: 5,
		Prompt: generateDummyText(2048) + " Summarize the above.",
	})
	if err != nil {
		return nil, err
	}
	result.Benchmarks.Summarization = *summStats
	allLoadDurations = append(allLoadDurations, loadDurs...)

	// 5. Reasoning
	if r.Debug {
		fmt.Println("[DEBUG] Starting Reasoning...")
	}
	reasonStats, loadDurs, err := r.RunProfile(modelName, ProfileConfig{
		Input: 100, Output: 150, Name: "Reasoning", Iterations: 5,
		Prompt: "Solve this math problem step by step: If x=2 and y=3, what is 2x + 3y?",
	})
	if err != nil {
		return nil, err
	}
	result.Benchmarks.Reasoning = *reasonStats
	allLoadDurations = append(allLoadDurations, loadDurs...)

	// Analyze load durations:
	// - InitialLoadMs: First iteration load duration (potential cold start)
	// - SteadyStateLoadMs: Mean of all subsequent iterations
	if len(allLoadDurations) > 0 {
		result.InitialLoadMs = allLoadDurations[0] // First iteration is closest to cold start

		if len(allLoadDurations) > 1 {
			var sum float64
			for _, d := range allLoadDurations[1:] {
				sum += d
			}
			result.SteadyStateLoadMs = sum / float64(len(allLoadDurations)-1)
		}
	}

	return result, nil
}

type ProfileConfig struct {
	Name       string
	Input      int
	Output     int
	Iterations int
	Prompt     string
}

func (r *Runner) RunProfile(model string, cfg ProfileConfig) (*models.ProfileStats, []float64, error) {
	// Skip warmup here - cold start is captured at suite level
	// Each profile just runs iterations with model already warm

	var ttfts []float64
	var genTPS []float64
	var promptTPS []float64
	var loadDurations []float64

	for i := 0; i < cfg.Iterations; i++ {
		if r.Debug {
			fmt.Printf("[DEBUG] Iteration %d/%d\n", i+1, cfg.Iterations)
		}
		resp, err := r.client.Generate(GenerateRequest{
			Model: model, Prompt: cfg.Prompt, Stream: false,
			Options: map[string]interface{}{
				"num_predict": cfg.Output,
				"num_ctx":     r.ContextWindow,
				"temperature": 0.0,
			},
		})
		if err != nil {
			return nil, nil, err
		}

		// Calculate metrics
		// TTFT: total - eval - prompt_eval (approx)
		// Usually API gives them. Let's trust values.
		ttft := float64(resp.TotalDuration.Milliseconds()) - float64(resp.EvalDuration.Milliseconds()) - float64(resp.PromptEvalDuration.Milliseconds())
		if ttft < 0 {
			ttft = 0
		} // sanity

		ttfts = append(ttfts, ttft)
		loadDurations = append(loadDurations, float64(resp.LoadDuration.Milliseconds()))

		if resp.EvalDuration > 0 {
			genTPS = append(genTPS, float64(resp.EvalCount)/resp.EvalDuration.Seconds())
		}
		if resp.PromptEvalDuration > 0 {
			promptTPS = append(promptTPS, float64(resp.PromptEvalCount)/resp.PromptEvalDuration.Seconds())
		}
	}

	return &models.ProfileStats{
		Description: cfg.Name,
		Config:      models.Config{InputTokens: cfg.Input, OutputTokens: cfg.Output},
		Stats: models.Stats{
			TTFTMs:         calculateStats(ttfts),
			GenTPS:         calculateStats(genTPS),
			PromptTPS:      calculateStats(promptTPS),
			LoadDurationMs: calculateStats(loadDurations),
		},
	}, loadDurations, nil
}

func calculateStats(values []float64) *models.StatsMetric {
	if len(values) == 0 {
		return nil
	}
	sort.Float64s(values)

	var sum float64
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))
	median := values[len(values)/2]
	p99Index := int(math.Ceil(float64(len(values))*0.99)) - 1
	if p99Index < 0 {
		p99Index = 0
	}
	p99 := values[p99Index]

	return &models.StatsMetric{
		Mean:   mean,
		Median: median,
		P99:    p99,
	}
}

func generateDummyText(tokens int) string {
	// Crude approximation: 1 token ~= 4 chars (english)
	// We just replicate a string.
	base := "The quick brown fox jumps over the lazy dog. "
	needed := tokens * 4
	var b string
	for len(b) < needed {
		b += base
	}
	return b[:needed]
}
