package leetcode

import (
	"math"
	"sort"
	"strings"
)

type CuratedProblem struct {
	Problem     *Problem
	Companies   map[string]struct{}
	Appearances int
	FreqSum     float64
	Score       float64
}

func normalizeFrequency(freq string) float64 {
	switch strings.ToLower(strings.TrimSpace(freq)) {
	case "high":
		return 1.0
	case "medium":
		return 0.6
	case "low":
		return 0.3
	default:
		return 0.5
	}
}

func (s *ProblemStore) BuildCuratedProblems() {
	s.mu.Lock()
	defer s.mu.Unlock()

	index := make(map[string]*CuratedProblem)

	for company, problems := range s.byCompany {
		for _, p := range problems {
			key := strings.ToLower(p.Title)

			entry, ok := index[key]
			if !ok {
				entry = &CuratedProblem{
					Problem:   p,
					Companies: make(map[string]struct{}),
				}
				index[key] = entry
			}

			entry.Companies[company] = struct{}{}
			entry.FreqSum += normalizeFrequency(p.Frequency)
			entry.Appearances++
		}
	}

	s.curatedProblems = make([]*CuratedProblem, 0, len(index))
	for _, cp := range index {
		companyCount := len(cp.Companies)
		avgFreq := cp.FreqSum / float64(cp.Appearances)
		// score: log-scaled company breadth * average frequency weight
		cp.Score = math.Log1p(float64(companyCount)) * avgFreq
		s.curatedProblems = append(s.curatedProblems, cp)
	}

	sort.Slice(s.curatedProblems, func(i, j int) bool {
		return s.curatedProblems[i].Score > s.curatedProblems[j].Score
	})
}

func (s *ProblemStore) TopCurated(n int) []*CuratedProblem {
	s.curatedMu.RLock()
	defer s.curatedMu.RUnlock()

	if n > len(s.curatedProblems) {
		n = len(s.curatedProblems)
	}

	return s.curatedProblems[:n]
}

func (s *ProblemStore) CuratedCount() int {
	s.curatedMu.RLock()
	defer s.curatedMu.RUnlock()

	return len(s.curatedProblems)
}
