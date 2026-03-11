package leetcode

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// represents an individual leetcode problem found within the csv file
type Problem struct {
	Company        string
	Difficulty     string
	Title          string
	Frequency      string
	AcceptanceRate string
	Link           string
	Topics         []string
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

// search each company folder for the correct six month cvs file
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

// load the leetcode problems from the six month cvs files to the CompanyProblem struct
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

			AllProblems = append(AllProblems, Problem{
				Company:        companyName,
				Difficulty:     record[0],
				Title:          record[1],
				Frequency:      record[2],
				AcceptanceRate: record[3],
				Link:           record[4],
				Topics:         topics,
			})

			pp := &AllProblems[len(AllProblems)-1]

			createIndexes(pp)
		}

		f.Close()
	}

	fmt.Printf("Loaded %d problems across %d companies\n", len(AllProblems), len(entries))
	fmt.Printf("Loaded %d topics accross %d problems\n", len(ValidTopics), len(AllProblems))

	return AllProblems, nil
}

// index for company, difficulty, and topics
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

func parseTopics(columns []string) []string {
	var topics []string

	for _, col := range columns {
		// im using an old go version (1.23) so split is preferable to splitseq here
		for _, part := range strings.Split(col, ",") {
			topic := strings.TrimSpace(part)
			if topic != "" {
				topics = append(topics, topic)
			}
		}
	}

	return topics
}
