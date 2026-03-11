package leetcode

import (
	"encoding/json"
	"eostrix/config"
	"eostrix/utils"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type problemResponse struct {
	Data struct {
		Question struct {
			QuestionID       string `json:"questionId"`
			Title            string `json:"title"`
			TitleSlug        string `json:"titleSlug"`
			Difficulty       string `json:"difficulty"`
			Content          string `json:"content"`
			ExampleTestcases string `json:"exampleTestcases"`
			CodeSnippets     []struct {
				Lang     string `json:"lang"`
				LangSlug string `json:"langSlug"`
				Code     string `json:"code"`
			} `json:"codeSnippets"`
			Hints          []string `json:"hints"`
			SampleTestCase string   `json:"sampleTestCase"`
			MetaData       string   `json:"metaData"`
		} `json:"question"`
	} `json:"data"`
}

type LeetcodeProblem struct {
	QuestionID   string
	Title        string
	TitleSlug    string
	Difficulty   string
	Content      string
	Examples     string
	CodeSnippets map[string]string
	Hints        []string
}

func GetProblemBySlug(titleSlug string) (LeetcodeProblem, error) {
	url := "https://leetcode.com/graphql"

	query := fmt.Sprintf(`{
		"query": "query questionData($titleSlug: String!) { question(titleSlug: $titleSlug) { questionId title titleSlug difficulty content exampleTestcases codeSnippets { lang langSlug code } hints sampleTestCase metaData } }",
		"variables": {"titleSlug": "%s"}
	}`, titleSlug)

	req, err := http.NewRequest("POST", url, strings.NewReader(query))
	if err != nil {
		return LeetcodeProblem{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return LeetcodeProblem{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return LeetcodeProblem{}, err
	}

	var pr problemResponse
	if err := json.Unmarshal(body, &pr); err != nil {
		return LeetcodeProblem{}, err
	}

	snippets := make(map[string]string)
	for _, cs := range pr.Data.Question.CodeSnippets {
		snippets[cs.LangSlug] = cs.Code
	}

	return LeetcodeProblem{
		QuestionID:   pr.Data.Question.QuestionID,
		Title:        pr.Data.Question.Title,
		TitleSlug:    pr.Data.Question.TitleSlug,
		Difficulty:   pr.Data.Question.Difficulty,
		Content:      pr.Data.Question.Content,
		Examples:     pr.Data.Question.ExampleTestcases,
		CodeSnippets: snippets,
		Hints:        pr.Data.Question.Hints,
	}, nil
}

func PostDailyChallenge(session *discordgo.Session, slug string) {
	var builder strings.Builder
	cfg := config.ParseConfig()

	daily, err := GetProblemBySlug(slug)
	if err != nil {
		log.Printf("Error fetching challenge: %v", err)
	}

	builder.WriteString(fmt.Sprintf("**Problem Name:** %s\n", daily.Title))
	builder.WriteString(fmt.Sprintf("**Difficulty:** %s\n", daily.Difficulty))

	link := fmt.Sprintf("https://leetcode.com/problems/%s/", daily.TitleSlug)
	builder.WriteString(fmt.Sprintf("**Link:** \n%s\n", link))

	if IsNeetcode150Slug(daily.TitleSlug) {
		builder.WriteString("\n**Part of Neetcode 150**")
	}

	ping := fmt.Sprintf("<@&%s> ", cfg.LeetcodeRoleID)

	utils.SendPingMessageComplex(session, cfg.DefaultChannel, "Daily Neetcode All Challenge", ping, builder.String())
}
