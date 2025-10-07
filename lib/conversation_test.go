package lib

import (
	"context"
	"os"
	"testing"

	"github.com/openai/openai-go"
)

func TestNewConversation(t *testing.T) {
	client := &openai.Client{}
	config := DefaultConfig()
	systemPrompt := "You are a helpful assistant."

	conv := NewConversation(client, config, systemPrompt)

	if conv == nil {
		t.Fatal("Expected non-nil conversation")
	}

	if conv.client != client {
		t.Error("Client not set correctly")
	}

	if conv.config != config {
		t.Error("Config not set correctly")
	}

	history := conv.GetHistory()
	if len(history) != 1 {
		t.Errorf("Expected 1 message in history (system), got %d", len(history))
	}

	if history[0].Role != "system" {
		t.Errorf("Expected first message role 'system', got '%s'", history[0].Role)
	}

	if history[0].Content != systemPrompt {
		t.Errorf("Expected system message '%s', got '%s'", systemPrompt, history[0].Content)
	}
}

func TestConversation_GetHistory(t *testing.T) {
	client := &openai.Client{}
	config := DefaultConfig()
	conv := NewConversation(client, config, "System prompt")

	history := conv.GetHistory()

	if len(history) != 1 {
		t.Errorf("Expected 1 message, got %d", len(history))
	}

	if history[0].Role != "system" {
		t.Error("Expected system message")
	}
}

func TestConversation_Reset(t *testing.T) {
	client := &openai.Client{}
	config := DefaultConfig()
	systemPrompt := "System prompt"
	conv := NewConversation(client, config, systemPrompt)

	conv.history = append(conv.history,
		Message{Role: "user", Content: "Hello"},
		Message{Role: "assistant", Content: "Hi there"},
	)
	conv.messages = append(conv.messages,
		openai.UserMessage("Hello"),
		openai.AssistantMessage("Hi there"),
	)

	if len(conv.GetHistory()) != 3 {
		t.Errorf("Expected 3 messages before reset, got %d", len(conv.GetHistory()))
	}

	conv.Reset()

	history := conv.GetHistory()
	if len(history) != 1 {
		t.Errorf("Expected 1 message after reset (system), got %d", len(history))
	}

	if history[0].Role != "system" {
		t.Errorf("Expected system message after reset, got '%s'", history[0].Role)
	}

	if history[0].Content != systemPrompt {
		t.Errorf("Expected system prompt '%s', got '%s'", systemPrompt, history[0].Content)
	}
}

func TestConversation_ResetWithoutSystemMessage(t *testing.T) {
	client := &openai.Client{}
	config := DefaultConfig()
	conv := &Conversation{
		client:   client,
		config:   config,
		messages: []openai.ChatCompletionMessageParamUnion{},
		history:  []Message{},
	}

	conv.history = append(conv.history,
		Message{Role: "user", Content: "Hello"},
		Message{Role: "assistant", Content: "Hi"},
	)

	conv.Reset()

	if len(conv.GetHistory()) != 0 {
		t.Errorf("Expected empty history after reset without system message, got %d", len(conv.GetHistory()))
	}
}

func TestConversation_MessageHistory(t *testing.T) {
	client := &openai.Client{}
	config := DefaultConfig()
	conv := NewConversation(client, config, "System")

	initialHistory := conv.GetHistory()
	if len(initialHistory) != 1 {
		t.Errorf("Expected 1 initial message, got %d", len(initialHistory))
	}

	conv.history = append(conv.history,
		Message{Role: "user", Content: "First message"},
	)
	conv.messages = append(conv.messages, openai.UserMessage("First message"))

	history := conv.GetHistory()
	if len(history) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(history))
	}

	if history[1].Role != "user" || history[1].Content != "First message" {
		t.Error("Second message not stored correctly")
	}
}

func TestConversation_ModelTokenParameterGPT5Mini(t *testing.T) {
	client := &openai.Client{}
	config := &Config{
		Model:     "gpt-5-mini",
		MaxTokens: 500,
	}
	conv := NewConversation(client, config, "System")

	if conv.config.Model != "gpt-5-mini" {
		t.Errorf("Expected model gpt-5-mini, got %s", conv.config.Model)
	}
}

func TestConversation_ModelTokenParameterOther(t *testing.T) {
	client := &openai.Client{}
	config := &Config{
		Model:     "gpt-4o",
		MaxTokens: 500,
	}
	conv := NewConversation(client, config, "System")

	if conv.config.Model != "gpt-4o" {
		t.Errorf("Expected model gpt-4o, got %s", conv.config.Model)
	}
}

// Integration test (requires OPENAI_API_KEY)
func TestConversation_SendMessage_Integration(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping integration test: OPENAI_API_KEY not set")
	}

	client, config, err := NewClientFromEnv()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	conv := NewConversation(client, config, "You are a helpful assistant. Keep responses very brief.")

	ctx := context.Background()
	response, err := conv.SendMessage(ctx, "Say 'hello' and nothing else")

	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response")
	}

	history := conv.GetHistory()
	if len(history) != 3 {
		t.Errorf("Expected 3 messages (system, user, assistant), got %d", len(history))
	}

	if history[1].Role != "user" {
		t.Errorf("Expected second message to be user, got %s", history[1].Role)
	}

	if history[2].Role != "assistant" {
		t.Errorf("Expected third message to be assistant, got %s", history[2].Role)
	}
}

// Integration test (requires OPENAI_API_KEY)
func TestConversation_MultiTurn_Integration(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping integration test: OPENAI_API_KEY not set")
	}

	client, config, err := NewClientFromEnv()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	conv := NewConversation(client, config, "You are a helpful assistant. Remember what the user tells you.")

	ctx := context.Background()

	_, err = conv.SendMessage(ctx, "My name is Alice.")
	if err != nil {
		t.Fatalf("First message failed: %v", err)
	}

	response, err := conv.SendMessage(ctx, "What is my name?")
	if err != nil {
		t.Fatalf("Second message failed: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response")
	}

	history := conv.GetHistory()
	if len(history) != 5 {
		t.Errorf("Expected 5 messages (system + 2 user + 2 assistant), got %d", len(history))
	}
}

// Integration test for conversation reset (requires OPENAI_API_KEY)
func TestConversation_ResetPreservesSystemMessage_Integration(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping integration test: OPENAI_API_KEY not set")
	}

	client, config, err := NewClientFromEnv()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	systemPrompt := "You are a helpful assistant."
	conv := NewConversation(client, config, systemPrompt)

	ctx := context.Background()

	_, err = conv.SendMessage(ctx, "Hello")
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	if len(conv.GetHistory()) != 3 {
		t.Errorf("Expected 3 messages before reset, got %d", len(conv.GetHistory()))
	}

	conv.Reset()

	history := conv.GetHistory()
	if len(history) != 1 {
		t.Errorf("Expected 1 message after reset, got %d", len(history))
	}

	if history[0].Role != "system" || history[0].Content != systemPrompt {
		t.Error("System message not preserved after reset")
	}

	_, err = conv.SendMessage(ctx, "New conversation")
	if err != nil {
		t.Fatalf("SendMessage after reset failed: %v", err)
	}

	if len(conv.GetHistory()) != 3 {
		t.Errorf("Expected 3 messages after new message, got %d", len(conv.GetHistory()))
	}
}
