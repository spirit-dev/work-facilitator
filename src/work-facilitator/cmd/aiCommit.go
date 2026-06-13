/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"spirit-dev/work-facilitator/work-facilitator/ai"
	"spirit-dev/work-facilitator/work-facilitator/helper"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// Cmd args
	noPushAICommitArg        bool
	allFilesAICommitArg      bool
	forceAICommitArg         bool
	skipPreCommitAICommitArg bool
	includeUnstagedAICommitArg bool

	// local variables
	commitMessageAICommit string
)

// aiCommitCmd represents the ai-commit command
var aiCommitCmd = &cobra.Command{
	Use:   "ai-commit [flags]",
	Short: "AI-assisted commit workflow",
	Long: `Commit work in current branch with AI-generated commit message.

The AI will analyze your staged changes and generate a meaningful commit message.
You can review, edit, or regenerate the message before committing.`,
	PreRun: aiCommitPreRunCommand,
	Run:    aiCommitCommand,
}

func aiCommitPreRunCommand(cmd *cobra.Command, args []string) {
	helper.WelcomeDisplay()
	RootConfig = helper.NewConfig()
	RootRepo = helper.NewRepo(RootConfig)

	log.Debug("pre run ai-commit")
	helper.SpinStartDisplay("Verifications - ai-commit...")

	// Check if AI is enabled
	if !RootConfig.AIEnabled {
		helper.SpinStopDisplay("fail")
		log.Warningln("AI features are not enabled in configuration.")
		log.Warningln("Set 'ai.enabled: true' in ~/.workflow.yaml")
		os.Exit(1)
	}

	// Check if in workflow (unless forced)
	if !forceAICommitArg && !RootRepo.HasCurrentWorkflow {
		helper.SpinStopDisplay("fail")
		log.Warningln("We are not in a current workflow.")
		log.Warningln("You can force the commit/push by running with -f flag")
		os.Exit(1)
	}

	helper.SpinUpdateDisplay("Verifications")
	helper.SpinStopDisplay("success")
}

func aiCommitCommand(cmd *cobra.Command, args []string) {
	log.Debug("run ai-commit")

	// Stage files if requested
	if allFilesAICommitArg {
		helper.SpinStartDisplay("Git add all files")
		helper.RepoAddAllFiles(RootConfig.CommitIgnorePatternsCompiled)
		helper.SpinStopDisplay("success")
	}

	// Get diff: staged only (index blobs) by default, or working tree with -U flag
	helper.SpinStartDisplay("Analyzing staged changes")
	var diff string
	var diffErr error
	if includeUnstagedAICommitArg {
		diff, diffErr = helper.GetWorkingTreeDiff()
	} else {
		diff, diffErr = helper.GetStagedDiff()
	}
	if diffErr != nil {
		helper.SpinStopDisplay("fail")
		log.Fatalln("Failed to get staged diff:", diffErr)
	}

	// Filter diff by exclude patterns
	if len(RootConfig.AIExcludePatterns) > 0 {
		diff = helper.FilterDiffByPatterns(diff, RootConfig.AIExcludePatterns)
	}

	helper.SpinStopDisplay("success")

	// Run pre-commit hooks (unless skipped)
	if !skipPreCommitAICommitArg {
		if err := helper.RunPreCommitHooks(); err != nil {
			log.Fatalln("Commit aborted due to pre-commit hook failure")
		}
	}

	// Create AI provider once (shared for initial generation and potential regeneration)
	timeout := time.Duration(RootConfig.AITimeout) * time.Second
	aiProvider := createAIProvider(timeout)

	// Prepare generation options
	options := &ai.GenerateOptions{
		MaxTokens:   RootConfig.AIMaxTokens,
		Temperature: RootConfig.AITemperature,
		BranchName:  RootRepo.CurrentWorkflowData.Branch,
	}
	if RootConfig.EnforceStandard {
		options.CommitStandard = RootConfig.CommitExpr
	}

	// Generate initial commit message
	helper.SpinStartDisplay("Generating commit message with AI")
	charCount, tokenEstimate := ai.GetPromptMetrics(diff, options)
	fmt.Printf("Model: %s | Context size: %d chars (~%d tokens)\n", RootConfig.AIModel, charCount, tokenEstimate)
	commitMessageAICommit = doGenerateMessage(aiProvider, diff, options, timeout)
	helper.SpinStopDisplay("success")

	// Interactive review with regeneration support
	finalMessage := reviewAndEditMessage(commitMessageAICommit, aiProvider, diff, options, timeout)

	// Validate against commit standard if enforced
	if RootConfig.EnforceStandard {
		preMessageCommit := ""
		if RootRepo.HasCurrentWorkflow {
			preMessageCommit = RootRepo.CurrentWorkflowData.Commit
		}
		activeBranch := RootRepo.CurrentWorkflowData.Branch

		fullMessage := fmt.Sprintf("%s%s", preMessageCommit, finalMessage)
		if !helper.TestStandard(fullMessage, RootConfig.CommitExpr, activeBranch, RootConfig.BranchExpr, RootConfig.EnforceStandard) {
			log.Fatalln("Commit message does not comply with standard")
		}
		finalMessage = fullMessage
	}

	// Perform commit
	helper.SpinStartDisplay("Git operations")
	helper.SpinUpdateDisplay("Git commit")
	helper.RepoCommit(finalMessage, RootConfig.CommitIgnorePatternsCompiled)

	// git push
	if !noPushAICommitArg {
		helper.SpinUpdateDisplay("Git push")
		helper.RepoPush(RootRepo.PublicAuthKey, RootRepo.CurrentWorkflowData.Branch)
	}

	helper.SpinUpdateDisplay("Git operations")
	helper.SpinStopDisplay("success")

	if !noPushAICommitArg {
		helper.SpinSideNoteDisplay("git push origin")
	}

	// Say GoodBye
	helper.ByeByeDisplay()
}

// createAIProvider creates the configured AI provider from RootConfig.
func createAIProvider(timeout time.Duration) ai.Provider {
	switch RootConfig.AIProvider {
	case "openai":
		return ai.NewOpenAIProvider(RootConfig.AIAPIKey, RootConfig.AIModel, timeout)
	case "claude":
		return ai.NewClaudeProvider(RootConfig.AIAPIKey, RootConfig.AIModel, timeout)
	case "llamacpp":
		return ai.NewLlamaCPPProvider(RootConfig.AIBaseURL, RootConfig.AIAPIKey, RootConfig.AIModel, timeout)
	case "vertexai":
		provider, err := ai.NewVertexAIProvider(
			RootConfig.AIGoogleProjectID,
			RootConfig.AIGoogleLocation,
			RootConfig.AIGoogleServiceAccountKey,
			RootConfig.AIModel,
			timeout,
		)
		if err != nil {
			log.Fatalln("Failed to create Vertex AI provider:", err)
		}
		return provider
	default:
		log.Fatalln("Unknown AI provider:", RootConfig.AIProvider)
		return nil
	}
}

// doGenerateMessage calls the AI provider to generate a commit message.
// Falls back to manual entry on failure.
func doGenerateMessage(provider ai.Provider, diff string, options *ai.GenerateOptions, timeout time.Duration) string {
	// Validate provider
	if err := provider.Validate(); err != nil {
		log.Warningln("AI provider validation failed:", err)
		log.Warningln("Falling back to manual message entry")
		return promptForMessage()
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	message, err := provider.GenerateCommitMessage(ctx, diff, options)
	if err != nil {
		log.Warningln("AI generation failed:", err)
		log.Warningln("Falling back to manual message entry")
		return promptForMessage()
	}

	return message
}

// generateAICommitMessage is kept for backward compatibility (initial generation only).
func generateAICommitMessage(diff string) string {
	helper.SpinStartDisplay("Generating commit message with AI")

	timeout := time.Duration(RootConfig.AITimeout) * time.Second
	aiProvider := createAIProvider(timeout)

	options := &ai.GenerateOptions{
		MaxTokens:   RootConfig.AIMaxTokens,
		Temperature: RootConfig.AITemperature,
		BranchName:  RootRepo.CurrentWorkflowData.Branch,
	}
	if RootConfig.EnforceStandard {
		options.CommitStandard = RootConfig.CommitExpr
	}

	charCount, tokenEstimate := ai.GetPromptMetrics(diff, options)
	fmt.Printf("Model: %s | Context size: %d chars (~%d tokens)\n", RootConfig.AIModel, charCount, tokenEstimate)

	message := doGenerateMessage(aiProvider, diff, options, timeout)
	helper.SpinStopDisplay("success")
	return message
}

func reviewAndEditMessage(initialMessage string, provider ai.Provider, diff string, options *ai.GenerateOptions, timeout time.Duration) string {
	message := initialMessage

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("=== AI Generated Commit Message ===")
		fmt.Println(message)
		fmt.Println("===================================")

		fmt.Print("Options: [A]ccept, [E]dit, [R]egenerate, [C]ancel? ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(strings.ToLower(choice))

		switch choice {
		case "a", "accept":
			return message
		case "e", "edit":
			return editMessage(message)
		case "r", "regenerate":
			helper.SpinStartDisplay("Regenerating commit message")
			newMsg := doGenerateMessage(provider, diff, options, timeout)
			helper.SpinStopDisplay("success")
			if newMsg != "" {
				message = newMsg
			}
			// Loop continues with new message displayed
		case "c", "cancel":
			log.Fatalln("Commit cancelled by user")
		default:
			fmt.Println("Invalid choice. Please enter A, E, R, or C.")
		}
	}
}

func editMessage(initialMessage string) string {
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "commit-msg-*.txt")
	if err != nil {
		log.Fatalln("Failed to create temp file:", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write initial message
	if _, err := tmpFile.WriteString(initialMessage); err != nil {
		log.Fatalln("Failed to write to temp file:", err)
	}
	tmpFile.Close()

	// Open in editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi" // Default to vi
	}

	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalln("Failed to open editor:", err)
	}

	// Read edited message
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		log.Fatalln("Failed to read edited message:", err)
	}

	return strings.TrimSpace(string(content))
}

func promptForMessage() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter commit message: ")
	message, _ := reader.ReadString('\n')
	return strings.TrimSpace(message)
}

func init() {
	rootCmd.AddCommand(aiCommitCmd)

	aiCommitCmd.Flags().BoolVarP(&noPushAICommitArg, "no-push", "n", false, "Activate option to avoid pushing commits")
	aiCommitCmd.Flags().BoolVarP(&allFilesAICommitArg, "all-files", "a", false, "Stage all modified files before commit")
	aiCommitCmd.Flags().BoolVarP(&forceAICommitArg, "force-commit", "f", false, "Force the commit if we are not in a workflow")
	aiCommitCmd.Flags().BoolVarP(&skipPreCommitAICommitArg, "skip-precommit", "s", false, "Skip pre-commit hooks")
	aiCommitCmd.Flags().BoolVarP(&includeUnstagedAICommitArg, "include-unstaged", "U", false, "Include working-tree modifications for staged files in the diff")

	aiCommitCmd.Flags().SortFlags = false
}
