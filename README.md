# Go AI Utils

:warning: This repo is largely AI slop, for my own careful use. Use at your own risk.

A Go library for interacting with OpenAI's API with first-class support for structured outputs and conversational AI.

## Features

- **Structured Outputs**: Automatic JSON schema generation from Go structs with 100% compliance guarantee
- **Conversational AI**: Multi-turn conversations with context management
- **Simple API**: Easy-to-use functions for quick queries and structured responses
- **Environment Configuration**: Automatic client setup from environment variables

## Installation

```bash
go get github.com/bharathcs/go-ai-utils
```

## Quick Start

### Environment Setup

Set your OpenAI API key:
```bash
export OPENAI_API_KEY="your-api-key-here"
export OPENAI_MODEL="gpt-4o-2024-08-06"  # Optional, defaults to gpt-5-mini
```

### Simple Query

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    ai "github.com/bharathcs/go-ai-utils/lib"
)

func main() {
    ctx := context.Background()
    
    response, err := ai.QuickQueryFromEnv(
        ctx,
        "What is the earliest use of a siphon?",
        "You are a helpful history assistant.",
    )
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Response:", response)
}
```

### Structured Outputs

Define your Go struct and get guaranteed JSON responses:

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    
    ai "github.com/bharathcs/go-ai-utils/lib"
)

// Define your data structure
type SciFiBook struct {
    Title          string `json:"title"`
    Author         string `json:"author"`
    YearPublished  int    `json:"year_published"`
    ShortSummary   string `json:"short_summary"`
}

type SciFiBooks struct {
    Books []SciFiBook `json:"books"`
}

func main() {
    ctx := context.Background()

    // The library automatically generates JSON schema from your struct
    var sciFiBooks SciFiBooks
    err := ai.StructuredQueryFromEnv(
        ctx,
        "Recommend 3 classic science fiction books to read",
        "You are a helpful assistant that provides book recommendations.",
        &sciFiBooks,
    )

    if err != nil {
        log.Fatal(err)
    }

    // Guaranteed structured response matching your Go struct
    prettyJSON, _ := json.MarshalIndent(sciFiBooks, "", "  ")
    fmt.Printf("Structured Response:\n%s\n", string(prettyJSON))
}
```

### Conversational AI

Maintain context across multiple exchanges:

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    ai "github.com/bharathcs/go-ai-utils/lib"
)

func main() {
    ctx := context.Background()
    
    // Create client from environment
    client, config, err := ai.NewClientFromEnv()
    if err != nil {
        log.Fatal(err)
    }
    
    // Create conversation with system prompt
    conv := ai.NewConversation(
        client,
        config,
        "You are a helpful Unix command assistant.",
    )
    
    // First message
    response1, err := conv.SendMessage(ctx, "How do I list files?")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Q: How do I list files?\nA: %s\n\n", response1)
    
    // Follow-up message (maintains context)
    response2, err := conv.SendMessage(ctx, "What about hidden files?")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Q: What about hidden files?\nA: %s\n", response2)
}
```

## API Reference

### Functions

#### `QuickQueryFromEnv(ctx, prompt, systemPrompt) (string, error)`
Performs a single query without conversation state.

#### `StructuredQueryFromEnv(ctx, prompt, systemPrompt, target) error`
Performs a structured query with automatic JSON schema generation from Go structs. Uses OpenAI's native structured outputs for 100% compliance.

#### `NewClientFromEnv() (*openai.Client, *Config, error)`
Creates an OpenAI client from environment variables.

#### `NewConversation(client, config, systemPrompt) *Conversation`
Creates a new conversation with context management.

### Conversation Methods

#### `SendMessage(ctx, message) (string, error)`
Sends a message and returns the AI's response while maintaining conversation context.

#### `GetHistory() []Message`
Returns the full conversation history.

#### `Reset()`
Clears conversation history except system message.

## Environment Variables

- `OPENAI_API_KEY` (required): Your OpenAI API key
- `OPENAI_MODEL` (optional): Model to use (defaults to "gpt-5-mini")

## Structured Outputs

This library uses OpenAI's native structured outputs feature with JSON schema validation. Key benefits:

- **100% Compliance**: Responses are guaranteed to match your Go struct schema
- **Automatic Schema Generation**: No manual JSON schema definition required
- **Type Safety**: Direct unmarshaling into your Go structs
- **Model Support**: Works with gpt-4o-2024-08-06 and gpt-4o-mini-2024-07-18+

The library automatically:
1. Uses reflection to generate JSON schemas from your Go structs
2. Sets the `response_format` parameter with `json_schema` type
3. Enables `strict: true` for guaranteed schema compliance
4. Handles JSON unmarshaling into your typed structs

## Examples

See `lib/examples/main.go` for comprehensive usage examples including:
- Quick queries
- Multi-turn conversations
- Structured outputs with complex nested types
- Error handling patterns

## CLI Tool - `idk`

The `idk` CLI tool provides an interactive TUI for AI-powered command-line help.

### Installation

```bash
go install github.com/bharathcs/go-ai-utils/cmd/idk@latest
```

Or download pre-built binaries from the [releases page](https://github.com/bharathcs/go-ai-utils/releases).

### Usage

```bash
# Interactive mode
idk

# With piped input (context is preserved)
cat error.log | idk
```

## Releases

To create a new release:

1. Tag the commit with a version:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```

2. GitHub Actions will automatically:
   - Run tests
   - Build binaries for Linux, macOS, and Windows (amd64 and arm64)
   - Create a GitHub release with the binaries attached

## License

MIT License - see LICENSE file for details.
