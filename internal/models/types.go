package models

// SystemInfo holds the hardware telemetry data.
type SystemInfo struct {
	Arch string `json:"arch"`
	CPU  CPU    `json:"cpu"`
	GPU  GPU    `json:"gpu"`
	RAM  RAM    `json:"ram"`
}

type CPU struct {
	Model           string  `json:"model"`
	CoresPhysical   int     `json:"cores_physical"`
	CoresLogical    int     `json:"cores_logical"`
	FrequencyMaxMHz float64 `json:"frequency_max_mhz"`
}

type GPU struct {
	Model       string `json:"model"`
	VRAMTotalMB int    `json:"vram_total_mb"`
	PCIeGen     string `json:"pcie_gen"`   // "gen3", "gen4", etc. or empty if unknown
	PCIeLanes   int    `json:"pcie_lanes"` // 0 if unknown
}

type RAM struct {
	TotalMB  uint64 `json:"total_mb"`
	Type     string `json:"type"`      // "DDR4", "LPDDR5", etc.
	SpeedMts int    `json:"speed_mts"` // MT/s
}

// BenchmarkResult holds the results of the inference tests.
type BenchmarkResult struct {
	MetricsVersion    string        `json:"metrics_version"`
	ModelMetadata     ModelMetadata `json:"model_metadata"`
	InitialLoadMs     float64       `json:"initial_load_ms"`      // Load duration of first benchmark iteration
	SteadyStateLoadMs float64       `json:"steady_state_load_ms"` // Mean load duration of subsequent iterations
	Benchmarks        Benchmarks    `json:"benchmarks"`
}

type ModelMetadata struct {
	Name         string `json:"name"`
	Quantization string `json:"quantization"`
	SizeMB       int    `json:"size_mb"`
}

type Benchmarks struct {
	Atomic        ProfileStats `json:"atomic"`
	CodeGen       ProfileStats `json:"code_gen"`
	StoryGen      ProfileStats `json:"story_gen"`
	Summarization ProfileStats `json:"summarization"`
	Reasoning     ProfileStats `json:"reasoning"`
}

type ProfileStats struct {
	Description string `json:"description"`
	Config      Config `json:"config"`
	Stats       Stats  `json:"stats"`
}

type Config struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type Stats struct {
	TTFTMs          *StatsMetric `json:"ttft_ms,omitempty"`
	GenTPS          *StatsMetric `json:"gen_tps,omitempty"`
	PromptTPS       *StatsMetric `json:"prompt_tps,omitempty"`
	TotalDurationMs *StatsMetric `json:"total_duration_ms,omitempty"`
	LoadDurationMs  *StatsMetric `json:"load_duration_ms,omitempty"`
}

type StatsMetric struct {
	Mean   float64 `json:"mean"`
	Median float64 `json:"median"`
	P99    float64 `json:"p99"`
}

// SuitabilityReport holds the analyzed ratings for each use case.
type SuitabilityReport struct {
	QuickQA        Suitability `json:"quick_qa"`
	Coding         Suitability `json:"coding"`
	Writing        Suitability `json:"writing"`
	Summarization  Suitability `json:"summarization"`
	DataAnalysis   Suitability `json:"data_analysis"`
	OverallVerdict string      `json:"overall_verdict"`
}

type Suitability struct {
	Rating string `json:"rating"` // EXCELLENT, GOOD, MARGINAL, POOR
	Reason string `json:"reason"`
}

// FullReport is the top-level structure for the JSON output.
type FullReport struct {
	SystemInfo         *SystemInfo        `json:"system_info"`
	InferenceResults   *BenchmarkResult   `json:"inference_results"`
	UseCaseSuitability *SuitabilityReport `json:"use_case_suitability"`
}
