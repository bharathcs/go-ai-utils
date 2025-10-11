# summon_homun

A Bubble Tea TUI utility for managing and running Homunculus Docker containers.

## Overview

`summon_homun` provides an interactive terminal interface for setting up and running the Homunculus autonomous Claude Code environment in a Docker container. It handles environment validation, directory setup, permissions, and Docker execution with a clean, user-friendly interface.

## Features

- ASCII art banner with homunculus tagline
- Environment variable validation with clear error messages
- Docker image verification
- Git repository detection
- Interactive prompts with default value suggestions
- Branch name validation (alphanumeric, hyphen, underscore only)
- Automatic directory structure creation
- Permission and ownership management
- Live timer during Docker execution
- Graceful error handling

## Prerequisites

- Docker installed and accessible
- Docker image `homun-dev` available locally
- Running in a git repository
- Required environment variables (see below)

## Required Environment Variables

The following environment variables **must** be set before running:

- `ANTHROPIC_API_KEY` - Anthropic API key for Claude Code
- `HOMUNCULUS_SSH_KEY_PRIVATE` - Private SSH key
- `HOMUNCULUS_SSH_KEY_PUBLIC` - Public SSH key
- `HOMUNCULUS_GITEA_API_KEY` - Gitea API key

## Optional Environment Variables

These can be set or will be prompted for with defaults:

- `HOMUNCULUS_REPO_URL` - Git repository URL (default: origin remote URL)
- `HOMUNCULUS_REPO` - Repository name (default: `$USER/$FOLDER_NAME`)

## Installation

```bash
go install github.com/bharathcs/go-ai-utils/cmd/summon_homun@latest
```

Or build from source:

```bash
go build ./cmd/summon_homun
```

## Usage

Simply run the command in any git repository:

```bash
summon_homun
```

The application will guide you through:

1. **Validation** - Checks Docker image, environment variables, and git repository
2. **Configuration** - Prompts for repository URL and name (with defaults)
3. **Branch Input** - Asks for a branch name (validates alphanumeric + hyphen/underscore)
4. **Directory Setup** - Creates `.homun/branches/$BRANCH/{workspace,report}` directories
5. **Permissions** - Sets ownership to `$USER:homun` with group permissions
6. **Edit Instructions** - Multiline editor for `report/instructions.md` (creates default template if not present)
7. **Confirmation** - Shows configuration summary and asks for confirmation (y/n)
8. **Execution** - Runs Docker container with live timer

## Directory Structure

The tool creates the following structure in the git repository root:

```
.homun/
└── branches/
    └── $BRANCH_NAME/
        ├── workspace/   # Mounted to /workspace in container
        └── report/      # Mounted to /report in container
```

## Docker Command

When you confirm, the tool executes:

```bash
docker run -it --rm \
    --cap-add=NET_ADMIN \
    --cap-add=NET_RAW \
    -e ANTHROPIC_API_KEY \
    -e HOMUNCULUS_GITEA_API_KEY \
    -e HOMUNCULUS_REPO_URL=$HOMUNCULUS_REPO_URL \
    -e HOMUNCULUS_REPO=$HOMUNCULUS_REPO \
    -e HOMUNCULUS_SSH_KEY_PRIVATE \
    -e HOMUNCULUS_SSH_KEY_PUBLIC \
    -v "$BRANCH_DIR/workspace:/workspace" \
    -v "$BRANCH_DIR/report:/report" \
    -w /workspace \
    homun-dev \
    bash -c /home/homun/runner.bash
```

## Keyboard Controls

- **Enter** - Confirm input / Execute selected action
- **Backspace** - Delete character
- **y** - Confirm at confirmation screen
- **n** - Decline at confirmation screen
- **Ctrl+S** - Save instructions and continue (in instruction editor)
- **Ctrl+C** - Exit application
- **ESC** - Exit application (or blur textarea in instruction editor)

## Error Handling

The application provides clear error messages for:

- Missing Docker image
- Missing required environment variables
- Not in a git repository
- Invalid branch names
- Directory creation failures
- Permission setting failures

## Development

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling

## File Location

`cmd/summon_homun/main.go`
