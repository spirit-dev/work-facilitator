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
	// cmd Args
	workEndArg string

	// local variables
	workToDeleteEnd string
	currentEnd      bool
)

// endCmd represents the end command
var endCmd = &cobra.Command{
	Use:    "end",
	Short:  "Finalize workflow",
	Long:   `Finalize a work done`,
	PreRun: endPreRunCommand,
	Run:    endCommand,
}

func endPreRunCommand(cmd *cobra.Command, args []string) {
	helper.WelcomeDisplay()
	RootConfig = helper.NewConfig()
	RootRepo = helper.NewRepo(RootConfig)

	log.Debug("pre run end")
	helper.SpinStartDisplay("Verifications - end...")

	log.Debugf("RootRepo.CurrentWorkflowName: %v\n", RootRepo.CurrentWorkflowName)
	log.Debugf("RootRepo.HasCurrentWorkflow: %v\n", RootRepo.HasCurrentWorkflow)

	if !RootRepo.HasCurrentWorkflow && workEndArg == c.NOTGIVEN {
		helper.SpinStopDisplay("fail")
		log.Fatalln("No current workflow set, or no -w arg given")
	}

	// Prepare values based on given values
	if RootRepo.HasCurrentWorkflow {
		workToDeleteEnd = RootRepo.CurrentWorkflowName
		currentEnd = true
	}
	if workEndArg != c.NOTGIVEN {
		if RootRepo.HasCurrentWorkflow && (workEndArg == RootRepo.CurrentWorkflowName) {
			currentEnd = true
		} else {
			currentEnd = false
		}
		workToDeleteEnd = workEndArg
	}
	log.Debugf("workEndArg: %v\n", workEndArg)
	log.Debugf("helper.NOTGIVEN: %v\n", c.NOTGIVEN)
	log.Debugf("workToDeleteEnd: %v\n", workToDeleteEnd)

	// Ensure the workflow exists
	if !helper.WorkflowExisting(workToDeleteEnd) {
		helper.SpinStopDisplay("fail")
		log.Warningln("No matching workflow for '" + workToDeleteEnd + "'")
		log.Warningln("Please use:")
		log.Warningln("#> " + RootConfig.ScriptName + " list")
		os.Exit(1)
	}

	helper.SpinUpdateDisplay("Verifications")
	helper.SpinStopDisplay("success")
}

func endCommand(cmd *cobra.Command, args []string) {

	// Some display
	log.Debug("run end")
	helper.SpinStartDisplay("Git operations")

	// Get the ref branch to switch back
	refBranch, err := helper.RepoGetWorkflowParam(workToDeleteEnd, helper.REFBRANCHPARAM)
	if err != nil {
		log.Fatalln("Workflow '" + workToDeleteEnd + "' has no '" + helper.REFBRANCHPARAM + "' parameter")
	}
	log.Debugf("refBranch: %v\n", refBranch)

	// Get branch ref (plumbing)
	workToDeleteRef := helper.RepoGetBranchRef(workToDeleteEnd)

	// execute checkout and pull only if we are in the current worflow
	pullInfo := ""
	if currentEnd {
		helper.SpinUpdateDisplay("git checkout " + refBranch)
		helper.RepoCheckout(refBranch, RootRepo.PublicAuthKey)

		helper.SpinUpdateDisplay("git pull")
		pullInfo = helper.RepoPull(RootRepo.PublicAuthKey)
	}

	// Delete Branch
	helper.SpinUpdateDisplay("git branch -D " + workToDeleteEnd)
	helper.RepoDeleteBranch(workToDeleteEnd, workToDeleteRef)

	// Cleanup workflow config
	helper.RepoConfigDeleteWorkflow(workToDeleteEnd)
	helper.RepoConfigDeleteBranch(workToDeleteEnd)
	if currentEnd {
		// Delete current workflow
		helper.RepoConfigDeleteCurrentWorkflow()
	}

	helper.RepoConfigWrite()

	helper.SpinUpdateDisplay("Git operations")
	helper.SpinStopDisplay("success")
	helper.SpinSideNoteDisplay("Branch deleted > " + workToDeleteEnd)
	if pullInfo != "" {
		helper.SpinSideNoteDisplay("Pull info: " + pullInfo)
	}

	// Say GoodBye
	helper.ByeByeDisplay()
}

func init() {
	rootCmd.AddCommand(endCmd)

	RootConfig = helper.NewConfig()
	RootRepo := helper.NewRepo(RootConfig)

	var worklistStr string
	for _, w := range RootRepo.Worklist {
		worklistStr += "\t - " + w + "\n"
	}

	endCmd.Flags().StringVarP(&workEndArg, "work", "w", c.NOTGIVEN, "Work to end \n"+worklistStr)

}
