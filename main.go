package main

import (
	"eostrix/commands"
	"eostrix/config"
	"eostrix/leetcode"
	"eostrix/utils"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func main() {
	config := config.ParseConfig()

	disc, err := discordgo.New("Bot " + config.SecurityToken)
	if err != nil {
		log.Fatal(err)
	}

	if err := disc.Open(); err != nil {
		log.Fatal(err)
	}
	defer disc.Close()

	store := leetcode.NewProblemStore()
	if err := store.Load("data"); err != nil {
		log.Fatalf("failed to load leetcode data: %v", err)
	}

	initHandlers(disc, store)
	loadFeatures(disc)

	fmt.Println("bot has started ...")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func initHandlers(session *discordgo.Session, store *leetcode.ProblemStore) {
	commands.RegisterCommands(session)

	// Single handler for all interaction types (cleaner + avoids duplicate registration)
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			handleSlashCommand(s, i, store)
		case discordgo.InteractionApplicationCommandAutocomplete:
			handleAutocomplete(s, i, store)
		case discordgo.InteractionMessageComponent:
			handleComponentInteraction(s, i, store)
		}
	})
}

func handleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate, store *leetcode.ProblemStore) {
	switch i.ApplicationCommandData().Name {
	case "company":
		commands.HandleCompanyCommand(s, i, store)
	case "randlc":
		commands.HandleRandCommand(s, i, store)
	case "topics":
		commands.HandleTopicsCommand(s, i, store)
	}
}

func handleAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate, store *leetcode.ProblemStore) {
	switch i.ApplicationCommandData().Name {
	case "company":
		commands.CompanyAutocomplete(s, i, store)
	case "topics":
		commands.TopicsAutocomplete(s, i, store)
	}
}

func handleComponentInteraction(s *discordgo.Session, i *discordgo.InteractionCreate, store *leetcode.ProblemStore) {
	cid := i.MessageComponentData().CustomID
	switch {
	case strings.HasPrefix(cid, "company_"):
		commands.HandleCompanyPageChange(s, i, 0)
	case strings.HasPrefix(cid, "topics_"):
		commands.HandleTopicsPageChange(s, i, store)
	}
}

func loadFeatures(disc *discordgo.Session) {
	utils.ScheduleMidnightUTCEvent(func() {
		leetcode.PostDailyChallenge(disc, leetcode.GetRandomNeetcodeSlug())
	})
}
