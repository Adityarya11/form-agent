package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var ollamaClient = &http.Client{
	Timeout: 120 * time.Second,
}

type AskLLMParams struct {
	Question string   `json:"question"`
	Context  string   `json:"context"`
	Options  []string `json:"options"`
}

type AskLLMResult struct {
	Answer string `json:"answer"`
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
}

func AskLLM(p AskLLMParams) (AskLLMResult, error) {
	prompt := buildPrompt(p)

	body, _ := json.Marshal(ollamaRequest{
		Model:  "qwen2.5:3b",
		Prompt: prompt,
		Stream: false,
	})

	resp, err := ollamaClient.Post("http://localhost:11434/api/generate", "application/json", bytes.NewReader(body))
	if err != nil {
		return AskLLMResult{}, fmt.Errorf("ollama unreachable: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return AskLLMResult{}, fmt.Errorf("reading ollama response: %w", err)
	}

	var out ollamaResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return AskLLMResult{}, fmt.Errorf("parsing ollama response: %w", err)
	}

	return AskLLMResult{Answer: strings.TrimSpace(out.Response)}, nil
}

func buildPrompt(p AskLLMParams) string {
	var sb strings.Builder

	if len(p.Options) > 0 {
		sb.WriteString("You are a form-filling assistant. You MUST reply with ONLY one of the provided options, word for word. No explanation. No punctuation. Just the option text.\n\n")
		if p.Context != "" {
			sb.WriteString("Context:\n")
			sb.WriteString(p.Context)
			sb.WriteString("\n\n")
		}
		sb.WriteString("Question: ")
		sb.WriteString(p.Question)
		sb.WriteString("\n")
		sb.WriteString("Options: ")
		sb.WriteString(strings.Join(p.Options, " | "))
		sb.WriteString("\n\nReply with exactly one option from the list above. Nothing else.")
	} else {
		sb.WriteString("You are a form-filling assistant. Reply with a short, direct answer only. No explanation.\n\n")
		if p.Context != "" {
			sb.WriteString("Context:\n")
			sb.WriteString(p.Context)
			sb.WriteString("\n\n")
		}
		sb.WriteString("Question: ")
		sb.WriteString(p.Question)
		sb.WriteString("\nAnswer:")
	}

	return sb.String()
}
