package commands

import (
	"eostrix/leetcode"
	"eostrix/utils"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// HandleCuratedCommand shows top aggregated problems, optionally filtered by difficulty
func HandleCuratedCommand(s *discordgo.Session, i *discordgo.InteractionCreate, store *leetcode.ProblemStore) {
	data := i.ApplicationCommandData()
	limit := 10
	difficulty := "all"

	for _, opt := range data.Options {
		switch opt.Name {
		case "limit":
			val := int(opt.IntValue())
			if val > 0 {
				limit = val
				if limit > 25 {
					limit = 25
				}
			}
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

	var sb strings.Builder

	header := "**Most Common LeetCode Problems**"
	if difficulty != "all" {
		header = fmt.Sprintf("**Top %s LeetCode Problems**", capitalize(difficulty))
	}

	sb.WriteString(fmt.Sprintf("%s\n*Ranked by company breadth × frequency*\n\n", header))

	for rank, cp := range top {
		companies := make([]string, 0, len(cp.Companies))
		for c := range cp.Companies {
			companies = append(companies, c)
		}

		sb.WriteString(fmt.Sprintf("%d. [%s](%s) - `%s`\n",
			rank+1, cp.Problem.Title, cp.Problem.Link, cp.Problem.Difficulty))
		sb.WriteString(fmt.Sprintf("%d companies - score: %.2f\n",
			len(cp.Companies), cp.Score))

		if len(companies) > 6 {
			sb.WriteString(fmt.Sprintf("*Seen at: %s (+%d more)*\n\n",
				strings.Join(companies[:6], ", "), len(companies)-6))
		} else {
			sb.WriteString(fmt.Sprintf("   *Seen at: %s*\n\n", strings.Join(companies, ", ")))
		}
	}

	utils.Response(s, i, "Top Curated Problems", sb.String())
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
