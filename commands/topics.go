package commands

import (
	"eostrix/leetcode"
	"eostrix/utils"
	"fmt"
	"strconv"
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

	renderTopicsPage(s, i, topic, difficulty, filtered, 0, true)
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

	if userInput == "" {
		if userInput == "" {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionApplicationCommandAutocompleteResult,
				Data: &discordgo.InteractionResponseData{
					Choices: []*discordgo.ApplicationCommandOptionChoice{},
				},
			})
			return
		}
	}

	allTopics := store.Topics()
	suggestions := make([]*discordgo.ApplicationCommandOptionChoice, 0, 25)

	for _, vt := range allTopics {
		if strings.Contains(strings.ToLower(vt), userInput) {
			suggestions = append(suggestions, &discordgo.ApplicationCommandOptionChoice{
				Name:  vt,
				Value: vt,
			})
		}
		if len(suggestions) == 25 {
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

func renderTopicsPage(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	topic, difficulty string,
	problems []*leetcode.Problem,
	page int,
	first bool) {
	totalPages := (len(problems) + topicsPageSize - 1) / topicsPageSize

	start := page * topicsPageSize
	end := start + topicsPageSize

	if start >= len(problems) {
		start = 0
		page = 0
	}
	if end > len(problems) {
		end = len(problems)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		"**Topic:** %s\n**Difficulty:** %s\n**Page:** %d / %d\n\n",
		topic,
		difficulty,
		page+1,
		totalPages,
	))

	for _, p := range problems[start:end] {
		sb.WriteString(fmt.Sprintf(
			"• %s (%s, %s Frequency)\n%s\n\n",
			p.Title, p.Difficulty, p.Frequency, p.Link,
		))
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Prev",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("topics_prev:%s:%s:%d", topic, difficulty, page),
					Disabled: page == 0,
				},
				discordgo.Button{
					Label:    "Next",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("topics_next:%s:%s:%d", topic, difficulty, page),
					Disabled: page >= totalPages-1,
				},
			},
		},
	}

	if first {
		utils.ResponseComponents(s, i, sb.String(), components)
	} else {
		utils.ResponseComponentsEdit(s, i, sb.String(), components)
	}
}

func HandleTopicsPageChange(s *discordgo.Session, i *discordgo.InteractionCreate, store *leetcode.ProblemStore) {
	parts := strings.Split(i.MessageComponentData().CustomID, ":")
	if len(parts) != 4 {
		return
	}

	action := parts[0]
	topic := parts[1]
	difficulty := parts[2]
	page, _ := strconv.Atoi(parts[3])

	if action == "topics_next" {
		page++
	} else {
		page--
	}

	problems := store.ByTopic(topic)
	filtered := filterByDifficulty(problems, difficulty)

	if len(filtered) == 0 {
		utils.ResponseError(s, i, "No problems match those filters anymore")
		return
	}

	renderTopicsPage(s, i, topic, difficulty, problems, page, false)
}

func filterByDifficulty(problems []*leetcode.Problem, diff string) []*leetcode.Problem {
	var out []*leetcode.Problem
	for _, p := range problems {
		if strings.EqualFold(p.Difficulty, diff) {
			out = append(out, p)
		}
	}
	return out
}
