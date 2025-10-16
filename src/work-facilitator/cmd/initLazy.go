/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	c "spirit-dev/work-facilitator/work-facilitator/common"
	"spirit-dev/work-facilitator/work-facilitator/helper"
	"spirit-dev/work-facilitator/work-facilitator/ticketing"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// Cmd args
	issueInitLArg          string
	issueInitLArgI         int
	branchTypeInitLArg     string
	commitTypeInitLArg     string
	refBranchInitLArg      string
	titleSeparatorInitLArg string

	// local variables
	currentWorkInitL string
	commitInitL      string
	summaryInitL     string

	initLazyArgs = []string{
		"issue\tIssue from GitLab or Jira",
		"branch_type\tBranch type ",
	}
)

// initLazyCmd represents the initLazy command
var initLazyCmd = &cobra.Command{
	Use:       "initLazy issue branch_type" + RootConfig.BranchContentStr + " [flags]",
	Short:     "initialize workflow lazy style",
	Long:      "Start a new workflow lazy style",
	Args:      cobra.ExactArgs(2),
	ValidArgs: initLazyArgs,
	PreRun:    initLazyPreRunCommand,
	Run:       initLazyCommand,
}

func initLazyPreRunCommand(cmd *cobra.Command, args []string) {
	helper.WelcomeDisplay()
	RootConfig = helper.NewConfig()
	RootRepo = helper.NewRepo(RootConfig)

	log.Debug("pre run init-lazy")
	helper.SpinStartDisplay("Verifications - init-lazy...")

	if RootRepo.HasCurrentWorkflow {
		helper.SpinStopDisplay("fail")
		log.Fatalln("You are in a workflow at this moment. Weird behaviour might occur")
	}

	issueInitLArg = args[0]                         // Sets for either Glab or Jira
	issueInitLArgI, _ = strconv.Atoi(issueInitLArg) // Set for gitlab with an attempt to str to int. Fails for Jira
	log.Debugf("issueInitLArg: %v\n", issueInitLArg)
	log.Debugf("issueInitLArgI: %v\n", issueInitLArgI)

	branchTypeInitLArg = args[1]

	// Override commit type if not given
	if commitTypeInitLArg == c.NOTGIVEN {
		commitTypeInitLArg = helper.DefineCommit(branchTypeInitLArg, RootConfig.TypeMapping)
	}

	// Define separator
	// Precedence:
	// 		1. cli
	//		2. repo
	// 		3. config
	if titleSeparatorInitLArg == c.NOTGIVEN && RootRepo.Separator == c.NOTGIVEN {
		titleSeparatorInitLArg = RootConfig.BranchSeparator
	}
	if titleSeparatorInitLArg == c.NOTGIVEN && RootRepo.Separator != c.NOTGIVEN {
		titleSeparatorInitLArg = RootRepo.Separator
	}

	if RootConfig.Ticketing == c.JIRA {
		// Prepare Jira
		ticketing.ClientJira(c.JiraConfig{
			Server:   RootConfig.TicketingJiraServer,
			Username: RootConfig.TicketingJiraUsername,
			Password: RootConfig.TicketingJiraPassword,
		})
		// get Jira issue
		issue := ticketing.GetJiraIssue(issueInitLArg)
		// Build summary
		summaryInitL = helper.CleanString(issue.Fields.Summary, titleSeparatorInitLArg)
		log.Debugf("summaryInitL: %v\n", summaryInitL)

		// Define branch template
		currentWorkInitL = helper.Template(RootConfig.BranchTemplate, map[string]interface{}{
			"type":    branchTypeInitLArg,
			"issue":   issueInitLArg,
			"summary": summaryInitL,
		})
		// Define commit template
		commitInitL = helper.Template(RootConfig.CommitTemplate, map[string]interface{}{
			"type":  commitTypeInitLArg,
			"issue": issueInitLArg,
		})
	}

	if RootConfig.Ticketing == c.GITLAB {

		// Prepare gitlab
		ticketing.ClientGlab(c.GlabConfig{
			BaseUrl: RootConfig.TicketingGlabServer,
			Token:   RootConfig.TicketingGlabToken,
		})
		// Get Gitlab issue
		issue := ticketing.GetGlabIssue(issueInitLArgI, RootRepo.FName)
		summaryInitL = helper.CleanString(helper.CleanGlabString(issue.Title), titleSeparatorInitLArg)
		log.Debugf("issue.Title: `%v` --> `%v`\n", issue.Title, summaryInitL)

		currentWorkInitL = issue.SourceBranch

		// Define commit pre message
		commitInitL = fmt.Sprintf("%s(!%s): ", commitTypeInitLArg, issueInitLArg)
	}

	// Set default ref branch
	if refBranchInitLArg == c.NOTGIVENBRANCH {
		refBranchInitLArg = RootConfig.DefaultBranch
	}

	// Ensure standard is correct (if enforced)
	if !helper.TestStandard(commitInitL, RootConfig.CommitExpr, currentWorkInitL, RootConfig.BranchExpr, RootConfig.EnforceStandard) {
		helper.SpinStopDisplay("fail")
		log.Fatalln("Standard not respected")
	}

	// Quit if a workflow already exists
	if helper.WorkflowExisting(currentWorkInitL) {
		helper.SpinStopDisplay("fail")
		log.Warningln("This workflow already exists")
		log.Warningln("If you want to use this work flow, you can run:")
		log.Warningln("#> " + RootConfig.ScriptName + " use " + currentWorkInitL)
		os.Exit(1)
	}

	helper.SpinUpdateDisplay("Verifications")
	helper.SpinStopDisplay("success")

	if RootConfig.Ticketing == c.JIRA {
		helper.SpinSideNoteDisplay("Got jira issue")
	}
	if RootConfig.Ticketing == c.GITLAB {
		helper.SpinSideNoteDisplay("Got GitLab mr")
	}
}

func initLazyCommand(cmd *cobra.Command, args []string) {
	log.Debug("run init-lazy")

	helper.SpinStartDisplay("Git operations")
	workflow := c.Workflow{
		CurrentWork: currentWorkInitL,
		BranchType:  branchTypeInitLArg,
		CommitType:  commitTypeInitLArg,
		Issue:       issueInitLArgI,
		Ticket:      issueInitLArg,
		Title:       summaryInitL,
		Commit:      commitInitL,
		RefBranch:   refBranchInitLArg,
		Branch:      currentWorkInitL,
	}
	// Set the current worklow
	helper.RepoConfigDefineWorkflow(RootConfig, workflow)

	// execute git actions
	helper.SpinUpdateDisplay("git checkout")
	helper.RepoCheckout(refBranchInitLArg, RootRepo.PublicAuthKey)
	helper.SpinUpdateDisplay("git pull")
	pullInfo := helper.RepoPull(RootRepo.PublicAuthKey)
	helper.SpinUpdateDisplay("git checkout")
	helper.RepoCheckout(currentWorkInitL, RootRepo.PublicAuthKey)

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
	rootCmd.AddCommand(initLazyCmd)

	initLazyCmd.Flags().StringVarP(&commitTypeInitLArg, "commit-type", "c", c.NOTGIVEN, "Specify the commit type to be treated "+RootConfig.CommitTypeStr)
	initLazyCmd.Flags().StringVarP(&refBranchInitLArg, "ref-branch", "r", c.NOTGIVENBRANCH, "Specify the source branch")
	initLazyCmd.Flags().StringVarP(&titleSeparatorInitLArg, "separator", "s", c.NOTGIVEN, "Specify the separator in the branch title")

	initLazyCmd.MarkFlagRequired("branch-type")

	initLazyCmd.Flags().SortFlags = false

}
