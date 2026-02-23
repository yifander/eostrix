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

var CuratedProblems []*CuratedProblem

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

func BuildCuratedProblems() {
	index := make(map[string]*CuratedProblem)

	for company, problems := range ProblemsByCompany {
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

	CuratedProblems = make([]*CuratedProblem, 0, len(index))

	for _, p := range index {
		companyCount := len(p.Companies)
		avgFreq := p.FreqSum / float64(p.Appearances)

		p.Score = math.Log1p(float64(companyCount)) * avgFreq
		CuratedProblems = append(CuratedProblems, p)
	}

	sort.Slice(CuratedProblems, func(i, j int) bool {
		return CuratedProblems[i].Score > CuratedProblems[j].Score
	})
}
