/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	c "spirit-dev/work-facilitator/work-facilitator/common"
	"spirit-dev/work-facilitator/work-facilitator/helper"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// Cmd Args
	workUseArg string

	// local
)

// useCmd represents the use command
var useCmd = &cobra.Command{
	Use:    "use",
	Short:  "Use an existing workflow",
	Long:   "Use an existing workflow",
	PreRun: usePreRunCommand,
	Run:    useCommand,
}

func usePreRunCommand(cmd *cobra.Command, args []string) {
	helper.WelcomeDisplay()
	RootConfig = helper.NewConfig()
	RootRepo = helper.NewRepo(RootConfig)

	log.Debug("pre run use")
	helper.SpinStartDisplay("Verifications - use...")

	// Warn if we are in a workflow
	// if RootRepo.HasCurrentWorkflow {
	// 	log.Warningln("You are in a worklow.")
	// }

	// Ensure the target workflow exists
	_, err := helper.RepoGetWorkflowParam(workUseArg, helper.REFBRANCHPARAM)
	if err != nil {
		helper.SpinStopDisplay("warning")
		log.Warningln("Workflow does not exist.")
		log.Warningln("Ensure it exist using the command:")
		log.Warningln("#> " + RootConfig.ScriptName + " list")
		log.Fatalln(err)
	}

	helper.SpinUpdateDisplay("Verifications")
	helper.SpinStopDisplay("success")
}

func useCommand(cmd *cobra.Command, args []string) {
	log.Debug("run use")
	helper.SpinStartDisplay("Git operations")

	// Checkout
	helper.SpinUpdateDisplay("git checkout " + workUseArg)
	helper.RepoCheckout(workUseArg, RootRepo.PublicAuthKey)
	// Pull
	helper.SpinUpdateDisplay("Git pull")
	pullInfo := helper.RepoPull(RootRepo.PublicAuthKey)

	// Set Current Workflow
	helper.SpinUpdateDisplay("Config update...")
	helper.RepoConfigDefineCurrentWorkflow(workUseArg)
	helper.RepoConfigWrite()

	helper.SpinUpdateDisplay("Git operations")
	helper.SpinStopDisplay("success")
	helper.SpinSideNoteDisplay("Pull info: " + pullInfo)

	latestWf := helper.RepoConfigGetCurrentWorkflow(workUseArg)

	helper.ShowSummary(RootConfig, latestWf)

	// Say GoodBye
	helper.ByeByeDisplay()
}

func init() {
	rootCmd.AddCommand(useCmd)

	helper.Quiet = true // Ensure the first call to newconfig is done quietly
	RootConfig = helper.NewConfig()
	RootRepo := helper.NewRepo(RootConfig)
	helper.Quiet = false // Ensure to reset the value

	var worklistStr string
	for _, w := range RootRepo.Worklist {
		worklistStr += "\t - " + w + "\n"
	}

	useCmd.Flags().StringVarP(&workUseArg, "work", "w", c.NOTGIVEN, "Work to use \n"+worklistStr)
}
