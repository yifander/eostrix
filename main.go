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
	"time"

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

	apiClient := leetcode.NewLeetCodeClient()
	store := leetcode.NewProblemStore()
	if err := store.Load("data"); err != nil {
		log.Fatalf("failed to load leetcode data: %v", err)
	}

	initHandlers(disc, store, apiClient)
	loadFeatures(disc, store, apiClient)

	fmt.Println("bot has started ...")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func initHandlers(session *discordgo.Session, store *leetcode.ProblemStore, apiClient *leetcode.LeetCodeClient) {
	commands.RegisterCommands(session)

	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			handleSlashCommand(s, i, store, apiClient)
		case discordgo.InteractionApplicationCommandAutocomplete:
			handleAutocomplete(s, i, store)
		case discordgo.InteractionMessageComponent:
			handleComponentInteraction(s, i, store)
		}
	})
}

func handleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate, store *leetcode.ProblemStore, apiClient *leetcode.LeetCodeClient) {
	switch i.ApplicationCommandData().Name {
	case "company":
		commands.HandleCompanyCommand(s, i, store)
	case "randlc":
		commands.HandleRandCommand(s, i, store)
	case "topics":
		commands.HandleTopicsCommand(s, i, store)
	case "curated":
		commands.HandleCuratedCommand(s, i, store)
	case "problem":
		commands.HandleProblemCommand(s, i, apiClient)
	case "eostrix":
		commands.HandleEostrixCommand(s, i)
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
	case strings.HasPrefix(cid, "company|"):
		commands.HandleCompanyPageChange(s, i, store)
	case strings.HasPrefix(cid, "topics|"):
		commands.HandleTopicsPageChange(s, i, store)
	case strings.HasPrefix(cid, "curated|"):
		commands.HandleCuratedPageChange(s, i, store)
	}
}

func loadFeatures(session *discordgo.Session, store *leetcode.ProblemStore, apiClient *leetcode.LeetCodeClient) {
	utils.ScheduleMidnightUTCEvent(func() {
		slug := leetcode.GetRandomNeetcodeSlug()
		leetcode.PostDailyChallenge(session, slug, apiClient)
	})

	store.BuildCuratedProblems()
	leetcode.InitPagination(15 * time.Minute)

	log.Printf("Built curated list: %d problems",
		store.CuratedCount())
}
