package tools

import "sync"

type RankParams struct {
	Question   string `json:"question"`
	CandidateA string `json:"candidate_a"`
	CandidateB string `json:"candidate_b"`
}

type RankResult struct {
	Winner string `json:"winner"`
	Source string `json:"source"`
}

type FieldJob struct {
	Question string
	Options  []string
	Context  string
}

type FieldResult struct {
	Question string
	Winner   string
	Source   string
	Error    error
}

func RankAnswers(p RankParams) (RankResult, error) {
	prompt := "You are evaluating two candidate answers to a form question.\n\n" +
		"Question: " + p.Question + "\n\n" +
		"Candidate A (from personal context): " + p.CandidateA + "\n" +
		"Candidate B (from web search): " + p.CandidateB + "\n\n" +
		"Reply with exactly one word: either A or B."

	result, err := AskLLM(AskLLMParams{
		Question: prompt,
		Context:  "",
		Options:  []string{"A", "B"},
	})

	if err != nil {
		return RankResult{}, err
	}

	answer := result.Answer
	if len(answer) > 0 && (answer[0] == 'A' || answer[0] == 'a') {
		return RankResult{
				Winner: p.CandidateA,
				Source: "personal",
			},
			nil
	}

	if len(answer) > 0 && (answer[0] == 'B' || answer[0] == 'b') {
		return RankResult{
				Winner: p.CandidateB,
				Source: "web",
			},
			nil
	}

	return RankResult{Winner: p.CandidateA, Source: "personal_fallback"}, nil

}

func ResolveBatch(jobs []FieldJob, concurrency int) []FieldResult {
	sem := make(chan struct{}, concurrency) // make semaphore see `readme`
	results := make([]FieldResult, len(jobs))

	var wg sync.WaitGroup

	for i, job := range jobs {
		wg.Add(1)

		go func(idx int, j FieldJob) {
			defer wg.Done()

			sem <- struct{}{}

			defer func() { <-sem }()

			results[idx] = resolveDone(j)
		}(i, job)
	}

	wg.Wait()
	return results
}

func resolveDone(job FieldJob) FieldResult {
	var (
		personalAns, webAns string
		wg                  sync.WaitGroup
	)

	wg.Add(2)

	go func() {
		defer wg.Done()

		res, err := AskLLM(AskLLMParams{
			Question: job.Question,
			Context:  job.Context,
			Options:  job.Options,
		})

		if err != nil {
			personalAns = res.Answer
		}
	}()

	go func() {
		defer wg.Done()
		results, err := Search(SearchParams{
			Query:      job.Question,
			MaxResults: 3,
		})

		if err != nil || len(results) == 0 {
			return
		}

		ctx := ""
		for _, r := range results {
			ctx = r.Title + ": " + r.Snippet + "\n"
		}

		res, err := AskLLM(AskLLMParams{
			Question: job.Question,
			Context:  ctx,
			Options:  job.Options,
		})

		if err == nil {
			webAns = res.Answer
		}
	}()

	wg.Wait()

	if personalAns == "" && webAns == "" {
		return FieldResult{
			Question: job.Question,
			Winner:   "",
			Source:   "none",
		}
	}

	if personalAns == "" {
		return FieldResult{
			Question: job.Question,
			Winner:   webAns,
			Source:   "web",
		}
	}

	if webAns == "" {
		return FieldResult{
			Question: job.Question,
			Winner:   personalAns,
			Source:   "personal",
		}
	}

	ranked, err := RankAnswers(RankParams{
		Question:   job.Question,
		CandidateA: personalAns,
		CandidateB: webAns,
	})

	if err != nil {
		return FieldResult{
			Question: job.Question,
			Winner:   personalAns,
			Source:   "personal_fallback",
		}
	}

	return FieldResult{
		Question: job.Question,
		Winner:   ranked.Winner,
		Source:   ranked.Source,
	}
}
