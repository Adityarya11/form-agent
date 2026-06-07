package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

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

	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewReader(body))
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

	sb.WriteString("You are a form-filling assistant. Answer the question below using the provided context.\n\n")

	if p.Context != "" {
		sb.WriteString("Context:\n")
		sb.WriteString(p.Context)
		sb.WriteString("\n\n")
	}

	sb.WriteString("Question: ")
	sb.WriteString(p.Question)
	sb.WriteString("\n")

	if len(p.Options) > 0 {
		sb.WriteString("Available options: ")
		sb.WriteString(strings.Join(p.Options, ", "))
		sb.WriteString("\nPick the single best option and reply with only that option text, nothing else.\n")
	} else {
		sb.WriteString("Reply with a concise, direct answer only.\n")
	}

	return sb.String()
}
