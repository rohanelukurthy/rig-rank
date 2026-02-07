package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rohanelukurthy/rig-rank/internal/models"
)

// --- Colors ---
var (
	colorExcellent = lipgloss.Color("42")  // Green
	colorGood      = lipgloss.Color("220") // Yellow
	colorPoor      = lipgloss.Color("160") // Red
	colorTitle     = lipgloss.Color("205") // Pink
	colorInfo      = lipgloss.Color("241") // Grey
	colorBorder    = lipgloss.Color("240") // Light grey for borders
)

// tpsToWords converts tokens/sec to approximate words/sec (tokens are ~1.3 per word on average)
func tpsToWords(tps float64) string {
	words := tps / 1.3
	if words >= 1000 {
		return fmt.Sprintf("~%.1fk", words/1000)
	}
	return fmt.Sprintf("~%.0f", words)
}

// formatMs formats milliseconds for display
func formatMs(ms float64) string {
	return fmt.Sprintf("%.0fms", ms)
}

// RenderReportCard renders the holistic report card table
func RenderReportCard(report *models.SuitabilityReport, result *models.BenchmarkResult, modelName string) string {
	s := strings.Builder{}

	// Title
	title := lipgloss.NewStyle().Foreground(colorTitle).Bold(true).Render(fmt.Sprintf("📊 Model Report Card: %s", modelName))
	s.WriteString("\n  " + title + "\n\n")

	// Table styles
	borderStyle := lipgloss.NewStyle().Foreground(colorBorder)
	headerStyle := lipgloss.NewStyle().Foreground(colorInfo)

	// Model Load Summary (Initial vs Steady State)
	initialLabel := lipgloss.NewStyle().Foreground(colorInfo).Render("Initial Load (1st request):")
	steadyLabel := lipgloss.NewStyle().Foreground(colorInfo).Render("Steady State (avg):")
	initialValue := formatMs(result.InitialLoadMs)
	steadyValue := formatMs(result.SteadyStateLoadMs)

	if result.InitialLoadMs > result.SteadyStateLoadMs*3 && result.SteadyStateLoadMs > 0 {
		// Initial load is significantly higher - possible cold start or VRAM constraints
		s.WriteString(lipgloss.NewStyle().Foreground(colorGood).Render("  ⚠️  Model Load: Initial request was slower (possible cold start).") + "\n")
		s.WriteString(fmt.Sprintf("     %s %s  |  %s %s\n", initialLabel, initialValue, steadyLabel, steadyValue))
		s.WriteString(lipgloss.NewStyle().Foreground(colorInfo).Render("     (This is normal if the model wasn't recently used)") + "\n\n")
	} else {
		s.WriteString(lipgloss.NewStyle().Foreground(colorExcellent).Render("  ✅ Model Load: Model was warm (already loaded).") + "\n")
		s.WriteString(fmt.Sprintf("     %s %s  |  %s %s\n\n", initialLabel, initialValue, steadyLabel, steadyValue))
	}

	// Table header
	topBorder := borderStyle.Render("  ┌────────────────────────────────────────────────────────────────────┐")
	s.WriteString(topBorder + "\n")

	header := fmt.Sprintf("  │  %-15s %-12s %-16s %-18s │", "Benchmark", "Startup", "Writing Speed", "Reading Speed")
	s.WriteString(headerStyle.Render(header) + "\n")

	subHeader := fmt.Sprintf("  │  %-15s %-12s %-16s %-18s │", "", "(first word)", "(output)", "(input)")
	s.WriteString(headerStyle.Render(subHeader) + "\n")

	midBorder := borderStyle.Render("  ├────────────────────────────────────────────────────────────────────┤")
	s.WriteString(midBorder + "\n")

	// Table rows
	renderRow := func(name string, stats *models.Stats) string {
		startup := formatMs(stats.TTFTMs.Mean)
		writeSpeed := tpsToWords(stats.GenTPS.Mean) + " words/sec"
		readSpeed := tpsToWords(stats.PromptTPS.Mean) + " words/sec"
		return fmt.Sprintf("  │  %-15s %-12s %-16s %-18s │", name, startup, writeSpeed, readSpeed)
	}

	s.WriteString(renderRow("Atomic Check", &result.Benchmarks.Atomic.Stats) + "\n")
	s.WriteString(renderRow("Code Gen", &result.Benchmarks.CodeGen.Stats) + "\n")
	s.WriteString(renderRow("Story Gen", &result.Benchmarks.StoryGen.Stats) + "\n")
	s.WriteString(renderRow("Summarization", &result.Benchmarks.Summarization.Stats) + "\n")
	s.WriteString(renderRow("Reasoning", &result.Benchmarks.Reasoning.Stats) + "\n")

	bottomBorder := borderStyle.Render("  └────────────────────────────────────────────────────────────────────┘")
	s.WriteString(bottomBorder + "\n\n")

	// Summary insights
	writingRating := report.Coding.Rating
	startupRating := report.QuickQA.Rating

	if writingRating == "EXCELLENT" || writingRating == "GOOD" {
		s.WriteString(lipgloss.NewStyle().Foreground(colorExcellent).Render("  ✅ Writing Speed: Excellent across all tasks.") + "\n")
	} else {
		s.WriteString(lipgloss.NewStyle().Foreground(colorPoor).Render("  ⚠️  Writing Speed: May feel slow for long outputs.") + "\n")
	}

	if startupRating == "POOR" {
		s.WriteString(lipgloss.NewStyle().Foreground(colorGood).Render("  ⚠️  Startup: Noticeable pause before responses begin.") + "\n")
	} else {
		s.WriteString(lipgloss.NewStyle().Foreground(colorExcellent).Render("  ✅ Startup: Responses begin quickly.") + "\n")
	}

	// Overall verdict
	s.WriteString("\n  " + lipgloss.NewStyle().Italic(true).Render(report.OverallVerdict) + "\n")

	return s.String()
}

// RenderChart is kept for backwards compatibility but now calls RenderReportCard
func RenderChart(report *models.SuitabilityReport, result *models.BenchmarkResult) string {
	return RenderReportCard(report, result, result.ModelMetadata.Name)
}
