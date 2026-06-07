package main

import (
	"encoding/json"
	"log"
	"net/http"

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

func mcpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body")
		return
	}

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
		result, err = tools.Search(p)

	case "ask_llm":
		var p tools.AskLLMParams
		if err = json.Unmarshal(req.Params, &p); err != nil {
			writeError(w, "invalid params for ask_llm")
			return
		}
		result, err = tools.AskLLM(p)

	case "rank_answers":
		var p tools.RankParams
		if err = json.Unmarshal(req.Params, &p); err != nil {
			writeError(w, "invalid params for rank_answers")
			return
		}
		result, err = tools.RankAnswers(p)

	default:
		writeError(w, "unknown tool: "+req.Tool)
		return
	}

	if err != nil {
		writeError(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MCPResponse{Result: result})
}

func writeError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(MCPResponse{Error: msg})
}

func main() {
	http.HandleFunc("/mcp", mcpHandler)
	log.Println("MCP server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
