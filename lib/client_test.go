package lib

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Model != "gpt-5-mini" {
		t.Errorf("Expected model gpt-5-mini, got %s", config.Model)
	}

	if config.MaxTokens != 1000 {
		t.Errorf("Expected MaxTokens 1000, got %d", config.MaxTokens)
	}
}

func TestNewClientFromEnv_MissingAPIKey(t *testing.T) {
	oldKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", oldKey)

	_, _, err := NewClientFromEnv()
	if err == nil {
		t.Error("Expected error when OPENAI_API_KEY is missing")
	}

	expectedMsg := "OPENAI_API_KEY environment variable is required"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestNewClientFromEnv_CustomModel(t *testing.T) {
	oldKey := os.Getenv("OPENAI_API_KEY")
	oldModel := os.Getenv("OPENAI_MODEL")

	os.Setenv("OPENAI_API_KEY", "test-key")
	os.Setenv("OPENAI_MODEL", "gpt-4o")

	defer func() {
		os.Setenv("OPENAI_API_KEY", oldKey)
		os.Setenv("OPENAI_MODEL", oldModel)
	}()

	_, config, err := NewClientFromEnv()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if config.Model != "gpt-4o" {
		t.Errorf("Expected model gpt-4o, got %s", config.Model)
	}
}

func TestGenerateJSONSchema_SimpleStruct(t *testing.T) {
	type TestStruct struct {
		Name  string  `json:"name"`
		Age   int     `json:"age"`
		Score float64 `json:"score"`
	}

	schema := generateJSONSchema(TestStruct{})

	if schema["type"] != "object" {
		t.Errorf("Expected type 'object', got %v", schema["type"])
	}

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be a map")
	}

	if len(properties) != 3 {
		t.Errorf("Expected 3 properties, got %d", len(properties))
	}

	nameSchema, ok := properties["name"].(map[string]interface{})
	if !ok || nameSchema["type"] != "string" {
		t.Error("Expected name property to be string type")
	}

	ageSchema, ok := properties["age"].(map[string]interface{})
	if !ok || ageSchema["type"] != "integer" {
		t.Error("Expected age property to be integer type")
	}

	scoreSchema, ok := properties["score"].(map[string]interface{})
	if !ok || scoreSchema["type"] != "number" {
		t.Error("Expected score property to be number type")
	}
}

func TestGenerateJSONSchema_NestedStruct(t *testing.T) {
	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
	}

	type Person struct {
		Name    string  `json:"name"`
		Address Address `json:"address"`
	}

	schema := generateJSONSchema(Person{})

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be a map")
	}

	addressSchema, ok := properties["address"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected address property to exist")
	}

	if addressSchema["type"] != "object" {
		t.Errorf("Expected address type 'object', got %v", addressSchema["type"])
	}

	addressProps, ok := addressSchema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected address properties to be a map")
	}

	if len(addressProps) != 2 {
		t.Errorf("Expected 2 address properties, got %d", len(addressProps))
	}
}

func TestGenerateJSONSchema_ArrayField(t *testing.T) {
	type TestStruct struct {
		Tags []string `json:"tags"`
	}

	schema := generateJSONSchema(TestStruct{})

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be a map")
	}

	tagsSchema, ok := properties["tags"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected tags property to exist")
	}

	if tagsSchema["type"] != "array" {
		t.Errorf("Expected tags type 'array', got %v", tagsSchema["type"])
	}

	items, ok := tagsSchema["items"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected items to be a map")
	}

	if items["type"] != "string" {
		t.Errorf("Expected items type 'string', got %v", items["type"])
	}
}

func TestGenerateJSONSchema_SkipsUnexportedFields(t *testing.T) {
	type TestStruct struct {
		Public  string `json:"public"`
		private string `json:"private"`
	}

	schema := generateJSONSchema(TestStruct{})

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be a map")
	}

	if len(properties) != 1 {
		t.Errorf("Expected 1 property (private should be skipped), got %d", len(properties))
	}

	if _, exists := properties["public"]; !exists {
		t.Error("Expected public field to exist")
	}

	if _, exists := properties["private"]; exists {
		t.Error("Expected private field to be skipped")
	}
}

func TestGenerateJSONSchema_SkipsJSONDashTag(t *testing.T) {
	type TestStruct struct {
		Include string `json:"include"`
		Exclude string `json:"-"`
	}

	schema := generateJSONSchema(TestStruct{})

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be a map")
	}

	if len(properties) != 1 {
		t.Errorf("Expected 1 property (json:\"-\" should be skipped), got %d", len(properties))
	}

	if _, exists := properties["include"]; !exists {
		t.Error("Expected include field to exist")
	}

	if _, exists := properties["Exclude"]; exists {
		t.Error("Expected field with json:\"-\" to be skipped")
	}
}

func TestGenerateFieldSchema_AllTypes(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"string", "", "string"},
		{"int", 0, "integer"},
		{"int8", int8(0), "integer"},
		{"int16", int16(0), "integer"},
		{"int32", int32(0), "integer"},
		{"int64", int64(0), "integer"},
		{"float32", float32(0), "number"},
		{"float64", float64(0), "number"},
		{"bool", false, "boolean"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := generateFieldSchema(reflect.TypeOf(tt.value))
			if schema["type"] != tt.expected {
				t.Errorf("Expected type '%s', got '%v'", tt.expected, schema["type"])
			}
		})
	}
}

func TestCreateChatCompletionParams_GPT5Mini(t *testing.T) {
	config := &Config{
		Model:     "gpt-5-mini",
		MaxTokens: 500,
	}

	params := createChatCompletionParams(config, nil)

	if params.Model != "gpt-5-mini" {
		t.Errorf("Expected model gpt-5-mini, got %s", params.Model)
	}
}

func TestCreateChatCompletionParams_OtherModels(t *testing.T) {
	config := &Config{
		Model:     "gpt-4o",
		MaxTokens: 500,
	}

	params := createChatCompletionParams(config, nil)

	if params.Model != "gpt-4o" {
		t.Errorf("Expected model gpt-4o, got %s", params.Model)
	}
}

func TestStructuredQueryFromEnv_SchemaGeneration(t *testing.T) {
	type CommandSolution struct {
		Command     string `json:"command"`
		Description string `json:"description"`
		Confidence  int    `json:"confidence"`
	}

	schema := generateJSONSchema(CommandSolution{})

	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal schema: %v", err)
	}

	var parsedSchema map[string]interface{}
	err = json.Unmarshal(schemaJSON, &parsedSchema)
	if err != nil {
		t.Fatalf("Failed to parse schema JSON: %v", err)
	}

	if parsedSchema["type"] != "object" {
		t.Error("Expected schema type to be 'object'")
	}

	properties := parsedSchema["properties"].(map[string]interface{})
	if len(properties) != 3 {
		t.Errorf("Expected 3 properties, got %d", len(properties))
	}
}

// Integration test (requires OPENAI_API_KEY)
func TestQuickQueryFromEnv_Integration(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping integration test: OPENAI_API_KEY not set")
	}

	ctx := context.Background()
	response, err := QuickQueryFromEnv(ctx, "Say 'test' and nothing else", "You are a helpful assistant.")

	if err != nil {
		t.Fatalf("QuickQueryFromEnv failed: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response")
	}
}

// Integration test (requires OPENAI_API_KEY)
func TestStructuredQueryFromEnv_Integration(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping integration test: OPENAI_API_KEY not set")
	}

	type TestResponse struct {
		Answer string `json:"answer"`
		Score  int    `json:"score"`
	}

	ctx := context.Background()
	var result TestResponse

	err := StructuredQueryFromEnv(
		ctx,
		"What is 2+2? Provide the answer and a confidence score from 1-10.",
		"You are a math assistant.",
		&result,
	)

	if err != nil {
		t.Fatalf("StructuredQueryFromEnv failed: %v", err)
	}

	if result.Answer == "" {
		t.Error("Expected non-empty answer")
	}

	if result.Score < 1 || result.Score > 10 {
		t.Errorf("Expected score between 1-10, got %d", result.Score)
	}
}
