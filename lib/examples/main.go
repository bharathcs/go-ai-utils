package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	ai "github.com/bharathcs/go-ai-utils/lib"
)

type SciFiBook struct {
	Title         string `json:"title"`
	Author        string `json:"author"`
	YearPublished int    `json:"year_published"`
	ShortSummary  string `json:"short_summary"`
}

type SciFiBooks struct {
	Books []SciFiBook `json:"books"`
}

func main() {
	ctx := context.Background()

	// Example 1: Quick query without conversation state
	fmt.Println("=== Example 1: Quick Query ===")
	response, err := ai.QuickQueryFromEnv(
		ctx,
		"What is the earliest use of a siphon?",
		"You are a helpful history assistant.",
	)
	if err != nil {
		log.Printf("Quick query error: %v", err)
	} else {
		fmt.Printf("Response: %s\n\n", response)
	}

	// Example 2: Multi-turn conversation
	fmt.Println("=== Example 2: Conversation ===")

	// Create client from environment
	client, config, err := ai.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	// Create conversation with system prompt
	conv := ai.NewConversation(
		client,
		config,
		"You are a helpful command-line assistant that provides Unix commands.",
	)

	// First message
	response1, err := conv.SendMessage(ctx, "How do I list all files in a directory?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Q: How do I list all files in a directory?\n")
	fmt.Printf("A: %s\n\n", response1)

	// Follow-up message (maintains context)
	response2, err := conv.SendMessage(ctx, "What about hidden files too?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Q: What about hidden files too?\n")
	fmt.Printf("A: %s\n\n", response2)

	// Example 3: Get conversation history
	fmt.Println("=== Example 3: Conversation History ===")
	history := conv.GetHistory()
	for i, msg := range history {
		fmt.Printf("%d. [%s] %s\n", i+1, msg.Role, msg.Content)
	}
	fmt.Println()

	// Example 4: Structured Output
	fmt.Println("=== Example 4: Structured Output ===")
	var sciFiBooks SciFiBooks
	err = ai.StructuredQueryFromEnv(
		ctx,
		"Recommend 3 classic science fiction books to read. Include title, author, year published, and a short summary.",
		"You are a helpful assistant that provides book recommendations.",
		&sciFiBooks,
	)
	if err != nil {
		log.Printf("Structured query error: %v\n", err)
	} else {
		// Pretty print the structured output
		prettyJSON, _ := json.MarshalIndent(sciFiBooks, "", "  ")
		fmt.Printf("Structured Response:\n%s\n\n", string(prettyJSON))

		// You can also access individual fields
		fmt.Printf("Found %d books:\n", len(sciFiBooks.Books))
		for i, book := range sciFiBooks.Books {
			fmt.Printf("%d. %s by %s (%d)\n   %s\n",
				i+1, book.Title, book.Author, book.YearPublished, book.ShortSummary)
		}
	}
}
