package tools

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

type SearchParams struct {
	Query      string `json:"query"`
	MaxResults int    `json:"max_results"`
}

type SearchResult struct {
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
	URL     string `json:"url"`
}

func Search(p SearchParams) ([]SearchResult, error) {
	if p.MaxResults <= 0 {
		p.MaxResults = 3
	}

	endpoint := "https://lite.duckduckgo.com/lite/"
	form := url.Values{}
	form.Set("q", p.Query)

	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/120.0 Safari/537.36")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return parseResults(string(body), p.MaxResults), nil
}

func parseResults(body string, max int) []SearchResult {
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return nil
	}

	var results []SearchResult
	var walk func(*html.Node)

	walk = func(n *html.Node) {
		if len(results) >= max {
			return
		}
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "result-link") {
					title := extractText(n)
					href := ""
					for _, a := range n.Attr {
						if a.Key == "href" {
							href = a.Val
						}
					}
					snippet := extractSiblingSnippet(n)
					if title != "" {
						results = append(results, SearchResult{
							Title:   title,
							Snippet: snippet,
							URL:     href,
						})
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(doc)
	return results
}

func extractText(n *html.Node) string {
	var sb strings.Builder
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return strings.TrimSpace(sb.String())
}

func extractSiblingSnippet(n *html.Node) string {
	if n.Parent == nil {
		return ""
	}
	for sib := n.NextSibling; sib != nil; sib = sib.NextSibling {
		if sib.Type == html.ElementNode {
			text := extractText(sib)
			if text != "" {
				return text
			}
		}
	}
	return ""
}
