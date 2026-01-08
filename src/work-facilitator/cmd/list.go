/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"spirit-dev/work-facilitator/work-facilitator/helper"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List workflows",
	Long:  `List available or paused workflows`,
	Run:   listCommand,
}

func listCommand(cmd *cobra.Command, args []string) {
	helper.WelcomeDisplay()
	RootConfig = helper.NewConfig()
	RootRepo = helper.NewRepo(RootConfig)

	log.Debug("run list")

	if RootRepo.HasCurrentWorkflow && RootRepo.CurrentWorkflowName != "" {
		helper.Addline("Current workflow\n")
		helper.SpinSideNoteDisplay(RootRepo.CurrentWorkflowName)
	}

	if len(RootRepo.Worklist) > 0 {
		helper.Addline("List of available workflows\n")
		for _, w := range RootRepo.Worklist {
			helper.SpinSideNoteDisplay(w)
		}
	} else {
		log.Infoln("There is now workflow initiated yet.")
		log.Infoln(`You may run :
#> ` + RootConfig.ScriptName + ` init`)
	}

	// Say GoodBye
	helper.ByeByeDisplay()
}

func init() {
	rootCmd.AddCommand(listCmd)
}
