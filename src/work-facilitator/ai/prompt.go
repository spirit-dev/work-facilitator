/*
Copyright © 2024 Jean Bordat bordat.jean@gmail.com
*/
package ai

import (
	"fmt"
	"strings"
)

// buildPrompt creates the prompt for the AI based on diff and options
func buildPrompt(diff string, options *GenerateOptions) string {
	var prompt strings.Builder

	prompt.WriteString("Task: Generate a git commit message for the provided diff.\n\n")
	prompt.WriteString("Strict Constraints:\n")
	prompt.WriteString("1. Output ONLY the raw message. No markdown, no quotes, no conversational filler.\n")
	prompt.WriteString("2. Format: Single line, under 72 characters.\n")
	prompt.WriteString("3. Content: Start directly with the action verb (e.g., 'update', 'fix', 'add').\n")
	prompt.WriteString("4. FORBIDDEN: Do NOT use prefixes like 'feat:', 'fix:', 'docs:', or 'refactor(scope):'.\n")
	prompt.WriteString("5. Character Set: Alphanumeric and spaces ONLY. Absolutely NO punctuation (no periods, no colons, no hyphens, no parentheses) and NO special symbols.\n\n")
	prompt.WriteString("Input Diff:\n")
	prompt.WriteString("```diff\n")
	prompt.WriteString(diff)
	prompt.WriteString("\n```\n\n")

	if options.BranchName != "" {
		prompt.WriteString(fmt.Sprintf("Current branch: %s\n", options.BranchName))
	}

	if options.AdditionalContext != "" {
		prompt.WriteString(fmt.Sprintf("Additional context: %s\n", options.AdditionalContext))
	}

	// prompt.WriteString("Generate a concise git commit message for the following changes:\n\n")
	// prompt.WriteString("```diff\n")
	// prompt.WriteString(diff)
	// prompt.WriteString("\n```\n\n")

	// if options.BranchName != "" {
	// 	prompt.WriteString(fmt.Sprintf("Current branch: %s\n", options.BranchName))
	// }

	// if options.AdditionalContext != "" {
	// 	prompt.WriteString(fmt.Sprintf("Additional context: %s\n", options.AdditionalContext))
	// }

	// prompt.WriteString("\nIMPORTANT FORMATTING REQUIREMENTS:\n")
	// prompt.WriteString("- Provide ONLY the commit message as a single line of plain text\n")
	// prompt.WriteString("- Do NOT use any prefixes like 'feat:', 'fix:', 'refactor:', 'chore:', 'docs:', etc.\n")
	// prompt.WriteString("- Do NOT use backticks, quotes, asterisks, or any markdown formatting\n")
	// prompt.WriteString("- Do NOT use special characters like colons, parentheses, brackets, or emojis\n")
	// prompt.WriteString("- Do NOT include any explanation, context, or additional text\n")
	// prompt.WriteString("- Start directly with the action verb (e.g., 'Add', 'Update', 'Remove', 'Fix', 'Enhance')\n")
	// prompt.WriteString("- Keep it simple and descriptive\n\n")
	// prompt.WriteString("Example of CORRECT format: 'Add user authentication to login page'\n")
	// prompt.WriteString("Example of INCORRECT format: 'feat: Add user authentication to login page'\n")
	// prompt.WriteString("Example of INCORRECT format: 'Add `user authentication` to login page'\n")

	return prompt.String()
}
