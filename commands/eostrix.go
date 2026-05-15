package commands

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

// HandleEostrixCommand displays a directory of all bot commands and features
func HandleEostrixCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	embed := &discordgo.MessageEmbed{
		Title:       "Eostrix Bot - Command Directory",
		Description: "An extinct owl species that matches my extinct desire to do Leetcode.",
		Color:       0xFFA116,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "[Command] `/company <name> <difficulty>`",
				Value:  "Browse problems asked by specific companies.",
				Inline: false,
			},
			{
				Name:   "[Command] `/topics <category> <difficulty>`",
				Value:  "Filter problems by topic/category.",
				Inline: false,
			},
			{
				Name:   "[Command] `/randlc <difficulty>`",
				Value:  "Get a random problem to practice.",
				Inline: false,
			},
			{
				Name:   "[Command] `/curated <limit> <difficulty>`",
				Value:  "See the most frequently asked problems across companies.",
				Inline: false,
			},
			{
				Name:   "[Command] `/problem <slug>`",
				Value:  "Get detailed info about a specific problem.",
				Inline: false,
			},
			{
				Name:   "[Feature] Daily Neetcode Challenge",
				Value:  "Automatic daily problem posted at midnight UTC with role ping.",
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Data sourced from LeetCode & liquidslr/leetcode-company-wise-problems • Eostrix v1.4",
		},
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
	if err != nil {
		log.Printf("failed to respond to /eostrix: %v", err)
	}
}
