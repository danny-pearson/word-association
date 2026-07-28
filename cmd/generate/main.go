package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/danny-pearson/word-association/internal/llm"
	"github.com/danny-pearson/word-association/internal/puzzle"
	"github.com/joho/godotenv"
)

// Small command for manually probing and evaluating model output
func main() {
	_ = godotenv.Load()

	client := llm.New(
		os.Getenv("OPENROUTER_BASE_URL"),
		os.Getenv("OPENROUTER_API_KEY"),
		os.Getenv("OPENROUTER_MODEL"),
		0.7,
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	avoid := []string{
		"rain", "clock", "fire", "ocean", "honey", "library", "guitar", "shadow",
		"coffee", "snow", "moon", "tree", "key", "sun", "mountain", "paper",
		"scale", "pitch", "crane", "bat", "trunk", "match", "lead", "bark",
		"bow", "box", "glass", "net", "court", "rock", "jam", "press", "fair", "train",
	}

	text, err := client.Send(ctx, fmt.Sprintf(puzzle.Prompt, strings.Join(avoid, ", ")))

	if err != nil {
		log.Fatalf("an error occurred: %v", err)
	}

	fmt.Printf("%s", text)
}
