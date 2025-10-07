# idk - AI Command Helper
 
## NAME
 
*idk* - Interactive AI-powered command-line helper
 
## SYNOPSIS
 
 idk
 [command] | idk
 
## DESCRIPTION
 
*idk* is an interactive TUI (Terminal User Interface) application that helps you find command-line solutions using AI. When you don't know how to accomplish something in the terminal, just type idk and describe what you want to do in natural language.
 
The application queries an AI assistant (ChatGPT) to provide you with ranked command-line solutions that you can select and execute directly.
 
*idk* also supports piped input, allowing you to provide context from other commands (error messages, logs, file contents, etc.) that will be included with your query.
 
## REQUIREMENTS
 
Before using *idk*, you must set the AI_FACTORY_API_KEY environment variable:
 
 bash
 export AI_FACTORY_API_KEY="your-api-key-here"
 
Optionally, you can also configure:
 
 bash
 export OPENAI_BASE_URL="https://custom-endpoint.com"  # For custom endpoints
 export OPENAI_MODEL="gpt-4o"                          # Specific model (default: gpt-4o)
 
## USAGE
 
### Starting the Application
 
Simply run:
 
 bash
 idk
 
The TUI will open with a prompt ready for your input.
 
### Using with Piped Input
 
Pipe content from other commands to provide context:
 
 bash
 cat error.log | idk
 go build 2>&1 | idk
 curl https://api.example.com/data | idk
 
The piped content is automatically:
 
 Truncated to 400 characters (prevents token overload)
 Displayed as a preview in the history
 Included as context with every query you make
 
This is especially useful for:
 
 *Error messages*: command-that-failed 2>&1 | idk "how do I fix this?"
 *Log analysis*: tail -n 50 app.log | idk "what's wrong here?"
 *File inspection*: cat config.yml | idk "validate this configuration"
 *Command output*: ls -la | idk "how do I filter this to show only directories?"
 
### Basic Workflow
 
 *Enter Your Query*
 Type your question in natural language
 Example: "sed command to strip all lines that do not contain FOO or BAZ"
 Press Enter to submit
 
2. *Wait for Solutions*
 A loading indicator appears while the AI processes your query
 Your prompt is recorded in the history with a > prefix
 
3. *Review Solutions*
 Up to 3 solutions are displayed, ranked by usefulness
 Solutions are numbered (1, 2, 3)
 Each solution shows:
 The command on the first line
 A brief description below in gray text
 
4. *Select and Execute*
 *With Arrow Keys*: Use ↑ and ↓ to navigate between solutions
 Selected solution is highlighted with a ► marker
 Press Enter to execute the selected command
 The command will be typed out character-by-character on your terminal
 
5. *Continue Conversation*
 *Without Arrow Keys*: Simply start typing to continue the conversation
 Your new input becomes a follow-up query
 The AI considers the context of previous queries
 Press Enter to submit the new query
 
## KEYBINDINGS
 
| Key | Action |
|-----|--------|
| Enter | Submit query (in input mode) / Execute command (when solution selected) |
| ↑ | Navigate to previous solution |
| ↓ | Navigate to next solution |
| Backspace | Delete character (in input mode) / Return to input mode (in solution view) |
| Esc | Exit application |
| Ctrl+C | Exit application |
| Any character | Start typing / Continue conversation |
 
## MODES
 
### Input Mode
 
 Default mode when application starts
 Prompt shows: > █
 Type your natural language query
 Press Enter to submit
 
### Loading Mode
 
 Activated after submitting a query
 Shows animated spinner: ⠋ Loading...
 Waits for AI response
 
### Solution Selection Mode
 
 Displays up to 3 ranked solutions
 Two interaction paths:
 *Navigate mode*: Use arrow keys to select, then Enter to execute
 *Continue mode*: Start typing to add more context or refine your query
 
## EXAMPLES
 
### Example 1: Finding Files
 
 > find all .go files modified in the last week
 
*Solutions:*
 
 find . -name "*.go" -mtime -7
   Find Go files modified in the last 7 days
 
2. find . -type f -name "*.go" -mtime -7 -ls
   Find and list details of Go files modified in the last week
 
3. fd -e go -t f --changed-within 7d
   Use fd tool for faster search with modern syntax
 
### Example 2: Text Processing
 
 > sed command to strip all lines that do not contain FOO or BAZ
 
*Solutions:*
 
1. grep -E "(FOO|BAZ)" file.txt
   Use grep to filter lines containing FOO or BAZ (most efficient)
 
2. sed -n '/FOO\|BAZ/p' file.txt
   Use sed to print only matching lines
 
3. awk '/FOO|BAZ/' file.txt
   Use awk to print lines matching the pattern
 
### Example 3: Follow-up Query
 
After receiving solutions, if you're unsure:
 
 > explain the first option
 
The AI will provide an explanation in the context of your previous query.
 
### Example 4: Using Piped Input
 
Pipe error messages or logs for context-aware help:
 
 bash
 go build 2>&1 | idk
 
The TUI opens showing:
 
 :clipboard: Context from stdin: ./main.go:42:15: undefined: myFunc...
 
 > how do I fix this error?
 
*Solutions:*
 
1. go mod tidy && go build
   Ensure dependencies are resolved and rebuild
 
2. grep -r "func myFunc" .
   Search for the function definition in your codebase
 
3. go list -m all | grep myFunc
   Check if myFunc is from an external package
 
## EXIT STATUS
 
 *0*: Successful execution
 *1*: Error occurred during execution
 
## CONFIGURATION
 
Configuration is handled through environment variables (see REQUIREMENTS section above). The application uses the go-utils/pkg/ai library for AI interactions, which supports:
 
 *AI_FACTORY_API_KEY*: Required API key for authentication
 *OPENAI_BASE_URL*: Optional custom endpoint URL
 *OPENAI_MODEL*: Optional model selection (defaults to gpt-4o)
 
The application maintains conversation context across multiple queries within a session, allowing for follow-up questions and refinements.
 
## NOTES
 
 When a solution is executed, it is typed out character-by-character into your terminal
 The screen clears before typing out the command
 You can interrupt the typing animation with Ctrl+C if needed
 History is maintained within a session but not persisted across invocations
 Piped input is truncated to 400 characters by default to avoid token limits
 The piped context is included with every query in the session
 When piped input is provided, the app reopens /dev/tty for interactive input
 
## ERROR HANDLING
 
*idk* includes comprehensive error handling for various scenarios:
 
### Configuration Errors
 
 *Missing API Key*: Clear warning message displayed at startup if AI_FACTORY_API_KEY is not set
 *Invalid Configuration*: Validation of environment variables before making API calls
 
### Network Errors
 
 *Timeout Protection*: Requests timeout after 30 seconds to prevent hanging
 *Connection Issues*: Clear messaging when network connectivity fails
 *Rate Limiting*: Friendly message when API rate limits are exceeded
 
### API Errors
 
 *Authentication Failures*: Specific guidance to check API key
 *Server Errors*: Informative messages for 500/502/503 errors
 *Empty Responses*: Validation that AI provided a meaningful response
 
### Response Validation
 
 *Empty Solutions*: Detection and handling of empty or invalid command suggestions
 *Malformed Responses*: Parsing errors handled gracefully with fallback to raw response
 *Solution Validation*: Each solution is trimmed and validated before display
 
### User Experience
 
 All errors are displayed in red with clear, actionable messages
 The application remains in input state after errors, allowing retry
 Context is maintained across error recovery
 No crashes or unexpected exits from error conditions
 
## IMPLEMENTATION DETAILS
 
 Built with Bubble Tea TUI framework
 Uses Lip Gloss for styling
 AI integration via OpenAI API with *structured outputs* for reliable parsing
 Uses go-utils/pkg/ai structured query functionality for type-safe responses
 Each solution includes:
 *Command*: The exact command to execute
 *Description*: Brief explanation of what it does
 *Usefulness*: Ranking (1-3) to help prioritize solutions
 
