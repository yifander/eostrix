package utils

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

// provides some discord related helper functions
// such as response/messaging creation

// use for non-error messages (native slash commands)
func Response(s *discordgo.Session, i *discordgo.InteractionCreate, title, description string) error {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       title,
					Description: description,
					Color:       0xFFA116,
				},
			},
		},
	})
	if err != nil {
		log.Printf("failed to send interaction response: %v", err)
		return fmt.Errorf("send response: %w", err)
	}

	return nil
}

// use for all error messages (native slash commands)
func ResponseError(s *discordgo.Session, i *discordgo.InteractionCreate, description string) error {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Error",
					Description: description,
					Color:       0xFFA116,
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Printf("failed to send error response: %v", err)
		return fmt.Errorf("send error response: %w", err)
	}

	return nil
}

// send message with embed
func SendMessageComplex(s *discordgo.Session, channelID, title, description string) (*discordgo.Message, error) {
	msg, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:       title,
				Description: description,
				Color:       0xFFA116,
			},
		},
	})
	if err != nil {
		log.Printf("failed to send message to channel %s: %v", channelID, err)
		return nil, fmt.Errorf("send message to %s: %w", channelID, err)
	}

	return msg, nil
}

// send message with embed and ping
func SendPingMessageComplex(s *discordgo.Session, channelID, title, ping, description string) error {
	_, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: ping,
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:       title,
				Description: description,
				Color:       0xFFA116,
			},
		},
	})
	if err != nil {
		log.Printf("failed to send ping message to channel %s: %v", channelID, err)
		return fmt.Errorf("send ping message to %s: %w", channelID, err)
	}

	return nil
}

// first pagination post
func ResponseComponents(s *discordgo.Session, i *discordgo.InteractionCreate, content string, components []discordgo.MessageComponent) error {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Description: content,
					Color:       0xFFA116,
				},
			},
			Components: components,
		},
	})
	if err != nil {
		log.Printf("failed to send paginated response: %v", err)
		return fmt.Errorf("send paginated response: %w", err)
	}

	return nil
}

// use for all pagination pages except the initial post (editing)
func ResponseComponentsEdit(s *discordgo.Session, i *discordgo.InteractionCreate, content string, components []discordgo.MessageComponent) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Description: content,
					Color:       0xFFA116,
				},
			},
			Components: components,
		},
	})
}
