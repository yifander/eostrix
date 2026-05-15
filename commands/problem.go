package commands

import (
	"context"
	"eostrix/leetcode"
	"eostrix/utils"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// HandleProblemCommand fetches a LeetCode problem by slug (API-only)
// Usage: /problem <slug>
func HandleProblemCommand(s *discordgo.Session, i *discordgo.InteractionCreate, apiClient *leetcode.LeetCodeClient) {
	data := i.ApplicationCommandData()

	if len(data.Options) == 0 {
		utils.ResponseError(s, i, "Missing required option: slug")
		return
	}

	slug := strings.TrimSpace(data.Options[0].StringValue())
	if slug == "" {
		utils.ResponseError(s, i, "Please provide a valid problem slug (e.g., two-sum, container-with-most-water")
		return
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "🔍 Looking up problem...",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Printf("failed to defer interaction for /problem: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	question, err := apiClient.GetBySlug(ctx, slug)
	if err != nil {
		utils.ResponseError(s, i, fmt.Sprintf("Problem '%s' not found. Check the slug and try again.", slug))
		return
	}

	message := buildProblemEmbed(question)

	_, followupErr := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:       "LeetCode Problem",
				Description: message,
				Color:       0xFFA116,
			},
		},
		Flags: discordgo.MessageFlagsEphemeral,
	})
	if followupErr != nil {
		log.Printf("failed to send problem followup: %v", followupErr)
	}
}

// buildProblemEmbed formats the problem data into a Discord-friendly message
func buildProblemEmbed(q *leetcode.QuestionDetail) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("**Problem Name:** %s\n", q.Title))
	sb.WriteString(fmt.Sprintf("**Difficulty:** %s\n", q.Difficulty))
	if q.AcRate > 0 {
		sb.WriteString(fmt.Sprintf("**Acceptance Rate:** %.1f%%\n", q.AcRate))
	}

	link := fmt.Sprintf("https://leetcode.com/problems/%s/", q.TitleSlug)
	sb.WriteString(fmt.Sprintf("**Link:**\n%s\n", link))

	if leetcode.IsNeetcode150Slug(q.TitleSlug) {
		sb.WriteString("\n**Part of Neetcode 150**")
	}

	return sb.String()
}
