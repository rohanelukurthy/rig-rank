package ui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rohanelukurthy/rig-rank/internal/benchmark"
	"github.com/rohanelukurthy/rig-rank/internal/models"
	"github.com/rohanelukurthy/rig-rank/internal/telemetry"
)

// Styles - Premium look
var (
	subtleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	titleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true) // Pinkish
	checkMark   = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓")
	crossMark   = lipgloss.NewStyle().Foreground(lipgloss.Color("160")).Render("✗")
	infoStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("39")) // Blue
)

type ValidationStep int

const (
	StepTelemetry ValidationStep = iota
	StepHealthCheck
	StepBenchmark
	StepDone
)

type Model struct {
	spinner   spinner.Model
	step      ValidationStep
	modelName string
	debug     bool

	// Data
	sysInfo *models.SystemInfo
	client  benchmark.BenchmarkClient
	runner  *benchmark.Runner
	results *models.BenchmarkResult

	// Errors
	err error

	// Pipeline state
	benchmarkProfileIndex int
	benchmarkProfiles     []string
}

func NewModel(modelName string, debug bool) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return Model{
		spinner:           s,
		modelName:         modelName,
		debug:             debug,
		step:              StepTelemetry,
		benchmarkProfiles: []string{"Atomic Check", "Code Generation", "Story Generation", "Summarization", "Reasoning"},
		results: &models.BenchmarkResult{
			MetricsVersion: "1.0",
			ModelMetadata:  models.ModelMetadata{Name: modelName},
		},
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		gatherTelemetryCmd(),
	)
}

// Messages
type telemetryMsg struct {
	info *models.SystemInfo
	err  error
}

type healthCheckMsg struct {
	client benchmark.BenchmarkClient
	err    error
}

type benchmarkProfileMsg struct {
	profileName string
	stats       *models.ProfileStats
	err         error
}

type finishedMsg struct{}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case telemetryMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		m.sysInfo = msg.info
		m.step = StepHealthCheck
		return m, checkHealthCmd()

	case healthCheckMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		m.client = msg.client
		m.runner = benchmark.NewRunner(m.client)
		m.runner.Debug = m.debug
		m.step = StepBenchmark
		return m, startNextProfileCmd(m.runner, m.modelName, m.benchmarkProfileIndex)

	case benchmarkProfileMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		// Store results
		// Mapping back based on index or name is a bit fragile, but efficient for this 1-shot tool.
		switch msg.profileName {
		case "Atomic Check":
			m.results.Benchmarks.Atomic = *msg.stats
		case "Code Generation":
			m.results.Benchmarks.CodeGen = *msg.stats
		case "Story Generation":
			m.results.Benchmarks.StoryGen = *msg.stats
		case "Summarization":
			m.results.Benchmarks.Summarization = *msg.stats
		case "Reasoning":
			m.results.Benchmarks.Reasoning = *msg.stats
		}

		m.benchmarkProfileIndex++
		if m.benchmarkProfileIndex >= len(m.benchmarkProfiles) {
			m.step = StepDone
			return m, tea.Quit
		}
		return m, startNextProfileCmd(m.runner, m.modelName, m.benchmarkProfileIndex)
	}

	return m, cmd
}

func (m Model) FinalOutput() string {
	if m.step != StepDone {
		return ""
	}
	return generateJSON(m.results, m.sysInfo)
}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\n%s Error: %v\n\n", crossMark, m.err)
	}

	s := strings.Builder{}
	s.WriteString(fmt.Sprintf("\n%s\n\n", titleStyle.Render("RigRank Benchmark")))

	// 1. Telemetry
	if m.sysInfo != nil {
		s.WriteString(fmt.Sprintf("%s System Telemetry Clean\n", checkMark))
		s.WriteString(fmt.Sprintf("  %s %s\n", subtleStyle.Render("• CPU:"), m.sysInfo.CPU.Model))
		s.WriteString(fmt.Sprintf("  %s %d MB\n", subtleStyle.Render("• RAM:"), m.sysInfo.RAM.TotalMB))
		if m.sysInfo.GPU.Model != "" {
			s.WriteString(fmt.Sprintf("  %s %s\n", subtleStyle.Render("• GPU:"), m.sysInfo.GPU.Model))
		}
	} else if m.step == StepTelemetry {
		s.WriteString(fmt.Sprintf("%s Gathering Telemetry...\n", m.spinner.View()))
	}

	// 2. Health Check
	if m.client != nil {
		s.WriteString(fmt.Sprintf("%s Ollama Connected\n", checkMark))
	} else if m.step == StepHealthCheck {
		s.WriteString(fmt.Sprintf("%s Connecting to Ollama...\n", m.spinner.View()))
	}

	// 3. Benchmarks
	if m.step == StepBenchmark {
		currentProfile := m.benchmarkProfiles[m.benchmarkProfileIndex]
		s.WriteString(fmt.Sprintf("\n%s Running Suite (%d/%d): %s\n", m.spinner.View(), m.benchmarkProfileIndex+1, len(m.benchmarkProfiles), currentProfile))
	}

	// Show completed profiles
	if m.benchmarkProfileIndex > 0 {
		// Just a simple summary of what's done
		s.WriteString("\n")
		// Determine which are done
		doneCount := m.benchmarkProfileIndex
		if m.step == StepDone {
			doneCount = len(m.benchmarkProfiles)
		}

		for i := 0; i < doneCount; i++ {
			name := m.benchmarkProfiles[i]
			s.WriteString(fmt.Sprintf("%s %s\n", checkMark, name))
		}
	}

	s.WriteString("\n")
	return s.String()
}

// Commands

func gatherTelemetryCmd() tea.Cmd {
	return func() tea.Msg {
		// Add artificial delay for UX so it doesn't flash too fast?
		// time.Sleep(500 * time.Millisecond)
		info, err := telemetry.GetSystemInfo()
		return telemetryMsg{info: info, err: err}
	}
}

func checkHealthCmd() tea.Cmd {
	return func() tea.Msg {
		client := benchmark.NewClient("http://localhost:11434")
		err := client.CheckHealth()
		return healthCheckMsg{client: client, err: err}
	}
}

func startNextProfileCmd(runner *benchmark.Runner, modelName string, index int) tea.Cmd {
	return func() tea.Msg {
		// Define configs map or switch
		var cfg benchmark.ProfileConfig
		// We could hardcode or switch.
		// "Atomic Check", "Code Generation", "Story Generation", "Summarization", "Reasoning"

		// NOTE: This logic duplicates the config in runner.go/RunSuite somewhat.
		//Ideally we'd share this config. For now, we redefine for the UI flow.

		switch index {
		case 0:
			cfg = benchmark.ProfileConfig{
				Input: 32, Output: 16, Name: "Atomic Check", Iterations: 5,
				Prompt: "What is the capital of France? Answer in one word.",
			}
		case 1:
			cfg = benchmark.ProfileConfig{
				Input: 80, Output: 256, Name: "Code Generation", Iterations: 5,
				Prompt: "Write a Python function to find the second largest element in a list.",
			}
		case 2:
			cfg = benchmark.ProfileConfig{
				Input: 50, Output: 400, Name: "Story Generation", Iterations: 5,
				Prompt: "Write a short story about a robot who discovers nature.",
			}
		case 3:
			cfg = benchmark.ProfileConfig{
				Input: 2048, Output: 128, Name: "Summarization", Iterations: 5,
				Prompt: generateDummyText(2048) + " Summarize the above.",
			}
		case 4:
			cfg = benchmark.ProfileConfig{
				Input: 100, Output: 150, Name: "Reasoning", Iterations: 5,
				Prompt: "Solve this math problem step by step: If x=2 and y=3, what is 2x + 3y?",
			}
		}

		stats, err := runner.RunProfile(modelName, cfg)
		return benchmarkProfileMsg{profileName: cfg.Name, stats: stats, err: err}
	}
}

func generateDummyText(tokens int) string {
	base := "The quick brown fox jumps over the lazy dog. "
	needed := tokens * 4
	var b string
	for len(b) < needed {
		b += base
	}
	return b[:needed]
}

func generateJSON(results *models.BenchmarkResult, sysInfo *models.SystemInfo) string {
	combined := map[string]interface{}{
		"system_info":       sysInfo,
		"inference_results": results,
		// Suitability and verdict would be calculated here too,
		// but skipping for now to keep it simple, or migrating logic from main if it existed.
	}
	b, _ := json.MarshalIndent(combined, "", "  ")
	return string(b)
}
