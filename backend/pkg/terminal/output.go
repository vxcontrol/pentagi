package terminal

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/fatih/color"
)

const termColumnWidth = 120

var (
	// Colors for different types of information
	infoColor      = color.New(color.FgCyan)
	successColor   = color.New(color.FgGreen)
	errorColor     = color.New(color.FgRed)
	warningColor   = color.New(color.FgYellow)
	headerColor    = color.New(color.FgBlue, color.Bold)
	keyColor       = color.New(color.FgBlue)
	valueColor     = color.New(color.FgMagenta)
	highlightColor = color.New(color.FgHiMagenta, color.Bold)
	separatorColor = color.New(color.FgWhite)
	mockColor      = color.New(color.FgHiYellow)

	// Predefined prefixes
	infoPrefix    = "[INFO] "
	successPrefix = "[SUCCESS] "
	errorPrefix   = "[ERROR] "
	warningPrefix = "[WARNING] "
	mockPrefix    = "[MOCK] "

	// Separators for output sections
	thinSeparator  = "--------------------------------------------------------------"
	thickSeparator = "=============================================================="
)

func Info(format string, a ...interface{}) {
	infoColor.Printf(format+"\n", a...)
}

func Success(format string, a ...interface{}) {
	successColor.Printf(format+"\n", a...)
}

func Error(format string, a ...interface{}) {
	errorColor.Printf(format+"\n", a...)
}

func Warning(format string, a ...interface{}) {
	warningColor.Printf(format+"\n", a...)
}

// PrintInfo prints an informational message
func PrintInfo(format string, a ...interface{}) {
	infoColor.Printf(infoPrefix+format+"\n", a...)
}

// PrintSuccess prints a success message
func PrintSuccess(format string, a ...interface{}) {
	successColor.Printf(successPrefix+format+"\n", a...)
}

// PrintError prints an error message
func PrintError(format string, a ...interface{}) {
	errorColor.Printf(errorPrefix+format+"\n", a...)
}

// PrintWarning prints a warning
func PrintWarning(format string, a ...interface{}) {
	warningColor.Printf(warningPrefix+format+"\n", a...)
}

// PrintMock prints information about a mock operation
func PrintMock(format string, a ...interface{}) {
	mockColor.Printf(mockPrefix+format+"\n", a...)
}

// PrintHeader prints a section header
func PrintHeader(text string) {
	headerColor.Println(text)
}

// PrintKeyValue prints a key-value pair
func PrintKeyValue(key, value string) {
	keyColor.Printf("%s: ", key)
	fmt.Println(value)
}

// PrintValueFormat prints colored string with formatted value
func PrintValueFormat(format string, a ...interface{}) {
	highlightColor.Printf(format+"\n", a...)
}

// PrintKeyValueFormat prints a key-value pair with formatted value
func PrintKeyValueFormat(key string, format string, a ...interface{}) {
	keyColor.Printf("%s: ", key)
	valueColor.Printf(format+"\n", a...)
}

// PrintThinSeparator prints a thin separating line
func PrintThinSeparator() {
	separatorColor.Println(thinSeparator)
}

// PrintThickSeparator prints a thick separating line
func PrintThickSeparator() {
	separatorColor.Println(thickSeparator)
}

// PrintJSON prints formatted JSON
func PrintJSON(data any) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		PrintError("Failed to format JSON: %v", err)
		return
	}
	fmt.Println(string(jsonBytes))
}

// RenderMarkdown renders markdown text and prints it to the terminal
func RenderMarkdown(markdown string) {
	if len(markdown) == 0 {
		return
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(termColumnWidth),
	)
	if err != nil {
		PrintError("Failed to create markdown renderer: %v", err)
		fmt.Println(markdown)
		return
	}

	out, err := renderer.Render(markdown)
	if err != nil {
		PrintError("Failed to render markdown: %v", err)
		fmt.Println(markdown)
		return
	}

	fmt.Print(out)
}

// InteractivePromptContext prompts the user for input with support for context cancellation
func InteractivePromptContext(ctx context.Context, message string, reader io.Reader) (string, error) {
	// Display the prompt
	infoColor.Printf("%s: ", message)

	// Create a channel for the user input
	inputCh := make(chan string, 1)
	errCh := make(chan error, 1)

	// Start a goroutine to read user input
	go func() {
		// Use a buffered reader to properly handle input
		r, ok := reader.(*os.File)
		if !ok {
			// If it's not a file (e.g., pipe or other reader), use normal scanner
			scanner := bufio.NewScanner(reader)
			if scanner.Scan() {
				inputCh <- strings.TrimSpace(scanner.Text())
			} else if err := scanner.Err(); err != nil {
				errCh <- err
			} else {
				errCh <- io.EOF
			}
			return
		}

		// Create a new reader just for this input to avoid buffering issues
		scanner := bufio.NewScanner(r)
		if scanner.Scan() {
			inputCh <- strings.TrimSpace(scanner.Text())
		} else if err := scanner.Err(); err != nil {
			errCh <- err
		} else {
			errCh <- io.EOF
		}
	}()

	// Wait for input or context cancellation
	select {
	case input := <-inputCh:
		return input, nil
	case err := <-errCh:
		return "", err
	case <-ctx.Done():
		// Context cancelled or timed out
		fmt.Println() // New line after prompt
		return "", ctx.Err()
	}
}

// GetYesNoInputContext prompts the user for a Yes/No input with context support
func GetYesNoInputContext(ctx context.Context, message string, reader io.Reader) (bool, error) {
	for {
		// Check if context is done before prompting
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
			// Continue with prompt
		}

		response, err := InteractivePromptContext(ctx, message+" (y/n)", reader)
		if err != nil {
			return false, err
		}

		switch strings.ToLower(response) {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			PrintWarning("Please enter 'y' or 'n'")
		}
	}
}

// IsMarkdownContent checks if the input string is likely markdown content
func IsMarkdownContent(content string) bool {
	// Determine if content is likely markdown by checking for common indicators
	if strings.HasPrefix(content, "#") ||
		strings.Contains(content, "\n#") ||
		strings.Contains(content, "[") && strings.Contains(content, "](") ||
		strings.Contains(content, "```") ||
		strings.Contains(content, "**") ||
		strings.Contains(content, "- ") && strings.Contains(content, "\n- ") {
		return true
	}
	return false
}

// PrintResult prints a result string that might be in markdown format
func PrintResult(result string) {
	if IsMarkdownContent(result) {
		RenderMarkdown(result)
	} else {
		fmt.Println(result)
	}
}

// PrintResultWithKey prints a key and a result that might be in markdown format
func PrintResultWithKey(key, result string) {
	keyColor.Printf("%s:\n", key)
	PrintThinSeparator()
	PrintResult(result)
}
