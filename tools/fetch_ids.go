package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type LCProblemList struct {
	Pairs []struct {
		Stat struct {
			ID    int    `json:"question_id"`
			Title string `json:"question__title"`
			Slug  string `json:"question__title_slug"`
		} `json:"stat"`
	} `json:"stat_status_pairs"`
}

func main() {
	resp, err := http.Get("https://leetcode.com/api/problems/algorithms/")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var list LCProblemList
	json.Unmarshal(body, &list)

	m := make(map[string]string)
	for _, p := range list.Pairs {
		if p.Stat.Slug != "" {
			m[fmt.Sprintf("%d", p.Stat.ID)] = p.Stat.Slug
		}
	}

	out, _ := json.MarshalIndent(m, "", "  ")
	os.WriteFile("leetcode_ids.json", out, 0644)
	log.Printf("Generated %d ID→slug mappings", len(m))
}
