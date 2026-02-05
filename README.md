# RigRank

> **Local LLM Benchmarking & Hardware Suitability Tool**

RigRank is a CLI tool written in Go that evaluates your local hardware capabilities (CPU, GPU, RAM) and benchmarks them against standard Large Language Models (LLMs) running via [Ollama](https://ollama.com). It provides a comprehensive report on whether your "rig" is ready for real-world AI use cases like coding, storytelling, and reasoning.

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

```json
{
  "system_info": {
    "arch": "arm64",
    "cpu": { "model": "Apple M2 Max", "cores_physical": 12 },
    "ram": { "total_mb": 32768, "type": "LPDDR5", "speed_mts": 6400 }
  },
  "benchmarks": {
    "atomic": {
      "stats": { "ttft_ms": { "mean": 12.5, "p99": 15.2 } }
    }
  }
}
```

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