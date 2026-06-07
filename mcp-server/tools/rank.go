package tools

import (
	"fmt"
	"strings"
)

type RankParams struct {
	Question   string `json:"question"`
	CandidateA string `json:"candidate_a"`
	CandidateB string `json:"candidate_b"`
}

type RankResult struct {
	Winner string `json:"winner"`
	Source string `json:"source"`
}

func RankAnswers(p RankParams) (RankResult, error) {
	prompt := fmt.Sprintf(
		`You are evaluating two candidate answers to a form question.

Question: %s

Candidate A (from personal context): %s
Candidate B (from web search): %s

Which candidate is more accurate and appropriate for this question?
Reply with exactly one word: either "A" or "B".`,
		p.Question, p.CandidateA, p.CandidateB,
	)

	result, err := AskLLM(AskLLMParams{
		Question: prompt,
		Context:  "",
		Options:  []string{"A", "B"},
	})
	if err != nil {
		return RankResult{}, err
	}

	answer := strings.ToUpper(strings.TrimSpace(result.Answer))

	switch {
	case strings.HasPrefix(answer, "A"):
		return RankResult{Winner: p.CandidateA, Source: "personal"}, nil
	case strings.HasPrefix(answer, "B"):
		return RankResult{Winner: p.CandidateB, Source: "web"}, nil
	default:
		return RankResult{Winner: p.CandidateA, Source: "personal_fallback"}, nil
	}
}
