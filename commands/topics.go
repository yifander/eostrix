package commands

import (
	"eostrix/leetcode"
	"eostrix/utils"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// syntax for command is /topics <category> <difficulty>
const topicsPageSize = 10

func HandleTopicsCommand(s *discordgo.Session, i *discordgo.InteractionCreate, store *leetcode.ProblemStore) {
	data := i.ApplicationCommandData()
	opts := data.Options

	topic := strings.ToLower(opts[0].StringValue())
	difficulty := strings.ToLower(opts[1].StringValue())

	problems := store.ByTopic(topic)
	if len(problems) == 0 {
		utils.ResponseError(s, i, fmt.Sprintf("No problems found for topic %s", topic))
		return
	}

	// filter by difficulty
	var filtered []*leetcode.Problem
	for _, p := range problems {
		if strings.EqualFold(p.Difficulty, difficulty) {
			filtered = append(filtered, p)
		}
	}

	if len(filtered) == 0 {
		utils.ResponseError(s, i, "No matching problems for that difficulty.")
		return
	}

	leetcode.DefaultPagination.Store(
		i.Member.User.ID,
		"topics",
		toAnySlice(filtered),
		topicsPageSize,
		map[string]string{
			"topic":      topic,
			"difficulty": difficulty,
		},
	)

	renderTopicsPage(s, i, 0, true)
}

func TopicsAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate, store *leetcode.ProblemStore) {
	data := i.ApplicationCommandData()

	var userInput string
	for _, opt := range data.Options {
		if opt.Focused {
			userInput = strings.ToLower(opt.StringValue())
			break
		}
	}

	allTopics := store.Topics()
	suggestions := make([]*discordgo.ApplicationCommandOptionChoice, 0, 25)

	for _, t := range allTopics {
		if strings.Contains(strings.ToLower(t), userInput) {
			suggestions = append(suggestions, &discordgo.ApplicationCommandOptionChoice{
				Name:  t,
				Value: t,
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

func renderTopicsPage(s *discordgo.Session, i *discordgo.InteractionCreate, page int, isFirst bool) {
	pageData := leetcode.DefaultPagination.Get(i.Member.User.ID, "topics")
	if pageData == nil {
		utils.ResponseError(s, i, "Session expired. Please run /topics again")
		return
	}

	problems := toProblemSlice(pageData.Problems)

	start := page * pageData.PageSize
	end := min(start+pageData.PageSize, len(problems))

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		"**Topic:** %s\n**Difficulty:** %s\n**Page:** %d / %d\n\n",
		strings.Title(pageData.Metadata["topic"]),
		strings.Title(pageData.Metadata["difficulty"]),
		page+1,
		pageData.TotalPages,
	))

	for _, p := range problems[start:end] {
		sb.WriteString(fmt.Sprintf(
			"• %s (%s, %s Frequency)\n%s\n\n",
			p.Title, p.Difficulty, p.Frequency, p.Link,
		))
	}

	components := leetcode.BuildPaginationButtons(
		"topics",
		page,
		pageData.TotalPages,
		pageData.Metadata["topic"],
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

// HandleTopicsPageChange handles pagination button clicks
func HandleTopicsPageChange(s *discordgo.Session, i *discordgo.InteractionCreate, store *leetcode.ProblemStore) {
	if i.Member == nil || i.Member.User == nil {
		utils.ResponseError(s, i, "Could not identify user")
		return
	}

	userID := i.Member.User.ID

	cmd, action, _, _, err := leetcode.ParseButtonID(i.MessageComponentData().CustomID)
	if err != nil || cmd != "topics" {
		utils.ResponseError(s, i, "Invalid pagination state")
		return
	}

	pageData, err := leetcode.DefaultPagination.Navigate(userID, "topics", action)
	if err != nil {
		utils.ResponseError(s, i, "Session expired. Please run /topics again")
		return
	}

	renderTopicsPage(s, i, pageData.Page, false)
}
