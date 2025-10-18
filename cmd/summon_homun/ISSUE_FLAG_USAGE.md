# Issue Flag Usage Guide

## Overview

The `summon_homun` tool now supports the `--issue` flag, which allows you to automatically fetch issue content from GitHub or Gitea and pre-fill it into the instructions editor. This feature is designed for CLI users who want to quickly address repository issues without manually copying issue details.

## Usage

```bash
summon_homun --issue <ISSUE_NUMBER>
```

### Example

```bash
# Fetch and use GitHub issue #42
summon_homun --issue 42

# Fetch and use Gitea issue #7
summon_homun --issue 7
```

## How It Works

1. **Platform Detection**: The tool automatically detects whether your repository is hosted on GitHub or Gitea by examining the git remote URL.

2. **Issue Fetching**:
   - For GitHub: Uses the `gh` CLI tool
   - For Gitea: Uses the `tea` CLI tool

3. **Instructions Pre-filling**: The fetched issue content is automatically inserted into the instructions template:

```markdown
# Instructions for Homunculus

Fix the following issue:

<ISSUE CONTENT IS PASTED HERE>
```

4. **Editor Opens**: The instruction editor opens with the cursor positioned to add any additional context before the issue content.

## Prerequisites

### For GitHub Repositories

- Install the GitHub CLI: https://cli.github.com/
- Authenticate with `gh auth login`

```bash
# Install gh (example for Linux)
sudo apt install gh

# Or using Homebrew
brew install gh

# Authenticate
gh auth login
```

### For Gitea Repositories

- Install the Gitea CLI: https://gitea.com/gitea/tea
- Configure tea with your Gitea instance

```bash
# Install tea
go install gitea.com/gitea/tea@latest

# Or download binary from releases
# https://gitea.com/gitea/tea/releases

# Login to your Gitea instance
tea login add
```

## CLI Tool Detection

The issue fetching process includes the following commands:

### GitHub (using `gh`)
```bash
gh issue view <ISSUE_NUMBER> --repo <owner/repo> --comments
```

### Gitea (using `tea`)
```bash
tea issue view <ISSUE_NUMBER> --repo <owner/repo>
```

## Implementation Details

### Files Modified/Created

1. **cmd/summon_homun/main.go**
   - Added flag parsing for `--issue`
   - Added issue fetching before model initialization
   - Passes issue content to the model

2. **cmd/summon_homun/model.go**
   - Added `issueContent` field to the model struct

3. **cmd/summon_homun/instructions.go**
   - Added `issueInstructionsTemplate` constant
   - Modified `loadInstructions()` to use issue template when issue content is provided

4. **cmd/summon_homun/issue.go** (NEW)
   - `fetchIssue()`: Main function to fetch issues
   - `isGitHub()` / `isGitea()`: Platform detection
   - `extractRepoFromURL()`: Extracts owner/repo from git URL
   - `fetchGitHubIssue()`: Fetches using `gh` CLI
   - `fetchGiteaIssue()`: Fetches using `tea` CLI
   - `getGitRemoteURL()`: Gets git remote origin URL

5. **cmd/summon_homun/issue_test.go** (NEW)
   - Unit tests for URL parsing and platform detection

## Error Handling

The tool provides clear error messages for common issues:

- **CLI tool not installed**: Suggests installation with helpful links
- **Failed to fetch issue**: Shows the error from the CLI tool
- **Invalid repository URL**: Reports parsing errors
- **Not in a git repository**: Reminds user to run from within a repo

## Configuration

The issue fetching respects the existing configuration in `~/.config/homun/config.yml`:

```yaml
github:
  username: your_github_username
  base_url: https://github.com

gitea:
  username: your_gitea_username
  base_url: https://gitea.example.com
```

The tool uses this configuration to:
- Determine which platform to use (GitHub vs Gitea)
- Construct proper repository references

## Workflow Example

```bash
# List issues in your repo
gh issue list
# or
tea issues

# Pick an issue number, say #42
summon_homun --issue 42

# The tool will:
# 1. Fetch issue #42 details
# 2. Pre-fill the instructions editor
# 3. You can add more context if needed
# 4. Save with Ctrl+S
# 5. Confirm and run Homunculus
```

## Benefits

- **No manual copying**: Issue details are automatically fetched
- **Full context**: Includes issue title, body, and comments
- **Consistent format**: Always uses the same template structure
- **Quick workflow**: Reduces time from issue to fix
- **CLI-friendly**: Perfect for terminal-based workflows

## Troubleshooting

### "gh CLI is not installed"
Install the GitHub CLI from https://cli.github.com/ and authenticate with `gh auth login`.

### "tea CLI is not installed"
Install the Gitea CLI from https://gitea.com/gitea/tea and configure it with `tea login add`.

### "failed to get git remote URL"
Make sure you're running the command from within a git repository that has a remote origin configured.

### "unsupported git hosting platform"
Currently, only GitHub and Gitea are supported. If you need support for another platform, please open an issue.

## Future Enhancements

Possible future improvements:
- Support for GitLab using `glab` CLI
- Support for issue URLs directly (not just numbers)
- Caching of fetched issues
- Offline mode with cached issues
- Support for pull requests via `--pr` flag
