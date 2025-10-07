# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go library for interacting with OpenAI's API with first-class support for structured outputs and conversational AI. The library wraps the openai-go SDK to provide ergonomic functions for common use cases.

## Development Commands

### Building
```bash
go build ./...
```

### Testing
```bash
go test ./lib/...
```

### Running Examples
```bash
# Run the examples
go run lib/examples/main.go

# Run the idk CLI tool
go run cmd/idk/main.go
```

### Installing the CLI
```bash
go install ./cmd/idk
```

## Architecture

### Core Library (`lib/`)

**client.go**: Main API functions and client creation
- `NewClientFromEnv()`: Creates OpenAI client from `OPENAI_API_KEY` environment variable
- `QuickQueryFromEnv()`: Single query without conversation state
- `StructuredQueryFromEnv()`: Structured query with automatic JSON schema generation using reflection
- Schema generation supports nested structs, arrays, and primitives
- Model-specific parameter handling: `gpt-5-mini` uses `MaxCompletionTokens`, other models use `MaxTokens`

**conversation.go**: Multi-turn conversation management
- `Conversation` maintains message history for context across multiple exchanges
- `SendMessage()`: Sends message and returns response while maintaining context
- `GetHistory()`: Returns full conversation history
- `Reset()`: Clears conversation except system message

### Structured Outputs

The library uses OpenAI's native structured outputs feature with JSON schema validation:
- Automatic schema generation from Go structs via reflection (`generateJSONSchema`)
- `strict: true` for guaranteed 100% schema compliance
- Supports nested structs, arrays, and primitive types
- Works with gpt-4o-2024-08-06 and gpt-4o-mini-2024-07-18+ models

### CLI Tool (`cmd/idk/`)

Interactive TUI application for AI-powered command-line help built with Bubble Tea. Uses structured outputs to get command solutions in a typed format (`CommandSolution` struct).

**States**:
- `StateInput`: User typing query
- `StateLoading`: Waiting for AI response with animated spinner
- `StateShowingSolutions`: Displaying up to 3 ranked solutions

**Key Features**:
- Piped input support: Reads from stdin (truncated to 400 chars) and reopens `/dev/tty` for interactive input
- Context preservation: Piped content is included with every query
- Solution execution: Selected commands are typed out character-by-character to terminal
- Error handling: Specific error messages for auth failures, timeouts, rate limits, server errors

## Environment Variables

- `OPENAI_API_KEY` (required): OpenAI API key
- `OPENAI_MODEL` (optional): Model to use (defaults to "gpt-5-mini")
- `OPENAI_BASE_URL` (optional): Custom endpoint for OpenAI API

## Dependencies

- `github.com/openai/openai-go`: OpenAI SDK
- `github.com/charmbracelet/bubbletea`: TUI framework for idk CLI
- `github.com/charmbracelet/lipgloss`: Styling for TUI

## Model Compatibility

The code handles model-specific parameters:
- `gpt-5-mini`: Uses `MaxCompletionTokens` parameter
- Other models: Use `MaxTokens` parameter
