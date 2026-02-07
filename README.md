# RigRank

> **Local LLM Benchmarking & Hardware Suitability Tool**

RigRank is a CLI tool written in Go that benchmarks how well Large Language Models (LLMs) run on **your specific hardware**. It measures real-world performance metrics via [Ollama](https://ollama.com) and tells you whether your "rig" is ready for AI workloads.

## 🎯 What RigRank Measures (and What It Doesn't)

**RigRank answers:** *"How fast does this model run on MY hardware?"*

| ✅ RigRank Measures | ❌ RigRank Does NOT Measure |
|---------------------|----------------------------|
| Tokens per second (throughput) | Model accuracy or intelligence |
| Time to first token (latency) | Response quality or correctness |
| Prompt processing speed | Benchmark scores (MMLU, HumanEval, etc.) |
| Model load times (cold start) | Reasoning or factual correctness |

**Why this distinction matters:** Local LLMs aren't 'one size fits all.' Depending on your CPU, GPU, and RAM, you might be forced to trade model intelligence (quantization) for usable speed. RigRank provides the telemetry you need to navigate these constraints, helping you find the perfect balance between reasoning depth and desktop snappiness.

## 🚀 Features

-   **Hardware Telemetry**: Automatically detects CPU cores, RAM type/speed (macOS), and GPU VRAM/Model.
-   **5-Stage Benchmark Suite**:
    -   **Atomic Check**: TTFT (Time To First Token) latency test.
    -   **Code Generation**: Evaluation of structured output performance.
    -   **Story Generation**: Long-context generation throughput.
    -   **Summarization**: Context ingestion speed testing.
    -   **Reasoning**: Logical processing capabilities.
-   **Ollama Integration**: Seamlessly connects to your local Ollama instance.
-   **JSON Reporting**: detailed, machine-readable output for analysis.

## 🛠️ Prerequisites

-   **Go 1.25+** (for building from source)
-   **Ollama**: Must be installed and running (`ollama serve`).
-   **Models**: You need at least one model pulled (e.g., `llama3`, `gemma:2b`).

## 📦 Installation

To install `rigrank` from source:

```bash
git clone https://github.com/rohanelukurthy/rig-rank.git
cd rigrank
make build-all # or go build -o rigrank ./cmd/rigrank
```

## 🏃 Usage

Ensure Ollama is running (`ollama serve`), then run the benchmark suite:

```bash
# Run with default model (llama3)
./rigrank run

# Run with a specific model
./rigrank run --model gemma2:9b

# Run with debug logging enabled
./rigrank run --model mistral --debug
```

### Options

| Flag | Shorthand | Description | Default |
| :--- | :--- | :--- | :--- |
| `--model` | `-m` | Ollama model name to benchmark | `llama3` |
| `--debug` | `-d` | Enable verbose debug logging | `false` |
| `--help` | `-h` | Show help for command | |

## 📊 Output Example

RigRank displays a human-friendly **Report Card** followed by detailed JSON metrics:

```
  📊 Model Report Card: gemma3:1b

  ┌────────────────────────────────────────────────────────────────────┐
  │  Benchmark       Startup      Writing Speed    Reading Speed      │
  │                  (first word) (output)         (input)            │
  ├────────────────────────────────────────────────────────────────────┤
  │  Atomic Check    343ms        ~70 words/sec    ~531 words/sec     │
  │  Code Gen        579ms        ~18 words/sec    ~657 words/sec     │
  │  Story Gen       624ms        ~19 words/sec    ~525 words/sec     │
  │  Summarization   732ms        ~17 words/sec    ~9.7k words/sec    │
  │  Reasoning       495ms        ~16 words/sec    ~1.1k words/sec    │
  └────────────────────────────────────────────────────────────────────┘

  ✅ Writing Speed: Excellent across all tasks.
  ⚠️  Startup: Noticeable pause before responses begin.

  This model is suitable for most tasks, but may struggle with some heavy workloads.
```

For the full JSON output schema, see [`examples/sample_output.json`](./examples/sample_output.json).

## 📈 Understanding the Metrics

The JSON output contains three key performance stats for each benchmark:

| Metric | Full Name | Plain English Explanation |
| :--- | :--- | :--- |
| **`ttft_ms`** | Time To First Token | **The "Snappiness" Metric.** How long you wait (in milliseconds) for the model to generate the *very first* word. Lower numbers mean the model feels more responsive. |
| **`gen_tps`** | Generation Tokens/Sec | **The "Writing Speed" Metric.** How fast the model generates the text of its response. Higher numbers mean long stories or code blocks finish faster. |
| **`prompt_tps`** | Prompt Processing Tokens/Sec | **The "Reading Speed" Metric.** How fast the model processes your input before it starts thinking. Crucial for summarizing large documents or chatting with long context. |

## 🏗️ Architecture

See [Architecture.md](./Architecture.md) for the high-level design and dependency graph.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1.  Fork the project
2.  Create your feature branch (`git checkout -b feature/AmazingFeature`)
3.  Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4.  Push to the branch (`git push origin feature/AmazingFeature`)
5.  Open a Pull Request

## 📄 License

Distributed under the MIT License.