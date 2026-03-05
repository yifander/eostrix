package leetcode

import (
	"eostrix/config"
	"eostrix/utils"
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var Neetcode150ProblemIDs = map[int]struct{}{
	// Arrays & Hashing (9)
	217: {}, 242: {}, 1: {}, 49: {}, 347: {}, 238: {}, 36: {}, 271: {}, 128: {},
	// Two Pointers (5)
	125: {}, 167: {}, 15: {}, 11: {}, 42: {},
	// Sliding Window (6)
	121: {}, 3: {}, 424: {}, 567: {}, 76: {}, 239: {},
	// Stack (7)
	20: {}, 155: {}, 150: {}, 22: {}, 739: {}, 853: {}, 84: {},
	// Binary Search (7)
	704: {}, 74: {}, 875: {}, 153: {}, 33: {}, 981: {}, 4: {},
	// Linked List (11)
	206: {}, 21: {}, 143: {}, 19: {}, 138: {}, 2: {}, 141: {}, 287: {}, 146: {}, 23: {}, 25: {},
	// Trees (15)
	226: {}, 104: {}, 543: {}, 110: {}, 100: {}, 572: {}, 235: {}, 102: {}, 199: {}, 1448: {}, 98: {}, 230: {}, 105: {}, 124: {}, 297: {},
	// Heap/Priority Queue (6)
	703: {}, 1046: {}, 973: {}, 621: {}, 355: {}, 295: {},
	// Backtracking (9)
	78: {}, 39: {}, 46: {}, 90: {}, 40: {}, 79: {}, 131: {}, 17: {}, 51: {},
	// Tries (3)
	208: {}, 211: {}, 212: {},
	// Graphs (13)
	200: {}, 133: {}, 695: {}, 417: {}, 130: {}, 994: {}, 286: {}, 207: {}, 210: {}, 684: {}, 547: {}, 261: {}, 127: {},
	// Advanced Graphs (6)
	332: {}, 1584: {}, 743: {}, 778: {}, 269: {}, 787: {},
	// 1-D DP (12)
	70: {}, 746: {}, 198: {}, 213: {}, 5: {}, 647: {}, 91: {}, 322: {}, 152: {}, 139: {}, 300: {}, 416: {},
	// 2-D DP (11)
	62: {}, 1143: {}, 309: {}, 518: {}, 494: {}, 97: {}, 72: {}, 115: {}, 312: {}, 10: {},
	// Greedy (7)
	53: {}, 55: {}, 45: {}, 134: {}, 846: {}, 1899: {}, 763: {},
	// Intervals (6)
	57: {}, 56: {}, 435: {}, 252: {}, 253: {}, 1851: {},
	// Math & Geometry (8)
	48: {}, 54: {}, 73: {}, 202: {}, 66: {}, 50: {}, 43: {}, 2013: {},
	// Bit Manipulation (7)
	136: {}, 191: {}, 338: {}, 190: {}, 268: {}, 371: {}, 7: {},
}

type NeetcodeIndex struct {
	AllProblems  []*Problem
	ByDifficulty map[string][]*Problem
}

var neetcodeIndex *NeetcodeIndex

func InitNeetcodeIndex(allProblems []Problem) {
	neetcodeIndex = &NeetcodeIndex{
		AllProblems:  make([]*Problem, 0, 150),
		ByDifficulty: make(map[string][]*Problem),
	}

	for i := range allProblems {
		p := &allProblems[i]
		if !p.IsNeetcode150 {
			continue
		}

		neetcodeIndex.AllProblems = append(neetcodeIndex.AllProblems, p)

		diffKey := strings.ToLower(p.Difficulty)
		neetcodeIndex.ByDifficulty[diffKey] = append(neetcodeIndex.ByDifficulty[diffKey], p)
	}
}

func IsNeetcode150(id int) bool {
	_, ok := Neetcode150ProblemIDs[id]
	return ok
}

func GetRandomNeetcode150(difficulty string) *Problem {
	if neetcodeIndex == nil || len(neetcodeIndex.AllProblems) == 0 {
		return nil
	}

	var pool []*Problem
	if difficulty == "" {
		pool = neetcodeIndex.AllProblems
	} else {
		pool = neetcodeIndex.ByDifficulty[strings.ToLower(difficulty)]
	}

	if len(pool) == 0 {
		return nil
	}

	return pool[rand.Intn(len(pool))]
}

func GetAllNeetcodeProblems() []*Problem {
	if neetcodeIndex == nil {
		return nil
	}
	return neetcodeIndex.AllProblems
}

func GetNeetcodeStats() map[string]int {
	if neetcodeIndex == nil {
		return map[string]int{"total": 0}
	}

	stats := map[string]int{
		"total": len(neetcodeIndex.AllProblems),
	}

	for diff, probs := range neetcodeIndex.ByDifficulty {
		stats[diff] = len(probs)
	}

	return stats
}

func PostRandomNeetcode(session *discordgo.Session, difficulty string) {
	cfg := config.ParseConfig()

	problem := GetRandomNeetcode150(difficulty)
	if problem == nil {
		log.Printf("No Neetcode 150 problems available (difficulty: %s)", difficulty)
		return
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("**Problem:** %s\n", problem.Title))
	builder.WriteString(fmt.Sprintf("**Difficulty:** %s\n", problem.Difficulty))
	builder.WriteString(fmt.Sprintf("**Topics:** %s\n", strings.Join(problem.Topics, ", ")))
	builder.WriteString(fmt.Sprintf("**Company:** %s\n", problem.Company))
	builder.WriteString(fmt.Sprintf("**Acceptance Rate:** %s\n", problem.AcceptanceRate))
	builder.WriteString(fmt.Sprintf("\n**Link:** https://leetcode.com/problems/%s/\n", problem.TitleSlug))

	ping := fmt.Sprintf("<@&%s> ", cfg.LeetcodeRoleID)

	title := "Random Neetcode 150 Problem"
	if difficulty != "" {
		title = fmt.Sprintf("Random %s Neetcode 150 Problem", difficulty)
	}

	utils.SendPingMessageComplex(session, cfg.DefaultChannel, title, ping, builder.String())
}
