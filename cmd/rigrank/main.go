package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rigrank",
		Short: "RigRank - Local LLM Benchmark Tool",
		Long:  `RigRank checks your hardware capabilities and benchmarks standard LLMs locally via Ollama.`,
	}

	cmd.AddCommand(newRunCmd())
	return cmd
}
