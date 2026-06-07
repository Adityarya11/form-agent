# Form Agent

```mermaid
flowchart TB

    subgraph PC["Python Client"]
        FS["Form Scraper<br/>(Playwright)"]
        DI["Doc Ingester<br/>(Resume)"]
        FF["Form Filler<br/>(Playwright)"]

        MCP["MCP Client<br/>(JSON-RPC)"]

        FS --> MCP
        DI --> MCP
        FF --> MCP
    end

    MCP -- "HTTP / stdio" --> SERVER

    subgraph SERVER["Go MCP Server"]
        SEARCH["Tool: search<br/>(DDG scrape)"]
        ASK["Tool: ask_llm<br/>(Ollama)"]
        RANK["Tool: rank_answers"]

        SEARCH --> ASK
        ASK --> RANK
    end

    SERVER --> OLLAMA

    OLLAMA["Ollama<br/>(Qwen 2.5)"]
```

### MCP (Go exposures)

Three tools over HTTP POST at `localhost:8080/mcp`:

```json
// Tool: ask_llm
{ "tool": "ask_llm", "params": { "question": "...", "context": "...", "options": ["a","b"] } }

// Tool: search
{ "tool": "search", "params": { "query": "...", "max_results": 3 } }

// Tool: rank_answers
{ "tool": "rank_answers", "params": { "question": "...", "candidate_a": "...", "candidate_b": "..." } }
```
