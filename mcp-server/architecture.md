**AI-Generated**

This architecture is a masterclass in using Go's lightweight concurrency primitives to solve a heavy-lift hardware problem. By using a **channel-based semaphore** instead of a worker pool, you prioritize code simplicity and allow the Go scheduler to handle the heavy lifting.

The following technical note summarizes the mechanics, metrics, and the hardware-driven logic behind your design.

---

## 2D Architectural Blueprint: Batch Processing in Go

### 1. High-Level System Workflow

The diagram below illustrates the lifecycle of a batch request. It highlights the "bottleneck by design" where 25 goroutines are compressed into 3 execution slots to protect the local Ollama instance.

```mermaid
graph TD

    %% Styling
    classDef client fill:#E3F2FD,stroke:#1565C0,stroke-width:2px,color:#0D47A1;
    classDef runtime fill:#FFF3E0,stroke:#EF6C00,stroke-width:2px,color:#E65100;
    classDef active fill:#E8F5E9,stroke:#2E7D32,stroke-width:2px,color:#1B5E20,font-weight:bold;
    classDef blocked fill:#FFEBEE,stroke:#C62828,stroke-width:2px,color:#B71C1C;
    classDef infra fill:#F3E5F5,stroke:#6A1B9A,stroke-width:2px,color:#4A148C;

    %% Components
    User["Batch Request: 25 Jobs"]:::client
    Init["resolve_batch\nconcurrency=3"]:::runtime

    subgraph Semaphore_Gate["Semaphore: Buffered Channel"]
        Slot1["Slot 1: Full"]:::active
        Slot2["Slot 2: Full"]:::active
        Slot3["Slot 3: Full"]:::active
    end

    subgraph Waiting_Room["Blocked Goroutines"]
        W1["W1"]:::blocked
        W2["W2"]:::blocked
        Wetc["..."]:::blocked
        W25["W25"]:::blocked
    end

    subgraph Intra_Job["Per-Worker Parallelism"]
        LLM_A["Sub-Task A\nPersonal Context"]:::infra
        LLM_B["Sub-Task B\nWeb + LLM"]:::infra
    end

    Ollama["Local Ollama\nQwen 2.5:3B"]:::infra
    Release["<- sem"]:::runtime

    %% Connections
    User --> Init
    Init -->|"Spawn 25"| W4
    Init -->|"Spawn 25"| W5
    Init -->|"Spawn 25"| Wetc
    Init -->|"Spawn 25"| W25

    W4 -->|"Attempt sem <- struct{}"| Slot1
    W5 -->|"Attempt sem <- struct{}"| Slot2
    Wetc -->|"Attempt sem <- struct{}"| Slot3

    Slot1 -->|"Executes"| LLM_A
    Slot2 -->|"Executes"| LLM_A
    Slot3 -->|"Executes"| LLM_A

    LLM_A -->|"Request 1"| Ollama
    LLM_B -->|"Request 2"| Ollama

    LLM_A --> LLM_B

    LLM_B -->|"Done"| Release
    Release -->|"Freed Slot"| Slot1

```

---

### 2. The Execution Calculus

When your system is running at maximum capacity, the resource distribution follows a specific mathematical split. This ensures that while the LLM is the bottleneck, the CPU (for web searching) remains fully saturated.

| Component                 | Logic / Formula          | Value (at Peak)        |
| ------------------------- | ------------------------ | ---------------------- |
| **Total Goroutines**      | `N` (Total Jobs)         | **25**                 |
| **Active Parent Workers** | `C` (Concurrency Cap)    | **3**                  |
| **Blocked Workers**       | `N - C`                  | **22**                 |
| **Internal Sub-Tasks**    | `C * 2` (A + B per slot) | **6**                  |
| **Memory Overhead**       | `25 * ~2KB`              | **~50KB** (Negligible) |

---

### 3. Critical Technical Rationale

#### The "Lightweight Worker" Philosophy

Unlike languages where threads are expensive (like Java or Python), Go's goroutines are cheap.

- **Design Choice:** You chose to spawn all 25 workers immediately and let them block on a channel.
- **Benefit:** You don't need a complex "Dispatcher" or "Task Queue." The channel _is_ the queue. As soon as one worker finishes, the next one is "woken up" by the runtime scheduler instantly.

#### Handling the Ollama Bottleneck

Running a local **Qwen 2.5:3b** model presents a specific challenge: **VRAM and Compute serialization.**

1. **The Trap:** If you sent all 25 requests to Ollama at once, the inference engine would either crash with an Out-Of-Memory (OOM) error or queue them internally, adding massive latency due to context-switching overhead.
2. **The Solution:** By capping concurrency at **3**, you ensure that only **6 sub-tasks** (3 personal, 3 web-search) are hitting the LLM API at once.
3. **The Result:** This keeps the GPU memory stable and ensures the model is always "hot" and generating text rather than shuffling contexts in and out of memory.

#### Intra-Job Parallelism (`resolveOne`)

Each worker is "greedy." Once it gains a semaphore token, it doesn't just do one thing; it splits into two more goroutines:

- **Path A (Internal):** High speed, low latency.
- **Path B (External):** High latency (waiting for web results).
  By running these in parallel _within_ the worker, you ensure that the worker spends less time idle. The `sync.WaitGroup` inside `resolveOne` acts as a local synchronization point before the final ranking logic.

---

### 4. Summary for Implementation

> **The Key Takeaway:** Your architecture uses Go's **Concurrency** (the ability to manage many tasks) to mask the lack of **Parallelism** (the limited ability of local hardware to do many things at once). It is a "throttled firehose" approach that maximizes hardware safety without sacrificing the simplicity of the code.
