package leetcode

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// mockLeetCodeServer creates a test server that simulates LeetCode's GraphQL API
func mockLeetCodeServer(t *testing.T) (*httptest.Server, *LeetCodeClient) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var reqBody struct {
			Query     string            `json:"query"`
			Variables map[string]string `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		switch {
		case reqBody.Variables["titleSlug"] == "two-sum":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(twoSumSuccessResp))
		case reqBody.Variables["questionId"] == "1":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(twoSumSuccessResp))
		case reqBody.Variables["titleSlug"] == "not-found":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(notFoundResp))
		case reqBody.Variables["titleSlug"] == "graphql-error":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(graphqlErrorResponse))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	t.Cleanup(server.Close)

	client := &LeetCodeClient{
		Client:  server.Client(),
		BaseURL: server.URL,
	}

	return server, client
}

const twoSumSuccessResp = `{
  "data": {
    "question": {
      "questionId": "1",
      "title": "Two Sum",
      "titleSlug": "two-sum",
      "difficulty": "Easy",
      "topicTags": [{"name": "Array"}, {"name": "Hash Table"}],
      "stats": {"totalAccepted": "10,543,210", "totalSubmission": "23,456,789"}
    }
  }
}`

const notFoundResp = `{"data": {"question": {}}}`

const graphqlErrorResponse = `{
  "data": {"question": null},
  "errors": [{"message": "Rate limit exceeded"}]
}`

func TestLeetCodeClient_GetBySlug_Success(t *testing.T) {
	_, client := mockLeetCodeServer(t)

	q, err := client.GetBySlug(context.Background(), "two-sum")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertTwoSum(t, q)
}

func TestLeetCodeClient_GetByID_Success(t *testing.T) {
	_, client := mockLeetCodeServer(t)

	q, err := client.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertTwoSum(t, q)
}

func TestLeetCodeClient_NotFound(t *testing.T) {
	_, client := mockLeetCodeServer(t)

	_, err := client.GetBySlug(context.Background(), "not-found")
	if err == nil {
		t.Fatal("expected error for not found problem, got nil")
	}
	if err.Error() != "problem not found" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLeetCodeClient_GraphQLError(t *testing.T) {
	_, client := mockLeetCodeServer(t)

	_, err := client.GetBySlug(context.Background(), "graphql-error")
	if err == nil {
		t.Fatal("expected error for GraphQL error, got nil")
	}
	if err.Error() != "GraphQL error: Rate limit exceeded" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLeetCodeClient_ContextTimeout(t *testing.T) {
	// Create a slow server
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer slowServer.Close()

	client := &LeetCodeClient{
		Client:  slowServer.Client(),
		BaseURL: slowServer.URL,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.GetBySlug(ctx, "slow-problem")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

// Helper to verify parsed fields
func assertTwoSum(t *testing.T, q *QuestionDetail) {
	t.Helper()

	if q.QuestionId != "1" {
		t.Errorf("expected questionId '1', got %q", q.QuestionId)
	}
	if q.Title != "Two Sum" {
		t.Errorf("expected title 'Two Sum', got %q", q.Title)
	}
	if q.TitleSlug != "two-sum" {
		t.Errorf("expected titleSlug 'two-sum', got %q", q.TitleSlug)
	}
	if q.Difficulty != "Easy" {
		t.Errorf("expected difficulty 'Easy', got %q", q.Difficulty)
	}
	if len(q.TopicTags) != 2 {
		t.Errorf("expected 2 topic tags, got %d", len(q.TopicTags))
	}
	if q.TopicTags[0].Name != "Array" {
		t.Errorf("expected first tag 'Array', got %q", q.TopicTags[0].Name)
	}
}
