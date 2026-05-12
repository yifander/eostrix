package leetcode

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
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

type ProblemStore struct {
	mu sync.RWMutex

	allProblems    []Problem
	byCompany      map[string][]*Problem
	byDifficulty   map[string][]*Problem
	byTopic        map[string][]*Problem
	validCompanies []string
	validTopics    []string
	topicSet       map[string]struct{}
	lastLoadedTime time.Time
	sourceChecksum string

	// see curated.go for more on this list type
	curatedMu       sync.RWMutex
	curatedProblems []*CuratedProblem
}

func NewProblemStore() *ProblemStore {
	return &ProblemStore{
		byCompany:    make(map[string][]*Problem),
		byDifficulty: make(map[string][]*Problem),
		byTopic:      make(map[string][]*Problem),
		topicSet:     make(map[string]struct{}),
	}
}

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

func readAllCSVs(rootDir string) ([]Problem, []string, error) {
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read root dir %s: %w", rootDir, err)
	}

	var allProblems []Problem
	var companies []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		companyName := entry.Name()
		companyDir := filepath.Join(rootDir, companyName)
		companies = append(companies, companyName)

		csvPath, err := findSixMonthCSV(companyDir)
		if err != nil {
			log.Printf("Skipping %s: %v", companyName, err)
			continue
		}

		f, err := os.Open(csvPath)
		if err != nil {
			log.Printf("Failed to open %s: %v", csvPath, err)
			continue
		}

		problems, err := parseCompanyCSV(f, companyName)
		f.Close()
		if err != nil {
			log.Printf("Error parsing %s: %v", csvPath, err)
			continue
		}

		allProblems = append(allProblems, problems...)
	}

	return allProblems, companies, nil
}

func parseCompanyCSV(f *os.File, companyName string) ([]Problem, error) {
	r := csv.NewReader(f)

	if _, err := r.Read(); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	var problems []Problem
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading row: %v", err)
			continue
		}

		if len(record) < 5 {
			continue
		}

		problems = append(problems, Problem{
			Company:        companyName,
			Difficulty:     record[0],
			Title:          record[1],
			Frequency:      record[2],
			AcceptanceRate: record[3],
			Link:           record[4],
			Topics:         parseTopics(record[5:]),
		})
	}

	return problems, nil
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

// Load parses all company CSVs and builds in-memory indexes.
// Safe to call multiple times; replaces existing data atomically.
func (s *ProblemStore) Load(rootDir string) error {
	tempProblems, companies, err := readAllCSVs(rootDir)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.allProblems = tempProblems
	s.validCompanies = companies
	s.byCompany = make(map[string][]*Problem)
	s.byDifficulty = make(map[string][]*Problem)
	s.byTopic = make(map[string][]*Problem)
	s.topicSet = make(map[string]struct{})
	s.validTopics = s.validTopics[:0] // reset but keep capacity

	for i := range s.allProblems {
		s.createIndexes(&s.allProblems[i])
	}

	log.Printf("Loaded %d problems across %d companies", len(s.allProblems), len(s.validCompanies))
	log.Printf("Indexed %d unique topics", len(s.validTopics))
	return nil
}

// only call while holding s.mu.Lock()
func (s *ProblemStore) createIndexes(p *Problem) {
	companyKey := strings.ToLower(p.Company)
	s.byCompany[companyKey] = append(s.byCompany[companyKey], p)

	diffKey := strings.ToLower(p.Difficulty)
	s.byDifficulty[diffKey] = append(s.byDifficulty[diffKey], p)

	for _, t := range p.Topics {
		key := strings.ToLower(t)
		s.byTopic[key] = append(s.byTopic[key], p)

		if _, exists := s.topicSet[key]; !exists {
			s.topicSet[key] = struct{}{}
			s.validTopics = append(s.validTopics, t)
		}
	}
}

// accessors below

func (s *ProblemStore) All() []Problem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Problem, len(s.allProblems))
	copy(out, s.allProblems)

	return out
}

func (s *ProblemStore) ByCompany(company string) []*Problem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.byCompany[strings.ToLower(company)]
}

func (s *ProblemStore) ByTopic(topic string) []*Problem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.byTopic[strings.ToLower(topic)]
}

func (s *ProblemStore) ByDifficulty(difficulty string) []*Problem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.byDifficulty[strings.ToLower(difficulty)]
}

func (s *ProblemStore) Companies() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]string, len(s.validCompanies))
	copy(out, s.validCompanies)

	return out
}

func (s *ProblemStore) Topics() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]string, len(s.validTopics))
	copy(out, s.validTopics)

	return out
}

func (s *ProblemStore) TopCuratedByDifficulty(limit int, difficulty string) []*CuratedProblem {
	s.curatedMu.RLock()
	defer s.curatedMu.RUnlock()

	var filtered []*CuratedProblem
	for _, cp := range s.curatedProblems {
		if difficulty == "all" || strings.EqualFold(cp.Problem.Difficulty, difficulty) {
			filtered = append(filtered, cp)
		}
	}

	if limit > len(filtered) {
		limit = len(filtered)
	}
	return filtered[:limit]
}
