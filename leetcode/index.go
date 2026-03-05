package leetcode

import (
	"encoding/csv"
	"eostrix/config"
	"eostrix/utils"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Problem struct {
	Company        string
	Difficulty     string
	Title          string
	TitleSlug      string
	Frequency      string
	AcceptanceRate string
	Link           string
	Topics         []string
	ProblemID      int
	IsNeetcode150  bool
}

var (
	AllProblems          []Problem
	ProblemsByCompany    map[string][]*Problem
	ProblemsByDifficulty map[string][]*Problem
	ProblemsByTopic      map[string][]*Problem
	topicSet             = map[string]struct{}{}
	ValidCompanies       []string
	ValidTopics          []string
)

func findSixMonthCSV(companyDir string) (string, error) {
	entries, err := os.ReadDir(companyDir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if !e.IsDir() && strings.EqualFold(e.Name(), "3. Six Months.csv") {
			return filepath.Join(companyDir, e.Name()), nil
		}
	}
	return "", fmt.Errorf("a six month csv not found in %s", companyDir)
}

func LoadAllProblems(rootDir string) ([]Problem, error) {
	AllProblems = make([]Problem, 0)

	ProblemsByCompany = make(map[string][]*Problem)
	ProblemsByDifficulty = make(map[string][]*Problem)
	ProblemsByTopic = make(map[string][]*Problem)

	ValidCompanies = nil
	ValidTopics = nil
	topicSet = make(map[string]struct{})

	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		companyName := entry.Name()
		companyDir := filepath.Join(rootDir, companyName)
		ValidCompanies = append(ValidCompanies, companyName)

		csvPath, err := findSixMonthCSV(companyDir)
		if err != nil {
			fmt.Printf("Skipping %s: %v\n", companyName, err)
			continue
		}

		f, err := os.Open(csvPath)
		if err != nil {
			fmt.Printf("Failed to open %s: %v\n", csvPath, err)
			continue
		}

		r := csv.NewReader(f)

		if _, err := r.Read(); err != nil {
			f.Close()
			fmt.Printf("Failed to read header of %s: %v\n", csvPath, err)
			continue
		}

		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Printf("error reading %s: %v\n", csvPath, err)
				continue
			}

			if len(record) < 5 {
				fmt.Printf("skipping bad row in %s: %v\n", csvPath, record)
				continue
			}

			topics := parseTopics(record[5:])
			link := record[4]
			problemID := parseProblemID(record, link)

			AllProblems = append(AllProblems, Problem{
				Company:        companyName,
				Difficulty:     record[0],
				Title:          record[1],
				Frequency:      record[2],
				AcceptanceRate: record[3],
				TitleSlug:      extractTitleSlug(record[4]),
				Link:           link,
				Topics:         topics,
				ProblemID:      problemID,
				IsNeetcode150:  IsNeetcode150(problemID),
			})

			pp := &AllProblems[len(AllProblems)-1]
			createIndexes(pp)
		}

		f.Close()
	}

	fmt.Printf("Loaded %d problems across %d companies\n", len(AllProblems), len(entries))
	fmt.Printf("Loaded %d topics across %d problems\n", len(ValidTopics), len(AllProblems))

	InitNeetcodeIndex(AllProblems)
	fmt.Printf("Indexed %d Neetcode 150 problems\n", len(GetAllNeetcodeProblems()))

	return AllProblems, nil
}

func createIndexes(p *Problem) {
	companyKey := strings.ToLower(p.Company)
	ProblemsByCompany[companyKey] = append(ProblemsByCompany[companyKey], p)

	diffKey := strings.ToLower(p.Difficulty)
	ProblemsByDifficulty[diffKey] = append(ProblemsByDifficulty[diffKey], p)

	for _, t := range p.Topics {
		key := strings.ToLower(t)
		ProblemsByTopic[key] = append(ProblemsByTopic[key], p)

		if _, exists := topicSet[key]; !exists {
			topicSet[key] = struct{}{}
			ValidTopics = append(ValidTopics, t)
		}
	}
}

func parseProblemID(record []string, link string) int {
	if len(record) > 6 {
		if id, err := strconv.Atoi(record[6]); err == nil {
			return id
		}
	}
	return extractIDFromLink(link)
}

func extractIDFromLink(link string) int {
	parts := strings.Split(link, "/")
	for i, part := range parts {
		if part == "problems" && i+1 < len(parts) {
			slug := parts[i+1]
			return lookupIDBySlug(slug)
		}
	}
	return 0
}

var slugToIDMap = map[string]int{
	"two-sum":         1,
	"add-two-numbers": 2,
	"longest-substring-without-repeating-characters": 3,
	"median-of-two-sorted-arrays":                    4,
	"longest-palindromic-substring":                  5,
}

func lookupIDBySlug(slug string) int {
	if id, ok := slugToIDMap[slug]; ok {
		return id
	}
	return 0
}

func parseTopics(columns []string) []string {
	var topics []string

	for _, col := range columns {
		for part := range strings.SplitSeq(col, ",") {
			topic := strings.TrimSpace(part)
			if topic != "" {
				topics = append(topics, topic)
			}
		}
	}

	return topics
}

func extractTitleSlug(link string) string {
	parts := strings.Split(link, "/")
	for i, part := range parts {
		if part == "problems" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func PostRandomProblem(session *discordgo.Session, difficulty string) {
	cfg := config.ParseConfig()

	var problem *Problem

	if difficulty != "" {
		diffKey := strings.ToLower(difficulty)
		candidates := ProblemsByDifficulty[diffKey]
		if len(candidates) == 0 {
			log.Printf("No problems found with difficulty: %s", difficulty)
			return
		}
		problem = candidates[rand.Intn(len(candidates))]
	} else {
		if len(AllProblems) == 0 {
			log.Printf("No problems loaded")
			return
		}
		problem = &AllProblems[rand.Intn(len(AllProblems))]
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("**Problem:** %s\n", problem.Title))
	builder.WriteString(fmt.Sprintf("**Difficulty:** %s\n", problem.Difficulty))
	builder.WriteString(fmt.Sprintf("**Topics:** %s\n", strings.Join(problem.Topics, ", ")))
	builder.WriteString(fmt.Sprintf("**Company:** %s\n", problem.Company))
	builder.WriteString(fmt.Sprintf("**Acceptance Rate:** %s\n", problem.AcceptanceRate))
	builder.WriteString(fmt.Sprintf("\n**Link:** %s\n", problem.Link))

	if problem.IsNeetcode150 {
		builder.WriteString("\n**Part of Neetcode 150**")
	}

	title := "Random LeetCode Problem"
	if difficulty != "" {
		title = fmt.Sprintf("Random %s LeetCode Problem", difficulty)
	}

	utils.SendMessageComplex(session, cfg.DefaultChannel, title, builder.String())
}
