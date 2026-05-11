package commands

import (
	"eostrix/leetcode"
	"eostrix/utils"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// provides a random leetcode problem based on difficulty
// /randlc <difficulty>

func HandleRandCommand(s *discordgo.Session, i *discordgo.InteractionCreate, store *leetcode.ProblemStore) {
	data := i.ApplicationCommandData()
	opts := data.Options

	difficulty := strings.ToLower(opts[0].StringValue())

	switch difficulty {
	case "easy", "medium", "hard", "all":
	default:
		utils.ResponseError(s, i, "Invalid difficulty provided.")
		return
	}

	var candidates []*leetcode.Problem
	if difficulty == "all" {
		all := store.All()
		candidates = make([]*leetcode.Problem, len(all))
		for i := range all {
			candidates[i] = &all[i]
		}
	} else {
		candidates = store.ByDifficulty(difficulty)
	}

	if len(candidates) == 0 {
		utils.ResponseError(s, i, fmt.Sprintf("No problems found for difficulty '%s'", difficulty))
		return
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomIndex := rng.Intn(len(candidates))
	problem := candidates[randomIndex]

	var builder strings.Builder

	// users said they prefered leetcode problems without the topics because it felt
	// more realistic towards an interview/challenge
	builder.WriteString(fmt.Sprintf("**Challenge Name:** %s\n", problem.Title))
	builder.WriteString(fmt.Sprintf("**Difficulty:** %s\n", problem.Difficulty))
	builder.WriteString(fmt.Sprintf("**Link:** \n%s\n", problem.Link))

	utils.Response(s, i, fmt.Sprintf("Random %s NeetCode Problem", problem.Difficulty), builder.String())
}
