package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"form-agent/mcp-server/tools"
)

type MCPRequest struct {
	Tool   string          `json:"tool"`
	Params json.RawMessage `json:"params"`
}

type MCPResponse struct {
	Result any    `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

var logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelDebug,
}))

func mcpHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request", "error", err)
		writeError(w, "invalid request body")
		return
	}

	logger.Info("incoming request", "tool", req.Tool)

	var (
		result any
		err    error
	)

	switch req.Tool {
	case "search":
		var p tools.SearchParams
		if err = json.Unmarshal(req.Params, &p); err != nil {
			writeError(w, "invalid params for search")
			return
		}
		logger.Debug("search params", "query", p.Query, "max_results", p.MaxResults)
		result, err = tools.Search(p)

	case "ask_llm":
		var p tools.AskLLMParams
		if err = json.Unmarshal(req.Params, &p); err != nil {
			writeError(w, "invalid params for ask_llm")
			return
		}
		logger.Debug("ask_llm params", "question", p.Question, "options_count", len(p.Options))
		result, err = tools.AskLLM(p)

	case "rank_answers":
		var p tools.RankParams
		if err = json.Unmarshal(req.Params, &p); err != nil {
			writeError(w, "invalid params for rank_answers")
			return
		}
		logger.Debug("rank_answers params", "question", p.Question)
		result, err = tools.RankAnswers(p)

		/*
		* this is the portion for the batch resolving, means now the go routines will be activated and
		* LLM calling will be done in 3 async threads.
		 */
	case "resolve_batch":
		var jobs []tools.FieldJob
		if err = json.Unmarshal(req.Params, &jobs); err != nil {
			writeError(w, "invalid params for the resolve_batch")
			return
		}

		logger.Debug("resolve_batch", "job_count", len(jobs))
		result = tools.ResolveBatch(jobs, 3)

	default:
		logger.Warn("unknown tool called", "tool", req.Tool)
		writeError(w, "unknown tool: "+req.Tool)
		return
	}

	if err != nil {
		logger.Error("tool execution failed", "tool", req.Tool, "error", err)
		writeError(w, err.Error())
		return
	}

	logger.Info("tool completed", "tool", req.Tool, "duration_ms", time.Since(start).Milliseconds())

	resultJSON, _ := json.Marshal(result)
	logger.Debug("tool result", "tool", req.Tool, "result", string(resultJSON))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MCPResponse{Result: result})
}

func writeError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(MCPResponse{Error: msg})
}

func main() {
	logger.Info("MCP server starting", "addr", ":8080")
	http.HandleFunc("/mcp", mcpHandler)
	logger.Info("MCP server ready")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
