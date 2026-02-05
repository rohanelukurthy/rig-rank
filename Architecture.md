# RigRank CLI Architecture

This document visualizes the internal structure of the Go-based `RigRank` CLI tool and its interactions with the system.

## component Diagram (Go)

```mermaid
graph TD
    User([User]) -->|Runs| CLI[RigRank CLI]

    subgraph RigRankApp [RigRank Application]
        Core[Core Logic]
        
        subgraph Interface [Interface Layer]
            Cobra>"Cobra (Flags/Cmds)"]
            TUI>"Bubbletea (Active)"]
        end
        
        subgraph Telemetry [Telemetry Module]
            Gopsutil[["Gopsutil (CPU/RAM)"]]
            GHW[["GHW (GPU/PCIe)"]]
            SysProfiler[["macOS system_profiler"]]
        end
        
        subgraph BenchEngine [Benchmark Engine]
            Client[[Ollama Client]]
            Runner[[Benchmark Runner]]
        end
        
        Model{Data Model}
    end
    
    subgraph HostSystem [Host System]
        Kernel[OS Kernel / Sysfs]
        Ollama[Ollama Service]
        Filesystem[(Local Disk)]
    end
    
    User --> Cobra
    Cobra --> Core
    Core --> TUI
    
    Core --> Telemetry
    Core --> BenchEngine
    
    Gopsutil -->|CPU/RAM| Kernel
    GHW -->|Linux/Win GPU| Kernel
    SysProfiler -->|macOS RAM/GPU| Kernel
    
    Runner -->|Orchestrates| Client
    Client -->|POST /api/generate| Ollama
    Client -->|HEAD /| Ollama
    
    Telemetry -->|Populates| Model
    Runner -->|Populates| Model
    
    Model -->|Serializes JSON| Filesystem
    
    style CLI fill:#f9f,stroke:#333,stroke-width:2px
```

## Dependency Trace

| Module | Go Package | Purpose |
| :--- | :--- | :--- |
| **CLI Framework** | `github.com/spf13/cobra` | Command parsing, flags (`--model`, `--debug`), help generation. |
| **System Info** | `github.com/shirou/gopsutil` | Cross-platform CPU, Memory stats. |
| **Hardware Info** | `github.com/jaypipes/ghw` | Deep introspection (PCIe, GPU) for Linux/Windows. |
| **HTTP Client** | `github.com/go-resty/resty` | Robust HTTP client for talking to Ollama API. |
| **TUI / UX** | `github.com/charmbracelet/bubbletea` | Animated, interactive terminal UI. |
| **Spinner** | `github.com/charmbracelet/bubbles` | Loading indicators during benchmark. |
