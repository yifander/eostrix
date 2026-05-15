package commands

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func RegisterCommands(s *discordgo.Session) []*discordgo.ApplicationCommand {
	commands := []*discordgo.ApplicationCommand{
		{
			// /company <name> <difficulty>
			Name:        "company",
			Description: "Get LeetCode problems by company",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:         "name",
					Description:  "Company name - eg. ByteDance",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     true,
					Autocomplete: true,
				},
				{
					Name:        "difficulty",
					Description: "difficulty (easy/medium/hard/all)",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Easy", Value: "easy"},
						{Name: "Medium", Value: "medium"},
						{Name: "Hard", Value: "hard"},
						{Name: "All", Value: "all"},
					},
				},
			},
		},
		{
			// /randlc <difficulty>
			Name:        "randlc",
			Description: "Returns a random LeetCode problem by difficulty",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "difficulty",
					Description: "Select difficulty for random problem",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Easy", Value: "easy"},
						{Name: "Medium", Value: "medium"},
						{Name: "Hard", Value: "hard"},
					},
				},
			},
		},
		{
			// /topics <category> <difficulty>
			Name:        "topics",
			Description: "Returns LeetCode problems associated with a topic category",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:         "category",
					Description:  "Category - eg. Array",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     true,
					Autocomplete: true,
				},
				{
					Name:        "difficulty",
					Description: "difficulty (easy/medium/hard)",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Easy", Value: "easy"},
						{Name: "Medium", Value: "medium"},
						{Name: "Hard", Value: "hard"},
					},
				},
			},
		},
		{
			// /curated <difficulty>
			Name:        "curated",
			Description: "Show most common LeetCode problems across companies",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "difficulty",
					Description: "Filter by difficulty",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Easy", Value: "easy"},
						{Name: "Medium", Value: "medium"},
						{Name: "Hard", Value: "hard"},
						{Name: "All", Value: "all"},
					},
				},
			},
		},
		{
			// /problem <slug>
			Name:        "problem",
			Description: "Get LeetCode problem details by slug",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "slug",
					Description: "slug (e.g., two-sum)",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},
		{
			Name:        "eostrix",
			Description: "Show bot command directory and features",
		},
	}

	_, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, "", commands)
	if err != nil {
		log.Printf("Failed to register commands: %v", err)
	} else {
		log.Printf("Registered %d commands", len(commands))
	}

	return commands
}
