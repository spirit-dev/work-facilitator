/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"spirit-dev/work-facilitator/work-facilitator/helper"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:    "status",
	Short:  "Display the workflow status",
	Long:   "Display usefull information about the current workflow",
	PreRun: statusPreRunCommand,
	Run:    statusCommand,
}

func statusPreRunCommand(cmd *cobra.Command, args []string) {
	helper.WelcomeDisplay()
	RootConfig = helper.NewConfig()
	RootRepo = helper.NewRepo(RootConfig)

	log.Debug("pre run status")
	helper.SpinStartDisplay("Verifications - status...")

	if !RootRepo.HasCurrentWorkflow {
		helper.SpinStopDisplay("warning")
		log.Warningln("No current workflow set up")
		log.Warningln("Please use:")
		log.Warningln("#> " + RootConfig.ScriptName + " use")
		os.Exit(1)
	}

	helper.SpinUpdateDisplay("Verifications")
	helper.SpinStopDisplay("success")
}
func statusCommand(cmd *cobra.Command, args []string) {
	log.Debug("run status")

	helper.SpinStartDisplay("Git operations")

	status := helper.RepoStatus()

	helper.SpinUpdateDisplay("Git operations")
	helper.SpinStopDisplay("success")

	helper.ShowSummary(RootConfig, RootRepo.CurrentWorkflowData)
	helper.ShowBox(status)

	// Say GoodBye
	helper.ByeByeDisplay()
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
