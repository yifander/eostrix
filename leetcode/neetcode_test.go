package leetcode

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// TestCheckNeetcodePremiumStatus checks all NeetcodeAll slugs for isPaidOnly flag
func TestCheckNeetcodePremiumStatus(t *testing.T) {
	if len(NeetcodeAll) == 0 {
		t.Fatal("NeetcodeAll is empty")
	}

	var (
		premiumSlugs []string
		freeSlugs    []string
		errors       []string
	)

	client := &http.Client{Timeout: 10 * time.Second}

	for i, slug := range NeetcodeAll {
		t.Logf("Checking %d/%d: %s", i+1, len(NeetcodeAll), slug)

		isPaid, err := checkIsPaidOnly(client, slug)
		if err != nil {
			t.Logf("  Error checking %s: %v", slug, err)
			errors = append(errors, fmt.Sprintf("%s: %v", slug, err))
			continue
		}

		if isPaid {
			premiumSlugs = append(premiumSlugs, slug)
			t.Logf("PREMIUM")
		} else {
			freeSlugs = append(freeSlugs, slug)
			t.Logf("Free")
		}

		time.Sleep(100 * time.Millisecond)
	}

	t.Logf("\n========== RESULTS ==========")
	t.Logf("Total checked: %d", len(NeetcodeAll))
	t.Logf("Free: %d", len(freeSlugs))
	t.Logf("Premium: %d", len(premiumSlugs))
	t.Logf("Errors: %d", len(errors))

	if len(premiumSlugs) > 0 {
		t.Logf("\nPREMIUM SLUGS:")
		for _, slug := range premiumSlugs {
			t.Logf("  - %s", slug)
		}
	}

	if len(errors) > 0 {
		t.Logf("\nERRORS:")
		for _, err := range errors {
			t.Logf("  - %s", err)
		}
	}

	fmt.Printf("\n========== NEETCODE PREMIUM CHECK ==========\n")
	fmt.Printf("Total: %d\n", len(NeetcodeAll))
	fmt.Printf("Free: %d (%.1f%%)\n", len(freeSlugs), float64(len(freeSlugs))/float64(len(NeetcodeAll))*100)
	fmt.Printf("Premium: %d (%.1f%%)\n", len(premiumSlugs), float64(len(premiumSlugs))/float64(len(NeetcodeAll))*100)
	fmt.Printf("Errors: %d\n", len(errors))

	if len(premiumSlugs) > 0 {
		fmt.Printf("\nPremium slugs:\n")
		for _, slug := range premiumSlugs {
			fmt.Printf("  %s\n", slug)
		}
	}
}

// checkIsPaidOnly queries LeetCode API for a single problem's premium status
func checkIsPaidOnly(client *http.Client, slug string) (bool, error) {
	url := "https://leetcode.com/graphql"

	query := fmt.Sprintf(`{
		"query": "query questionData($titleSlug: String!) { question(titleSlug: $titleSlug) { titleSlug isPaidOnly } }",
		"variables": {"titleSlug": "%s"}
	}`, slug)

	req, err := http.NewRequest("POST", url, strings.NewReader(query))
	if err != nil {
		return false, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Eostrix-Test)")
	req.Header.Set("Referer", "https://leetcode.com/")
	req.Header.Set("Origin", "https://leetcode.com")

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	if len(body) > 0 && body[0] == '<' {
		return false, fmt.Errorf("HTML response (blocked)")
	}

	var result struct {
		Data struct {
			Question struct {
				TitleSlug  string `json:"titleSlug"`
				IsPaidOnly bool   `json:"isPaidOnly"`
			} `json:"question"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return false, err
	}

	if len(result.Errors) > 0 {
		return false, fmt.Errorf("graphql error: %s", result.Errors[0].Message)
	}

	if result.Data.Question.TitleSlug == "" {
		return false, fmt.Errorf("problem not found")
	}

	return result.Data.Question.IsPaidOnly, nil
}
