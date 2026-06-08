# mcp-server

## the usage of the GO

The Go server is just an HTTP wrapper around three functions — search, ask_llm, rank_answers. Python calls it sequentially, one field at a time. No goroutines, no concurrency yet.

Right now the bottleneck is:

```text
scrape field → ask_llm (personal) → search DDG → ask_llm (web) → rank → fill
```

For a 20-field form, that's 20 sequential LLM calls + 20 searches. Slow.

Go can parallelize the two independent calls — personal context answer and web search answer — since they don't depend on each other:

```text
                    ┌──  ask_llm(personal context)  ──┐
scrape field ──────►│                                 ├──────► rank_answers ──────► fill
                    └── search + ask_llm(web result) ─┘
```

Both goroutines fire simultaneously, rank waits for both, then filler gets the winner. That's where the architecture actually pays off vs pure Python.

We'll add this inside `rank.go` — a single `sync.WaitGroup` pattern when we wire `agent.py` and do end-to-end testing. It's a 20-line change.
For now Go is the right home for this because:

- LLM calls are network I/O — goroutines are ideal
- Keeps Python clean — Python only does DOM work (scrape + fill)
- Easy to extend later — add more tools as Go functions

## architect until [MVP-v1](https://github.com/Adityarya11/form-agent/commit/27b00b955c036b3403b531ee65dbac38748dd9d2)

Great questions both. Let me answer them clearly.

---

### JSON-RPC vs REST — what we're actually doing

- **plain REST right now** — POST to `/mcp` with a `tool` field. It's not JSON-RPC.

True JSON-RPC looks like:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "search",
  "params": { "query": "golang" }
}
```

The difference:
| | REST | JSON-RPC |
|---|---|---|
| Endpoint per tool | `/mcp/search`, `/mcp/ask_llm` | single `/rpc` |
| Tool name | in URL | in body as `method` |
| Batching | no | yes — send array of calls in one request |
| Standard | informal | formal spec |

**For this — REST is correct.** JSON-RPC makes sense when this would have a protocol-level client that auto-discovers tools (like actual MCP spec from Anthropic). I am building a custom agent, so REST is simpler and good enough.

---

### Buffered channels

- **worker pool with a buffered channel as a semaphore**. The idea:

```
fields: [f0, f1, f2, f3, f4, f5, f6...]
           ↓
     semaphore (buffer=3)
     ┌──────────────┐
     │ goroutine f0 │──► ask_llm + search simultaneously
     │ goroutine f1 │──► ask_llm + search simultaneously
     │ goroutine f2 │──► ask_llm + search simultaneously
     └──────────────┘
     f3 blocks until one slot frees
```

Each goroutine itself fires two sub-goroutines — personal and web — with a `WaitGroup`. So at peak this would have 3 fields × 2 calls = 6 concurrent Ollama/DDG calls.

The flow is now:

```
Python scrapes all fields
        ↓
Single HTTP call → Go resolve_batch
        ↓
Go: 3 fields at a time via buffered chan
    each field: personal + web fire simultaneously via WaitGroup
        ↓
All results back in one response
        ↓
Python fills sequentially (DOM work, can't parallelize)
```
