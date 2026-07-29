package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"slices"
	"strings"

	"github.com/danny-pearson/word-association/internal/db"
	"github.com/danny-pearson/word-association/internal/llm"
	"github.com/danny-pearson/word-association/internal/puzzle"
	"github.com/joho/godotenv"
)

// Small command for manually probing and evaluating model output
func main() {
	_ = godotenv.Load()

	var (
		model   = os.Getenv("OPENROUTER_MODEL")
		baseURL = os.Getenv("OPENROUTER_BASE_URL")
		apiKey  = os.Getenv("OPENROUTER_API_KEY")
	)

	llmClient := llm.New(baseURL, apiKey, model, 0.7)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	conn, err := db.Open("data.db")

	if err != nil {
		log.Fatalf("failed to open connection: %v", err)
	}

	store := db.NewStore(conn)

	avoid, err := store.RecentAnswers(ctx, 200)

	if err != nil {
		log.Fatalf("recent answers query failed: %v", err)
	}

	text, err := llmClient.Send(ctx, fmt.Sprintf(puzzle.Prompt, strings.Join(avoid, ", ")))

	if err != nil {
		log.Fatalf("an error occurred: %v", err)
	}

	var puzzles []puzzle.Puzzle

	if err := json.Unmarshal([]byte(extractJSON(text)), &puzzles); err != nil {
		log.Fatalf("parsing response: %v\n\nraw:\n%s", err, text)
	}

	var (
		validPuzzles []puzzle.Puzzle
		validAnswers []string
	)

	for i, p := range puzzles {
		if err := puzzle.Validate(p); err != nil {
			fmt.Printf("%d REJECT %-15s %v\n", i+1, p.Answer, strings.Join(p.Clues, ", "))
			continue
		}

		validPuzzles = append(validPuzzles, p)
		validAnswers = append(validAnswers, p.Answer)

		fmt.Printf("%d. OK	%-15s %s\n", i+1, p.Answer, strings.Join(p.Clues, ", "))
	}

	existingAnswers, err := store.FindExistingAnswers(ctx, validAnswers)

	if err != nil {
		log.Fatalf("existing answers query failed: %v", err)
	}

	for _, p := range validPuzzles {
		if slices.Contains(existingAnswers, p.Answer) {
			continue
		}

		_, err := store.SavePuzzle(ctx, p, model)

		if err != nil {
			log.Fatalf("saving puzzle failed: %v", err)
		}
	}
}

func extractJSON(s string) string {
	s = strings.TrimSpace(s)

	if i := strings.Index(s, "["); i >= 0 {
		if j := strings.LastIndex(s, "]"); j > i {
			return s[i : j+1]
		}
	}
	return s
}
