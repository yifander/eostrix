package commands

import (
	"eostrix/leetcode"
	"eostrix/utils"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const curatedPageSize = 10

// HandleCuratedCommand shows top aggregated problems, optionally filtered by difficulty
// Now supports pagination for browsing beyond the initial page
func HandleCuratedCommand(s *discordgo.Session, i *discordgo.InteractionCreate, store *leetcode.ProblemStore) {
	data := i.ApplicationCommandData()

	limit := 40
	difficulty := "all"

	for _, opt := range data.Options {
		switch opt.Name {
		case "difficulty":
			diff := strings.ToLower(opt.StringValue())
			switch diff {
			case "easy", "medium", "hard", "all":
				difficulty = diff
			default:
				utils.ResponseError(s, i, "Invalid difficulty. Choose: easy, medium, hard, or all")
				return
			}
		}
	}

	top := store.TopCuratedByDifficulty(limit, difficulty)
	if len(top) == 0 {
		utils.ResponseError(s, i, fmt.Sprintf("No curated problems found for difficulty '%s'", difficulty))
		return
	}

	leetcode.DefaultPagination.Store(
		i.Member.User.ID,
		"curated",
		toAnySliceCurated(top),
		curatedPageSize,
		map[string]string{
			"difficulty": difficulty,
			"total":      fmt.Sprintf("%d", len(top)),
		},
	)

	renderCuratedPage(s, i, 0, true)
}

func renderCuratedPage(s *discordgo.Session, i *discordgo.InteractionCreate, page int, isFirst bool) {
	pageData := leetcode.DefaultPagination.Get(i.Member.User.ID, "curated")
	if pageData == nil {
		utils.ResponseError(s, i, "Session expired. Please run /curated again")
		return
	}

	items := toCuratedSlice(pageData.Problems)

	start := page * pageData.PageSize
	end := min(start+pageData.PageSize, len(items))

	difficulty := pageData.Metadata["difficulty"]
	header := "**Most Common LeetCode Problems**"
	if difficulty != "all" {
		header = fmt.Sprintf("**Top %s LeetCode Problems**", capitalize(difficulty))
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s\n*Ranked by company breadth × frequency • Page %d/%d*\n\n",
		header, page+1, pageData.TotalPages))

	for idx, cp := range items[start:end] {
		rank := start + idx + 1
		companies := make([]string, 0, len(cp.Companies))
		for c := range cp.Companies {
			companies = append(companies, c)
		}

		sb.WriteString(fmt.Sprintf("%d. [%s](%s) • `%s`\n",
			rank, cp.Problem.Title, cp.Problem.Link, cp.Problem.Difficulty))
		sb.WriteString(fmt.Sprintf("Score: %.2f\n", cp.Score))

		if len(companies) > 6 {
			sb.WriteString(fmt.Sprintf("*Companies: %s (+%d more)*\n\n",
				strings.Join(companies[:6], ", "), len(companies)-6))
		} else {
			sb.WriteString(fmt.Sprintf("*Companies: %s*\n\n", strings.Join(companies, ", ")))
		}
	}

	components := leetcode.BuildPaginationButtons(
		"curated",
		page,
		pageData.TotalPages,
		difficulty,
	)

	if isFirst {
		if err := utils.ResponseComponents(s, i, sb.String(), components); err != nil {
			log.Printf("failed to send curated response: %v", err)
			utils.ResponseError(s, i, "An error occurred. Please try again.")
		}
	} else {
		utils.ResponseComponentsEdit(s, i, sb.String(), components)
	}
}

// HandleCuratedPageChange handles pagination button clicks for /curated
func HandleCuratedPageChange(s *discordgo.Session, i *discordgo.InteractionCreate, store *leetcode.ProblemStore) {
	if i.Member == nil || i.Member.User == nil {
		utils.ResponseError(s, i, "Could not identify user")
		return
	}

	userID := i.Member.User.ID

	cmd, action, _, _, err := leetcode.ParseButtonID(i.MessageComponentData().CustomID)
	if err != nil || cmd != "curated" {
		utils.ResponseError(s, i, "Invalid pagination state")
		return
	}

	pageData, err := leetcode.DefaultPagination.Navigate(userID, "curated", action)
	if err != nil {
		utils.ResponseError(s, i, "Session expired. Please run /curated again")
		return
	}

	renderCuratedPage(s, i, pageData.Page, false)
}

// toAnySliceCurated converts []*CuratedProblem to []any for pagination storage
func toAnySliceCurated(items []*leetcode.CuratedProblem) []any {
	out := make([]any, len(items))
	for i, item := range items {
		out[i] = item
	}
	return out
}

// toCuratedSlice converts []any back to []*CuratedProblem
func toCuratedSlice(items []any) []*leetcode.CuratedProblem {
	out := make([]*leetcode.CuratedProblem, len(items))
	for i, item := range items {
		if cp, ok := item.(*leetcode.CuratedProblem); ok {
			out[i] = cp
		}
	}
	return out
}

// capitalize returns title-cased string (avoids deprecated strings.Title)
func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
