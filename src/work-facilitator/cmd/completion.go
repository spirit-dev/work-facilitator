/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"spirit-dev/work-facilitator/work-facilitator/helper"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	completionArgs = []string{
		"bash",
		"zsh",
		"fish",
		"powershell",
	}
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Create auto-completion",
	Long: `To load completions:

Bash:

  $ source <(work-flow completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ work-flow completion bash > /etc/bash_completion.d/work-flow
  # macOS:
  $ work-flow completion bash > $(brew --prefix)/etc/bash_completion.d/work-flow

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ work-flow completion zsh > "${fpath[1]}/_work-flow"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ work-flow completion fish | source

  # To load completions for each session, execute once:
  $ work-flow completion fish > ~/.config/fish/completions/work-flow.fish

PowerShell:

  PS> work-flow completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> work-flow completion powershell > work-flow.ps1
  # and source this file from your PowerShell profile.`,
	DisableFlagsInUseLine: true,
	ValidArgs:             completionArgs,
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run:                   completionRunCommand,
}

func completionRunCommand(cmd *cobra.Command, args []string) {

	// Some display
	log.Debug("run end")
	helper.SpinStartDisplay("Generate completion")

	var fileName string

	switch args[0] {
	case "bash":
		fileName = "/tmp/.wf-comp-bash"
		cmd.Root().GenBashCompletionFile(fileName)
		helper.SpinStopDisplay("success")
		helper.SpinSideNoteDisplay("file created " + fileName)
	case "zsh":
		// initialize file
		fileName = "/tmp/.wf-comp-zsh"
		cmd.Root().GenZshCompletionFile(fileName)
		helper.SpinStopDisplay("success")
		helper.SpinSideNoteDisplay("file created " + fileName)
	case "fish":
		fileName = "/tmp/.wf-comp-fish"
		cmd.Root().GenFishCompletionFile(fileName, true)
		helper.SpinStopDisplay("success")
		helper.SpinSideNoteDisplay("file created " + fileName)
	case "powershell":
		fileName = "/tmp/.wf-comp-pw"
		cmd.Root().GenPowerShellCompletionFile(fileName)
		helper.SpinStopDisplay("success")
		helper.SpinSideNoteDisplay("file created " + fileName)
	}

	// Add alias
	aliases := []string{
		"",
		"# Alias definition",
		"alias wf='work-facilitator'",
		"alias wfc='work-facilitator commit'",
		"alias wfe='work-facilitator end'",
		"alias wfi='work-facilitator init'",
		"alias wfil='work-facilitator initLazy'",
		"alias wfl='work-facilitator list'",
		"alias wfs='work-facilitator status'",
		"alias wfu='work-facilitator use'",
		"alias wfo='work-facilitator open'",
		"alias wfp='work-facilitator pause'",
	}
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	if _, err = f.WriteString(strings.Join(aliases, "\n")); err != nil {
		log.Fatalln(err)
	}
}

func init() {
	rootCmd.AddCommand(completionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// completionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// completionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
