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
