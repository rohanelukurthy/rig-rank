# RigRank CLI Architecture

This document visualizes the internal structure of the Go-based `RigRank` CLI tool and its interactions with the system.

## component Diagram (Go)

```mermaid
graph TD
    User([User]) -->|Runs| CLI[RigRank CLI]

    subgraph RigRankApp [RigRank Application]
        Core[Core Logic]
        
        subgraph Interface [Interface Layer]
            Cobra>CLI Framework]
            TUI>Rich UI]
        end
        
        subgraph Telemetry [Telemetry Module]
            Gopsutil[[System Stats]]
            GHW[[HW Info]]
            SysCalls[[Direct Syscalls]]
        end
        
        subgraph BenchEngine [Benchmark Engine]
            HTTP[[HTTP Client]]
            Metrics[[Metric Collector]]
        end
        
        Model{Data Model}
    end
    
    subgraph HostSystem [Host System]
        Kernel[OS Kernel / Sysfs / WMI]
        Ollama[Ollama Service]
        Filesystem[(Local Disk)]
    end
    
    User --> Cobra
    Cobra --> Core
    Core --> TUI
    
    Core --> Telemetry
    Core --> BenchEngine
    
    Gopsutil -->|CPU/RAM Usage| Kernel
    GHW -->|PCIe/GPU Info| Kernel
    SysCalls -->|Low-level Info| Kernel
    
    HTTP -->|/api/generate| Ollama
    HTTP -->|/api/ps| Ollama
    
    Telemetry -->|Populates| Model
    BenchEngine -->|Populates| Model
    
    Model -->|Serializes JSON| Filesystem
    
    style CLI fill:#f9f,stroke:#333,stroke-width:2px
```

## dependency Trace

| module | Go Package | Purpose |
| :--- | :--- | :--- |
| **CLI Framework** | `github.com/spf13/cobra` | Command parsing, flags, help generation. |
| **TUI / UX** | `github.com/charmbracelet/bubbletea` | Animated, interactive terminal UI. |
| **Parsing** | `github.com/charmbracelet/lipgloss` | Styling and layout for the terminal. |
| **System Info** | `github.com/shirou/gopsutil` | Cross-platform CPU, Memory, Load stats. |
| **Hardware Info** | `github.com/jaypipes/ghw` | Deep introspection (PCIe, GPU, DMI) for Linux/Windows. |
| **HTTP Client** | `github.com/go-resty/resty` | Robust HTTP client for talking to Ollama API. |
| **Spinner** | `github.com/charmbracelet/bubbles` | Loading indicators during benchmark. |
