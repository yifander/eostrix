package commands

import (
	"eostrix/leetcode"
	"eostrix/utils"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const companyPageSize = 10

// HandleCompanyCommand sends the first page of company problems
func HandleCompanyCommand(s *discordgo.Session, i *discordgo.InteractionCreate, store *leetcode.ProblemStore) {
	data := i.ApplicationCommandData()
	companyName := data.Options[0].StringValue()
	difficulty := data.Options[1].StringValue()
	problems := store.ByCompany(companyName)

	switch difficulty {
	case "easy", "medium", "hard", "all":
	default:
		utils.ResponseError(s, i, "Invalid difficulty provided.")
		return
	}

	companyProblems := store.ByCompany(companyName)
	if len(problems) == 0 {
		utils.ResponseError(s, i, fmt.Sprintf("No problems found for company '%s'", companyName))
		return
	}

	// filter by difficulty
	var filtered []*leetcode.Problem
	for _, p := range companyProblems {
		if difficulty == "all" || strings.EqualFold(p.Difficulty, difficulty) {
			filtered = append(filtered, p)
		}
	}

	if len(filtered) == 0 {
		utils.ResponseError(s, i, "No matching problems for that difficulty.")
		return
	}

	leetcode.DefaultPagination.Store(
		i.Member.User.ID,
		"company",
		toAnySlice(filtered),
		companyPageSize,
		map[string]string{
			"company":    companyName,
			"difficulty": difficulty,
		},
	)

	renderCompanyPage(s, i, 0, true)
}

// CompanyAutocomplete provides suggestions as user types
func CompanyAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate, store *leetcode.ProblemStore) {
	data := i.ApplicationCommandData()
	if len(data.Options) == 0 {
		return
	}

	userInput := strings.ToLower(data.Options[0].StringValue())
	var suggestions []*discordgo.ApplicationCommandOptionChoice

	for _, company := range store.Companies() {
		if strings.Contains(strings.ToLower(company), userInput) {
			suggestions = append(suggestions, &discordgo.ApplicationCommandOptionChoice{
				Name:  company,
				Value: company,
			})
		}
		if len(suggestions) >= 25 {
			break
		}
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: suggestions,
		},
	})
}

// renderCompanyPage renders a page of results (initial or edit)
func renderCompanyPage(s *discordgo.Session, i *discordgo.InteractionCreate, page int, isFirst bool) {
	pageData := leetcode.DefaultPagination.Get(i.Member.User.ID, "company")
	if pageData == nil {
		utils.ResponseError(s, i, "Session expired. Please run /company again")
		return
	}

	problems := toProblemSlice(pageData.Problems)
	start := page * pageData.PageSize
	end := min(start+pageData.PageSize, len(problems))

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		"**Company: %s**\n**Difficulty: %s**\n**Page %d / %d**\n\n",
		pageData.Metadata["company"],
		pageData.Metadata["difficulty"],
		page+1,
		pageData.TotalPages,
	))

	for _, p := range problems[start:end] {
		sb.WriteString(fmt.Sprintf("• %s (%s) %s freq\n%s\n\n",
			p.Title, p.Difficulty, p.Frequency, p.Link))
	}

	components := leetcode.BuildPaginationButtons(
		"company",
		page,
		pageData.TotalPages,
		pageData.Metadata["company"],
		pageData.Metadata["difficulty"],
	)

	if isFirst {
		if err := utils.ResponseComponents(s, i, sb.String(), components); err != nil {
			log.Printf("failed to send initial response: %v", err)
			utils.ResponseError(s, i, "An error occurred. Please try again.")
		}
	} else {
		utils.ResponseComponentsEdit(s, i, sb.String(), components)
	}
}

// HandleCompanyPageChange handles pagination button clicks
func HandleCompanyPageChange(s *discordgo.Session, i *discordgo.InteractionCreate, store *leetcode.ProblemStore) {
	if i.Member == nil || i.Member.User == nil {
		utils.ResponseError(s, i, "Could not identify user")
		return
	}

	userID := i.Member.User.ID

	cmd, action, _, _, err := leetcode.ParseButtonID(i.MessageComponentData().CustomID)
	if err != nil || cmd != "company" {
		utils.ResponseError(s, i, "Invalid pagination state")
		return
	}

	pageData, err := leetcode.DefaultPagination.Navigate(userID, "company", action)
	if err != nil {
		utils.ResponseError(s, i, "Session expired. Please run /company again")
		return
	}

	renderCompanyPage(s, i, pageData.Page, false)
}

func toAnySlice(problems []*leetcode.Problem) []any {
	out := make([]any, len(problems))
	for i, p := range problems {
		out[i] = p
	}
	return out
}

func toProblemSlice(items []any) []*leetcode.Problem {
	out := make([]*leetcode.Problem, len(items))
	for i, item := range items {
		if p, ok := item.(*leetcode.Problem); ok {
			out[i] = p
		}
	}
	return out
}
