package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rohanelukurthy/rig-rank/internal/telemetry"
	"github.com/rohanelukurthy/rig-rank/internal/ui"
	"github.com/spf13/cobra"
)

type runOptions struct {
	model         string
	debug         bool
	output        string
	contextWindow int
	quietWait     bool
	quietCPU      float64
	quietRAMMB    uint64
	quietTimeout  int
	quietWaitSecs int
}

func newRunCmd() *cobra.Command {
	opts := runOptions{}

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Execute the standard benchmark suite",
		Run: func(cmd *cobra.Command, args []string) {
			runBenchmark(opts)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.model, "model", "m", "llama3", "Ollama model name to benchmark")
	flags.BoolVarP(&opts.debug, "debug", "d", false, "Enable verbose debug logging")
	flags.StringVarP(&opts.output, "output", "o", "", "Path to save JSON results")
	flags.IntVarP(&opts.contextWindow, "context-window", "c", 4096, "Context window size for the model")

	flags.BoolVar(&opts.quietWait, "quiet-wait", false, "Wait for system to become idle before benchmarking")
	flags.Float64Var(&opts.quietCPU, "quiet-cpu", 15.0, "Maximum CPU usage percentage allowed during quiet wait")
	flags.Uint64Var(&opts.quietRAMMB, "quiet-ram-mb", 2048, "Minimum free RAM (MB) required during quiet wait")
	flags.IntVar(&opts.quietTimeout, "quiet-timeout", 60, "Timeout in seconds to wait for quiet state")
	flags.IntVar(&opts.quietWaitSecs, "quiet-wait-secs", 5, "Duration in seconds of sustained quiet state required")

	return cmd
}

func runBenchmark(opts runOptions) {
	// 3. Output

	quietCfg := telemetry.QuietStateConfig{
		Timeout:      time.Duration(opts.quietTimeout) * time.Second,
		WaitDuration: time.Duration(opts.quietWaitSecs) * time.Second,
		CPUThreshold: opts.quietCPU,
		RAMMinFreeMB: opts.quietRAMMB,
	}

	p := tea.NewProgram(ui.NewModel(opts.model, opts.debug, opts.output, opts.contextWindow, opts.quietWait, quietCfg), tea.WithOutput(os.Stderr))
	m, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Alas, there's been an error: %v\n", err)
		os.Exit(1)
	}

	if finalModel, ok := m.(ui.Model); ok {
		view, jsonBytes := finalModel.FinalOutput()

		// 1. Print visual view to Stderr (so it's separate from data)
		// Actually, if we are in interactive mode, View() loop already printed it?
		// No, View() prints the "Checking..." states.
		// We want the final table to stay.
		// Tea clears the screen usually? No, "Altscreen" does. We are not using Altscreen.
		// So the final view from the tea loop should remain if we return it in View() for StepDone?
		// My Model.View() handles StepDone?
		// Let's check model.go...
		// "if m.step == StepDone ... return tea.Quit" in Update.
		// View() just shows "✓ Checked" list?
		// Actually Model.View runs one last time after Quit?

		// Explicitly print the view if it wasn't the last frame.
		if view != "" {
			fmt.Fprintln(os.Stderr, view)
		}

		// 2. Handle JSON Data
		if opts.output != "" {
			// Write to file
			if err := os.WriteFile(opts.output, jsonBytes, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "Results saved to %s\n", opts.output)
			}
		} else {
			// Print to Stdout for piping
			fmt.Println(string(jsonBytes))
		}
	}
}
