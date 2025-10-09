package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// Config holds configuration for the AI client
type Config struct {
	Model     openai.ChatModel
	MaxTokens int
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Model:     openai.ChatModel("gpt-5-mini"),
		MaxTokens: 1000,
	}
}

// createChatCompletionParams creates the appropriate chat completion parameters based on the model
func createChatCompletionParams(config *Config, messages []openai.ChatCompletionMessageParamUnion) openai.ChatCompletionNewParams {
	params := openai.ChatCompletionNewParams{
		Model:    config.Model,
		Messages: messages,
	}

	// Use different token parameter based on model
	modelStr := string(config.Model)
	if modelStr == "gpt-5-mini" {
		params.MaxCompletionTokens = openai.Int(int64(config.MaxTokens))
	} else {
		params.MaxTokens = openai.Int(int64(config.MaxTokens))
	}

	return params
}

// NewClientFromEnv creates an OpenAI client from environment variables
func NewClientFromEnv() (*openai.Client, *Config, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	// Get base URL from environment or use default
	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	)
	config := DefaultConfig()

	// Override defaults with environment variables if set
	if model := os.Getenv("OPENAI_MODEL"); model != "" {
		switch model {
		case "gpt-4":
			config.Model = openai.ChatModelGPT4
		case "gpt-4o":
			config.Model = openai.ChatModelGPT4o
		case "gpt-4o-mini":
			config.Model = openai.ChatModelGPT4oMini
		case "gpt-3.5-turbo":
			config.Model = openai.ChatModelGPT3_5Turbo
		case "gpt-5-mini":
			config.Model = openai.ChatModel("gpt-5-mini")
		default:
			config.Model = openai.ChatModel(model)
		}
	}

	return &client, config, nil
}

// QuickQueryFromEnv performs a single query using environment configuration
func QuickQueryFromEnv(ctx context.Context, prompt, systemPrompt string) (string, error) {
	client, config, err := NewClientFromEnv()
	if err != nil {
		return "", err
	}

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemPrompt),
		openai.UserMessage(prompt),
	}

	resp, err := client.Chat.Completions.New(ctx, createChatCompletionParams(config, messages))

	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned from API (model: %s, id: %s)", config.Model, resp.ID)
	}

	content := resp.Choices[0].Message.Content
	if content == "" {
		// Debug: output full response
		respJSON, _ := json.MarshalIndent(resp, "", "  ")
		return "", fmt.Errorf("empty response content from API\nModel: %s\nFinish Reason: %s\nID: %s\nFull Response:\n%s",
			config.Model, resp.Choices[0].FinishReason, resp.ID, string(respJSON))
	}

	return content, nil
}

// generateJSONSchema generates a JSON schema from a Go struct type using reflection
func generateJSONSchema(v interface{}) map[string]interface{} {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	schema := map[string]interface{}{
		"type":                 "object",
		"properties":           map[string]interface{}{},
		"required":             []string{},
		"additionalProperties": false,
	}

	properties := schema["properties"].(map[string]interface{})
	required := []string{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		fieldName := field.Name
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				fieldName = parts[0]
			}
		}

		fieldSchema := generateFieldSchema(field.Type)
		properties[fieldName] = fieldSchema
		required = append(required, fieldName)
	}

	schema["required"] = required
	return schema
}

// generateFieldSchema generates schema for a single field type
func generateFieldSchema(t reflect.Type) map[string]interface{} {
	switch t.Kind() {
	case reflect.String:
		return map[string]interface{}{"type": "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return map[string]interface{}{"type": "integer"}
	case reflect.Float32, reflect.Float64:
		return map[string]interface{}{"type": "number"}
	case reflect.Bool:
		return map[string]interface{}{"type": "boolean"}
	case reflect.Slice:
		elemSchema := generateFieldSchema(t.Elem())
		return map[string]interface{}{
			"type":  "array",
			"items": elemSchema,
		}
	case reflect.Struct:
		// For nested structs, generate a schema recursively
		properties := map[string]interface{}{}
		required := []string{}

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}

			jsonTag := field.Tag.Get("json")
			if jsonTag == "-" {
				continue
			}

			fieldName := field.Name
			if jsonTag != "" {
				parts := strings.Split(jsonTag, ",")
				if parts[0] != "" {
					fieldName = parts[0]
				}
			}

			properties[fieldName] = generateFieldSchema(field.Type)
			required = append(required, fieldName)
		}

		return map[string]interface{}{
			"type":                 "object",
			"properties":           properties,
			"required":             required,
			"additionalProperties": false,
		}
	default:
		return map[string]interface{}{"type": "string"} // fallback
	}
}

// StructuredQueryFromEnv performs a structured query using OpenAI's native structured outputs
func StructuredQueryFromEnv(ctx context.Context, prompt, systemPrompt string, target interface{}) error {
	client, config, err := NewClientFromEnv()
	if err != nil {
		return err
	}

	// Generate JSON schema from the target struct
	schema := generateJSONSchema(target)

	// Create schema name from struct type
	t := reflect.TypeOf(target)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	schemaName := strings.ToLower(t.Name())

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemPrompt),
		openai.UserMessage(prompt),
	}

	// Create params with structured output
	params := openai.ChatCompletionNewParams{
		Model:    config.Model,
		Messages: messages,
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:   schemaName,
					Schema: schema,
					Strict: openai.Bool(true),
				},
			},
		},
	}

	// Set token limits based on model
	modelStr := string(config.Model)
	if modelStr == "gpt-5-mini" {
		params.MaxCompletionTokens = openai.Int(int64(config.MaxTokens))
	} else {
		params.MaxTokens = openai.Int(int64(config.MaxTokens))
	}

	resp, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		return err
	}

	if len(resp.Choices) == 0 {
		return fmt.Errorf("no response choices returned from API (model: %s, id: %s)", config.Model, resp.ID)
	}

	content := resp.Choices[0].Message.Content
	if content == "" {
		// Debug: output full response
		respJSON, _ := json.MarshalIndent(resp, "", "  ")
		return fmt.Errorf("empty response content from API\nModel: %s\nFinish Reason: %s\nID: %s\nFull Response:\n%s",
			config.Model, resp.Choices[0].FinishReason, resp.ID, string(respJSON))
	}

	err = json.Unmarshal([]byte(content), target)
	if err != nil {
		return fmt.Errorf("failed to parse JSON response: %w (content preview: %.100s...)", err, content)
	}
	return nil
}
