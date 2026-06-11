package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var ollamaClient = &http.Client{Timeout: 120 * time.Second}
var geminiClient = &http.Client{Timeout: 30 * time.Second}

var UseCloudLLM = false

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

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiCandidate struct {
	Content geminiContent `json:"content"`
}

type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
}

func AskLLM(p AskLLMParams) (AskLLMResult, error) {
	if UseCloudLLM {
		return askGemini(p)
	}
	return askOllama(p)
}

func askOllama(p AskLLMParams) (AskLLMResult, error) {
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

func askGemini(p AskLLMParams) (AskLLMResult, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return AskLLMResult{}, fmt.Errorf("GEMINI_API_KEY not set")
	}

	prompt := buildPrompt(p)

	reqBody, _ := json.Marshal(geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: prompt}}},
		},
	})

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=" + apiKey

	resp, err := geminiClient.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return AskLLMResult{}, fmt.Errorf("gemini unreachable: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return AskLLMResult{}, fmt.Errorf("reading gemini response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return AskLLMResult{}, fmt.Errorf("gemini error %d: %s", resp.StatusCode, string(raw))
	}

	var out geminiResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return AskLLMResult{}, fmt.Errorf("parsing gemini response: %w", err)
	}

	if len(out.Candidates) == 0 || len(out.Candidates[0].Content.Parts) == 0 {
		return AskLLMResult{}, fmt.Errorf("gemini returned empty response")
	}

	answer := strings.TrimSpace(out.Candidates[0].Content.Parts[0].Text)
	return AskLLMResult{Answer: answer}, nil
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
