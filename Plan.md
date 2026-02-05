# Plan: RigRank Local Command Implementation

## Goal
Develop a local CLI tool to execute hardware telemetry extraction and LLM inference benchmarking, producing a standardized JSON output file. This tool is the client-side component of RigRank, built in **Go (Golang)** for ease of distribution and performance.

## Architecture
See [Architecture.md](./Architecture.md) for a detailed component diagram and dependency graph.

## 1. System Telemetry Module
**Objective**: accurately fingerprint the host hardware using Go libraries.

### 1.1 CPU & OS Detection
- **Library**: `github.com/shirou/gopsutil`
- **Fields**:
    - Architecture (runtime.GOARCH)
    - Model Name (`cpu.Info()`)
    - Core Counts (`cpu.Counts()`)

### 1.2 RAM Telemetry
- **Library**: `github.com/shirou/gopsutil/mem` + direct system calls where needed.
- **Fields**: Total Capacity, Type, Speed.
- **Strategy**:
  - **Windows**: WMI queries via `github.com/yusufpapurcu/wmi`.
  - **Linux**: Parsing `/proc/meminfo` or `dmidecode` (via exec).
  - **macOS**: `sysctl` calls.

### 1.3 GPU & PCIe Analysis
- **Library**: `github.com/jaypipes/ghw` (Great hardware discovery for Linux/Windows).
- **Fields**: Model, VRAM, PCIe Topology.

## 2. Benchmark Engine Integration
**Objective**: Run a "Standard Suite" of inference tests to capture performance across the compute-bandwidth spectrum. Analysis of competitors (LocalScore, llama-bench) shows that a single metric is insufficient.

### 2.1 Strategy: The "RigRank Standard Suite"
We run **5 distinct test profiles**, each designed to directly answer a use-case question from Section 6. The prompts and token lengths are derived from industry benchmarks (LocalScore, HumanEval, MT-Bench, GSM8K).

| Profile | Use-Case Target | Input Tokens | Output Tokens | Primary Metric |
| :------ | :-------------- | :----------- | :------------ | :------------- |
| **Atomic** | Quick Q&A | 32 | 16 | TTFT |
| **Code Gen** | Coding | 80 | 256 | Gen TPS |
| **Story Gen** | Writing | 50 | 400 | Gen TPS |
| **Summarize** | Summarization/RAG | 2048 | 128 | Prompt TPS |
| **Reasoning** | Math/Analysis | 100 | 150 | Combined |

---

1.  **Atomic Check (Quick Q&A)**
    *   **Goal**: Measure responsiveness for simple interactive queries.
    *   **Prompt**: `"What is the capital of France? Answer in one word."`
    *   **Config**: `input: 32, output: 16`
    *   **Metric**: **TTFT (ms)**. Threshold: <50ms = EXCELLENT.

2.  **Code Generation (Coding)**
    *   **Goal**: Measure generation speed for code completion and short function writing.
    *   **Prompt**: `"Write a Python function that takes a list of integers and returns the second largest element. Include docstring and type hints."`
    *   **Config**: `input: 80, output: 256`
    *   **Metric**: **Gen TPS**. Threshold: >30 tok/s = GOOD.

3.  **Story Generation (Creative Writing)**
    *   **Goal**: Measure sustained generation for long-form creative output.
    *   **Prompt**: `"Write a short story about a robot who discovers the meaning of friendship. The story should be approximately 300 words."`
    *   **Config**: `input: 50, output: 400`
    *   **Metric**: **Gen TPS**. Threshold: >25 tok/s = GOOD.

4.  **Summarization (RAG / Document Processing)**
    *   **Goal**: Measure prompt ingestion speed for large context windows.
    *   **Prompt**: `[Standard 2000-token Lorem Ipsum block] + "Summarize the above text in 3 bullet points."`
    *   **Config**: `input: 2048, output: 128`
    *   **Metric**: **Prompt TPS**. Threshold: >150 tok/s = GOOD.

5.  **Reasoning (Math / Data Analysis)**
    *   **Goal**: Measure balanced performance for multi-step analytical tasks.
    *   **Prompt**: `"A store sells apples for $0.50 each and oranges for $0.75 each. If Maria buys 12 apples and 8 oranges, how much does she spend in total? Show your reasoning step by step."`
    *   **Config**: `input: 100, output: 150`
    *   **Metric**: **Combined (Avg of Prompt TPS & Gen TPS)**. Threshold: P50 across all profiles.

### 2.2 Execution Flow per Profile
For EACH profile above:
1.  **Warm-up**: Run the prompt once (discard results) to fill KV cache and wake up GPU.
2.  **Sampling**: Run the prompt **5 times**.
3.  **Aggregation**: Calculate Mean, Median, and P99 for TTFT and Token Speeds.

### 2.3 Backend Interaction (Ollama)
- **Client**: `github.com/go-resty/resty`.
- **API Endpoint**: `/api/generate`.
- **Payload Config**:
    ```json
    {
      "model": "llama3-8b-instruct",
      "prompt": "[...]",
      "options": {
        "temperature": 0.0,        // Deterministic
        "num_ctx": 4096,           // Standard context
        "num_predict": 512,        // Limit generation
        "seed": 42                 // Reproducibility
      },
      "stream": false              // We need precise timing metadata from valid HTTP response
    }
    ```
    *Note: We might switch to `stream: true` if we need to calculate TTFT manually client-side, standard Ollama API provides `total_duration`, `load_duration`, `prompt_eval_count`, `prompt_eval_duration`, `eval_count`, `eval_duration` in the final response object which is sufficient.*

### 2.4 Metrics Formulation
We will derive metrics directly from Ollama's response object:
-   **TTFT** = `total_duration` - `eval_duration` - `prompt_eval_duration` (Approximate) OR measured client-side (Request Start -> First Byte).
-   **Prompt Speed** = `prompt_eval_count` / `prompt_eval_duration`.
-   **Gen Speed** = `eval_count` / `eval_duration`.

## 3. User Interface (CLI)
- **Framework**: `github.com/spf13/cobra` for commands (`rigrank run`, `rigrank version`).
- **TUI**: `github.com/charmbracelet/bubbletea` for a premium, animated progress view during the benchmark run.

## 4. Output Generation
**Objective**: Produce the schema-compliant JSON.

### 4.1 Data Aggregation
- Combine `system_info`, `inference_results`, and `quality_scores` dictionaries.

### 4.2 JSON Serialization
- Save to `./rigrank_result.json`.
- Validate against schema defined below.

## 5. Data Schema

```json
{
  "system_info": {
    "arch": "arm64 | x86_64",
    "cpu": {
      "model": "Apple M2 Ultra",
      "cores_physical": 24,
      "cores_logical": 24,
      "frequency_max_mhz": 3500
    },
    "gpu": {
      "model": "Integrated Apple GPU",
      "vram_total_mb": 131072,
      "pcie_gen": null,
      "pcie_lanes": null
    },
    "ram": {
      "total_mb": 131072,
      "type": "LPDDR5",
      "speed_mts": 6400
    }
  },
  "inference_results": {
    "metrics_version": "1.0",
    "model_metadata": {
      "name": "llama3-8b-instruct",
      "quantization": "Q4_K_M",
      "size_mb": 4500
    },
    "benchmarks": {
      "atomic": {
        "description": "Quick Q&A (Latency Focus)",
        "config": { "input_tokens": 32, "output_tokens": 16 },
        "stats": {
          "ttft_ms": { "mean": 12.5, "median": 12.0, "p99": 15.2 }
        }
      },
      "code_gen": {
        "description": "Coding (Generation Speed)",
        "config": { "input_tokens": 80, "output_tokens": 256 },
        "stats": {
          "gen_tps": { "mean": 48.0, "median": 48.5, "p99": 44.0 }
        }
      },
      "story_gen": {
        "description": "Creative Writing (Sustained Generation)",
        "config": { "input_tokens": 50, "output_tokens": 400 },
        "stats": {
          "gen_tps": { "mean": 42.0, "median": 43.0, "p99": 38.0 }
        }
      },
      "summarization": {
        "description": "RAG / Document Processing (Prompt Ingestion)",
        "config": { "input_tokens": 2048, "output_tokens": 128 },
        "stats": {
          "prompt_tps": { "mean": 210.5, "median": 212.0, "p99": 190.0 }
        }
      },
      "reasoning": {
        "description": "Math / Analysis (Balanced Performance)",
        "config": { "input_tokens": 100, "output_tokens": 150 },
        "stats": {
          "prompt_tps": { "mean": 180.0, "median": 182.0, "p99": 165.0 },
          "gen_tps": { "mean": 45.0, "median": 46.0, "p99": 41.0 }
        }
      }
    }
  },
  "use_case_suitability": {
    "quick_qa": { "rating": "EXCELLENT", "reason": "TTFT of 12ms is very responsive." },
    "coding": { "rating": "GOOD", "reason": "Generation speed of 48 tok/s is sufficient for code completion." },
    "writing": { "rating": "GOOD", "reason": "Generation speed of 42 tok/s is comfortable for drafting." },
    "summarization": { "rating": "EXCELLENT", "reason": "Prompt processing is fast for large documents." },
    "data_analysis": { "rating": "GOOD", "reason": "Balanced performance for analytical tasks." }
  },
  "overall_verdict": "This model performs well on your hardware for most interactive tasks."
}
```

## Execution Flow
1. **Init**: User runs `rigrank run`.
2. **Detect**: System sweeps hardware info.
3. **Benchmark**:
   - Downloads/Checks Model.
   - Runs Inference.
   - Captures Metrics.
4. **Report**: Saves JSON and prints summary table to console.

---

## 6. User-Friendly Output: Use-Case Suitability

**Philosophy**: Raw numbers (tokens/sec, TTFT) mean nothing to most users. The goal is to answer: *"Will this model on my hardware work for my use case at a reasonable speed?"*

### 6.1 Use-Case Categories
Inspired by LMSYS Chatbot Arena and OpenLLM Leaderboard, we map the 5 benchmark results to common use-case categories:

| Use Case        | Mapped Test Profile | Primary Metric          | Threshold for "Good"     |
| :-------------- | :------------------ | :---------------------- | :----------------------- |
| **Quick Q&A**   | Atomic              | TTFT (ms)               | < 50ms = EXCELLENT       |
| **Coding**      | Code Gen            | Gen TPS                 | > 30 tok/s = GOOD        |
| **Writing**     | Story Gen           | Gen TPS                 | > 25 tok/s = GOOD        |
| **Summarization / RAG** | Summarization   | Prompt TPS              | > 150 tok/s = GOOD       |
| **Data Analysis** | Reasoning         | Combined (Prompt + Gen) | Balanced performance     |

### 6.2 Plain-English Verdict
The CLI will output a summary section that answers the user's core question directly:

```
╭─────────────────────────────────────────────────────────────────────╮
│  RigRank Use-Case Suitability Report                                 │
├─────────────────────────────────────────────────────────────────────┤
│  Model: llama3:8b-instruct-q4_K_M                                    │
│  Hardware: Apple M2 Ultra (128GB RAM)                                │
├─────────────────────────────────────────────────────────────────────┤
│  ✅ Quick Q&A:       EXCELLENT (TTFT: 12ms)                         │
│  ✅ Coding:          GOOD      (Gen Speed: 48 tok/s)                │
│  ✅ Creative Writing: GOOD      (Gen Speed: 42 tok/s)                │
│  ✅ Summarization:   EXCELLENT (Prompt Speed: 210 tok/s)            │
│  ✅ Data Analysis:   GOOD      (Balanced: 180/45 tok/s)             │
├─────────────────────────────────────────────────────────────────────┤
│  Overall Verdict:                                                    │
│  "This model performs well on your hardware for most interactive    │
│   tasks and analytical workloads."                                   │
╰─────────────────────────────────────────────────────────────────────╯
```

### 6.3 Scoring Logic
We will assign a rating per use-case based on thresholds:
-   **EXCELLENT**: Metric is in the top 20% of known hardware results OR exceeds a "gold standard" threshold.
-   **GOOD**: Metric is above a minimum usability threshold.
-   **MARGINAL**: Metric is borderline; users may notice slowdowns.
-   **POOR**: Metric is below the usability threshold; not recommended for this use case.

### 6.4 Schema Update for Suitability
Add a new top-level key to the output JSON:

```json
{
  "use_case_suitability": {
    "quick_qa": { "rating": "EXCELLENT", "reason": "TTFT of 12ms is very responsive." },
    "coding": { "rating": "GOOD", "reason": "Generation speed of 48 tok/s is sufficient for code completion." },
    "writing": { "rating": "GOOD", "reason": "Generation speed is comfortable for drafting." },
    "summarization": { "rating": "EXCELLENT", "reason": "Prompt processing is fast for large documents." },
    "data_analysis": { "rating": "MARGINAL", "reason": "Context window limits may affect complex analysis." }
  },
  "overall_verdict": "This model performs well on your hardware for most interactive tasks. You may experience delays for very long documents."
}
```