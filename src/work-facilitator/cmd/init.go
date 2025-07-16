/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	c "spirit-dev/work-facilitator/work-facilitator/common"
	"spirit-dev/work-facilitator/work-facilitator/helper"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	// Cmd args
	titleInitArg      string
	branchTypeInitArg string
	commitTypeInitArg string
	refBranchInitArg  string
	issueInitArg      int
	ticketInitArg     string

	// local variables
	currentWorkInit string
	commitInit      string

	initArgs = []string{
		"message\tCommit message",
	}
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:       "init issue title branch_type" + RootConfig.BranchContentStr + " [flags]",
	Short:     "initialize workflow",
	Long:      `Start a new workflow`,
	Args:      cobra.ExactArgs(3),
	ValidArgs: initArgs,
	PreRun:    initPreRunCommand,
	Run:       initCommand,
}

func initPreRunCommand(cmd *cobra.Command, args []string) {
	helper.WelcomeDisplay()
	RootConfig = helper.NewConfig()
	RootRepo = helper.NewRepo(RootConfig)

	helper.SpinStartDisplay("Verifications...")
	// Extract issue or ticket depending on GitLab or Jira
	// For Gitlab, we expect an int
	// For Jira, we expect a string
	if RootConfig.Ticketing == c.GITLAB {
		issueG, errG := strconv.Atoi(args[0])
		if errG != nil {
			log.Warningln("The issue should be an integer, corresponding to a GitLab MR number")
		}
		issueInitArg = issueG
	}
	if RootConfig.Ticketing == c.JIRA {
		ticketInitArg = args[0]
	}

	titleInitArg = args[1]
	branchTypeInitArg = args[2]

	// Set default ref branch
	if refBranchInitArg == c.NOTGIVENBRANCH {
		refBranchInitArg = RootRepo.DefaultBranch
	}

	// Clean title (remove any non word character)
	titleInitArg = helper.CleanString(titleInitArg)

	// Override commit type if not given
	if commitTypeInitArg == c.NOTGIVEN {
		commitTypeInitArg = helper.DefineCommit(branchTypeInitArg, RootConfig.TypeMapping)
	}
	// Prepare variables depending on ticketing service
	if RootConfig.Ticketing == c.GITLAB {
		currentWorkInit = titleInitArg
		commitInit = fmt.Sprintf("%s(!%d): ", commitTypeInitArg, issueInitArg)
	}
	if RootConfig.Ticketing == c.JIRA {
		// Define branch template
		currentWorkInit = helper.Template(RootConfig.BranchTemplate, map[string]interface{}{
			"type":    branchTypeInitArg,
			"issue":   ticketInitArg,
			"summary": titleInitArg,
		})
		// Define commit template
		commitInit = helper.Template(RootConfig.CommitTemplate, map[string]interface{}{
			"type":  commitTypeInitArg,
			"issue": ticketInitArg,
		})
	}

	// Ensure standard is correct (if enforced)
	if !helper.TestStandard(commitInit, RootConfig.CommitExpr, currentWorkInit, RootConfig.BranchExpr, RootConfig.EnforceStandard) {
		helper.SpinStopDisplay("fail")
		log.Fatalln("Standard not respected")
	}

	// Quit if a workflow already exists
	if helper.WorkflowExisting(currentWorkInit) {
		helper.SpinStopDisplay("fail")
		log.Warningln("This workflow already exists")
		log.Warningln("If you want to use this work flow, you can run:")
		log.Warningln("#> " + RootConfig.ScriptName + " use " + currentWorkInit)
		os.Exit(1)
	}

	helper.SpinUpdateDisplay("Verifications")
	helper.SpinStopDisplay("success")
}

func initCommand(cmd *cobra.Command, args []string) {

	helper.SpinStartDisplay("Git operations")
	workflow := c.Workflow{
		CurrentWork: currentWorkInit,
		BranchType:  branchTypeInitArg,
		CommitType:  commitTypeInitArg,
		Issue:       issueInitArg,
		Ticket:      ticketInitArg,
		Title:       titleInitArg,
		Commit:      commitInit,
		RefBranch:   refBranchInitArg,
		Branch:      currentWorkInit,
	}
	// Set the current worklow
	helper.RepoConfigDefineWorkflow(RootConfig, workflow)

	// execute git actions
	helper.SpinUpdateDisplay("git checkout")
	helper.RepoCheckout(refBranchInitArg, RootRepo.PublicAuthKey)
	helper.SpinUpdateDisplay("git pull")
	pullInfo := helper.RepoPull(RootRepo.PublicAuthKey)
	helper.SpinUpdateDisplay("git checkout")
	helper.RepoCheckout(currentWorkInit, RootRepo.PublicAuthKey)

	// Write workflow
	helper.RepoConfigWrite()
	helper.SpinUpdateDisplay("Git operations")
	helper.SpinStopDisplay("success")

	helper.SpinSideNoteDisplay("Pull info: " + pullInfo)

	helper.ShowSummary(RootConfig, workflow)

	// Say GoodBye
	helper.ByeByeDisplay()
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(&commitTypeInitArg, "commit-type", "c", c.NOTGIVEN, "Specify the commit type to be treated "+RootConfig.CommitTypeStr)
	initCmd.Flags().StringVarP(&refBranchInitArg, "ref-branch", "r", c.NOTGIVENBRANCH, "Specify the source branch")

	initCmd.MarkFlagRequired("title")
	initCmd.MarkFlagRequired("branch-type")

	initCmd.Flags().SortFlags = false
}
