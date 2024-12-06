/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	c "spirit-dev/work-facilitator/work-facilitator/common"
	"spirit-dev/work-facilitator/work-facilitator/helper"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// Cmd Args
	masterPauseArg string

	// local
	checkoutToPause string
)

// pauseCmd represents the pause command
var pauseCmd = &cobra.Command{
	Use:    "pause",
	Short:  "Pause current workflow",
	Long:   `Put current work on hold, without cleanup`,
	PreRun: pausePreRunCommand,
	Run:    pauseCommand,
}

func pausePreRunCommand(cmd *cobra.Command, args []string) {
	helper.WelcomeDisplay()
	RootConfig = helper.NewConfig()
	RootRepo = helper.NewRepo(RootConfig)

	log.Debug("pre run pause")
	helper.SpinStartDisplay("Verifications - pause...")

	// Ensure we are in a workflow
	if !RootRepo.HasCurrentWorkflow {
		helper.SpinStopDisplay("warning")
		log.Warningln("No current workflow set up")
		log.Warningln("Please use:")
		log.Warningln("#> " + RootConfig.ScriptName + " use")
		os.Exit(1)
	}

	// Define
	if masterPauseArg != c.GOMASTER {
		checkoutToPause = masterPauseArg
	} else {
		checkout, err := helper.RepoGetWorkflowParam(RootRepo.CurrentWorkflowName, helper.REFBRANCHPARAM)
		if err != nil {
			helper.SpinStopDisplay("fail")
			log.Fatalln(err)
		}
		checkoutToPause = checkout
	}

	helper.SpinUpdateDisplay("Verifications")
	helper.SpinStopDisplay("success")
}

func pauseCommand(cmd *cobra.Command, args []string) {
	log.Debug("run pause")
	helper.SpinStartDisplay("Git operations")

	// Checkout
	helper.SpinUpdateDisplay("git checkout " + checkoutToPause)
	helper.RepoCheckout(checkoutToPause, RootRepo.PublicAuthKey)
	// Pull
	helper.SpinUpdateDisplay("Git pull")
	pullInfo := helper.RepoPull(RootRepo.PublicAuthKey)

	// Delete current workflow
	helper.SpinUpdateDisplay("Config update...")
	helper.RepoConfigDeleteCurrentWorkflow()
	helper.RepoConfigWrite()

	helper.SpinUpdateDisplay("Git operations")
	helper.SpinStopDisplay("success")
	helper.SpinSideNoteDisplay("Pull info: " + pullInfo)

	// Say GoodBye
	helper.ByeByeDisplay()
}

func init() {
	rootCmd.AddCommand(pauseCmd)

	pauseCmd.Flags().StringVarP(&masterPauseArg, "master", "m", c.GOMASTER, "Go master branch rather than ref-branch")

}
