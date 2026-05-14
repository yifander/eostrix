package leetcode

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type LeetCodeClient struct {
	Client  *http.Client
	BaseURL string
}

func NewLeetCodeClient() *LeetCodeClient {
	return &LeetCodeClient{
		Client: &http.Client{Timeout: 15 * time.Second},
	}
}

type QuestionDetail struct {
	QuestionId string `json:"questionId"`
	Title      string `json:"title"`
	TitleSlug  string `json:"titleSlug"`
	Difficulty string `json:"difficulty"`
	TopicTags  []struct {
		Name string `json:"name"`
	} `json:"topicTags"`
	Stats  string  `json:"stats"`
	AcRate float64 `json:"acRate"`
}

type graphqlRequest struct {
	Query     string            `json:"query"`
	Variables map[string]string `json:"variables"`
}

type graphqlResponse struct {
	Data struct {
		Question QuestionDetail `json:"question"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

const questionQuery = `
query getQuestionDetail($titleSlug: String, $questionId: String) {
  question(titleSlug: $titleSlug, questionId: $questionId) {
    questionId
    title
    titleSlug
    difficulty
    topicTags { name }
    stats        # ← Scalar, no sub-selection
    acRate
  }
}`

func (c *LeetCodeClient) GetBySlug(ctx context.Context, slug string) (*QuestionDetail, error) {
	return c.fetch(ctx, "titleSlug", slug)
}

func (c *LeetCodeClient) GetByID(ctx context.Context, id int) (*QuestionDetail, error) {
	return c.fetch(ctx, "questionId", fmt.Sprintf("%d", id))
}

func (c *LeetCodeClient) fetch(ctx context.Context, key, value string) (*QuestionDetail, error) {
	reqBody := graphqlRequest{
		Query:     questionQuery,
		Variables: map[string]string{key: value},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	endpoint := c.BaseURL
	if endpoint == "" {
		endpoint = "https://leetcode.com/graphql"
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Origin", "https://leetcode.com")
	req.Header.Set("Referer", "https://leetcode.com/problems/")
	req.Header.Set("x-requested-with", "XMLHttpRequest")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var gqlResp graphqlResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %s", gqlResp.Errors[0].Message)
	}

	if gqlResp.Data.Question.TitleSlug == "" {
		return nil, fmt.Errorf("problem not found")
	}

	return &gqlResp.Data.Question, nil
}
