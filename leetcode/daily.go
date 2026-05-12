package leetcode

import (
	"context"
	"eostrix/config"
	"eostrix/utils"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// PostDailyChallenge fetches the problem via the shared API client and posts to Discord
func PostDailyChallenge(session *discordgo.Session, slug string, apiClient *LeetCodeClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	question, err := apiClient.GetBySlug(ctx, slug)
	if err != nil {
		log.Printf("failed to fetch problem '%s': %v", slug, err)
		return
	}

	message := buildDailyChallengeMessage(question)

	cfg := config.ParseConfig()
	ping := fmt.Sprintf("<@&%s> ", cfg.LeetcodeRoleID)

	err = utils.SendPingMessageComplex(session, cfg.DefaultChannel, "Daily Neetcode Challenge", ping, message)
	if err != nil {
		log.Printf("failed to respond: %v", err)
		return
	}
}

// buildDailyChallengeMessage formats the API response into a Discord-friendly message
func buildDailyChallengeMessage(q *QuestionDetail) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("**Problem Name:** %s\n", q.Title))
	sb.WriteString(fmt.Sprintf("**Difficulty:** %s\n", q.Difficulty))

	link := fmt.Sprintf("https://leetcode.com/problems/%s/", q.TitleSlug)
	sb.WriteString(fmt.Sprintf("**Link:**\n%s\n", link))

	if IsNeetcode150Slug(q.TitleSlug) {
		sb.WriteString("\n**Part of Neetcode 150**")
	}

	return sb.String()
}
