package lib

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
)

// Message represents a single message in a conversation
type Message struct {
	Role    string
	Content string
}

// Conversation manages a multi-turn conversation with the AI
type Conversation struct {
	client   *openai.Client
	config   *Config
	messages []openai.ChatCompletionMessageParamUnion
	history  []Message // Keep a simple history for easier access
}

// NewConversation creates a new conversation with a system prompt
func NewConversation(client *openai.Client, config *Config, systemPrompt string) *Conversation {
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemPrompt),
	}

	history := []Message{
		{Role: "system", Content: systemPrompt},
	}

	return &Conversation{
		client:   client,
		config:   config,
		messages: messages,
		history:  history,
	}
}

// SendMessage sends a user message and returns the AI's response
func (c *Conversation) SendMessage(ctx context.Context, message string) (string, error) {
	// Add user message to conversation history
	c.messages = append(c.messages, openai.UserMessage(message))
	c.history = append(c.history, Message{Role: "user", Content: message})

	// Get AI response
	params := openai.ChatCompletionNewParams{
		Model:    c.config.Model,
		Messages: c.messages,
	}

	// Use different token parameter based on model
	modelStr := string(c.config.Model)
	if modelStr == "gpt-5-mini" {
		params.MaxCompletionTokens = openai.Int(int64(c.config.MaxTokens))
	} else {
		params.MaxTokens = openai.Int(int64(c.config.MaxTokens))
	}

	resp, err := c.client.Chat.Completions.New(ctx, params)

	if err != nil {
		return "", fmt.Errorf("API request failed (model: %s): %w", c.config.Model, err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned from API (model: %s, id: %s)", c.config.Model, resp.ID)
	}

	// Add AI response to conversation history
	aiResponse := resp.Choices[0].Message.Content
	if aiResponse == "" {
		return "", fmt.Errorf("empty response content from API (model: %s, finish_reason: %s, id: %s)", c.config.Model, resp.Choices[0].FinishReason, resp.ID)
	}
	c.messages = append(c.messages, openai.AssistantMessage(aiResponse))
	c.history = append(c.history, Message{Role: "assistant", Content: aiResponse})

	return aiResponse, nil
}

// GetHistory returns the conversation history as a slice of Messages
func (c *Conversation) GetHistory() []Message {
	return c.history
}

// Reset clears the conversation history except for the system message
func (c *Conversation) Reset() {
	if len(c.history) > 0 && c.history[0].Role == "system" {
		systemMessage := c.history[0]
		c.messages = []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemMessage.Content),
		}
		c.history = []Message{systemMessage}
	} else {
		c.messages = []openai.ChatCompletionMessageParamUnion{}
		c.history = []Message{}
	}
}
