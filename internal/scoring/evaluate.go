package scoring

import (
	"fmt"

	"github.com/rohanelukurthy/rig-rank/internal/models"
)

const (
	RatingExcellent = "EXCELLENT"
	RatingGood      = "GOOD"
	RatingPoor      = "POOR"
)

func Evaluate(results *models.BenchmarkResult) *models.SuitabilityReport {
	report := &models.SuitabilityReport{}

	// 1. Quick Q&A (Atomic Check TTFT)
	ttft := results.Benchmarks.Atomic.Stats.TTFTMs.Mean
	if ttft < 50 {
		report.QuickQA = models.Suitability{Rating: RatingExcellent, Reason: fmt.Sprintf("TTFT of %.1fms is very responsive.", ttft)}
	} else if ttft < 200 {
		report.QuickQA = models.Suitability{Rating: RatingGood, Reason: fmt.Sprintf("TTFT of %.1fms is acceptable.", ttft)}
	} else {
		report.QuickQA = models.Suitability{Rating: RatingPoor, Reason: fmt.Sprintf("TTFT of %.1fms is sluggish.", ttft)}
	}

	// 2. Coding (Code Gen TPS)
	codeTPS := results.Benchmarks.CodeGen.Stats.GenTPS.Mean
	if codeTPS > 40 {
		report.Coding = models.Suitability{Rating: RatingExcellent, Reason: fmt.Sprintf("Generation speed of %.1f t/s is fluid.", codeTPS)}
	} else if codeTPS > 20 {
		report.Coding = models.Suitability{Rating: RatingGood, Reason: fmt.Sprintf("Generation speed of %.1f t/s is usable.", codeTPS)}
	} else {
		report.Coding = models.Suitability{Rating: RatingPoor, Reason: fmt.Sprintf("Generation speed of %.1f t/s is too slow.", codeTPS)}
	}

	// 3. Writing (Story Gen TPS)
	storyTPS := results.Benchmarks.StoryGen.Stats.GenTPS.Mean
	if storyTPS > 35 {
		report.Writing = models.Suitability{Rating: RatingExcellent, Reason: fmt.Sprintf("Speed of %.1f t/s is great for drafting.", storyTPS)}
	} else if storyTPS > 15 {
		report.Writing = models.Suitability{Rating: RatingGood, Reason: fmt.Sprintf("Speed of %.1f t/s is okay.", storyTPS)}
	} else {
		report.Writing = models.Suitability{Rating: RatingPoor, Reason: fmt.Sprintf("Speed of %.1f t/s is distracting.", storyTPS)}
	}

	// 4. Summarization (Prompt TPS)
	summTPS := results.Benchmarks.Summarization.Stats.PromptTPS.Mean
	if summTPS > 200 {
		report.Summarization = models.Suitability{Rating: RatingExcellent, Reason: fmt.Sprintf("Ingestion speed of %.1f t/s is fast.", summTPS)}
	} else if summTPS > 100 {
		report.Summarization = models.Suitability{Rating: RatingGood, Reason: fmt.Sprintf("Ingestion speed of %.1f t/s is decent.", summTPS)}
	} else {
		report.Summarization = models.Suitability{Rating: RatingPoor, Reason: fmt.Sprintf("Ingestion speed of %.1f t/s is slow.", summTPS)}
	}

	// 5. Data Analysis (Reasoning Mean TPS - Balance)
	// We'll use Gen TPS as the primary bottleneck for reasoning usually, but maybe average of both?
	// Let's use Gen TPS component since that's the waiting part.
	reasonTPS := results.Benchmarks.Reasoning.Stats.GenTPS.Mean
	if reasonTPS > 50 {
		report.DataAnalysis = models.Suitability{Rating: RatingExcellent, Reason: fmt.Sprintf("Complex gen speed of %.1f t/s is superb.", reasonTPS)}
	} else if reasonTPS > 25 {
		report.DataAnalysis = models.Suitability{Rating: RatingGood, Reason: fmt.Sprintf("Complex gen speed of %.1f t/s is good.", reasonTPS)}
	} else {
		report.DataAnalysis = models.Suitability{Rating: RatingPoor, Reason: fmt.Sprintf("Complex gen speed of %.1f t/s is low.", reasonTPS)}
	}

	// Verdict
	goodCount := 0
	if report.QuickQA.Rating != RatingPoor {
		goodCount++
	}
	if report.Coding.Rating != RatingPoor {
		goodCount++
	}
	if report.Writing.Rating != RatingPoor {
		goodCount++
	}
	if report.Summarization.Rating != RatingPoor {
		goodCount++
	}
	if report.DataAnalysis.Rating != RatingPoor {
		goodCount++
	}

	if goodCount == 5 {
		report.OverallVerdict = "This model performs well on your hardware for all tested use cases."
	} else if goodCount >= 3 {
		report.OverallVerdict = "This model is suitable for most tasks, but may struggle with some heavy workloads."
	} else {
		report.OverallVerdict = "This model may be too heavy for your hardware. Consider a smaller quantization or parameter count."
	}

	return report
}
