/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"spirit-dev/work-facilitator/work-facilitator/helper"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	// Cmd args
	noPushCommitArg   bool
	allFilesCommitArg bool
	messageCommitArg  string
	forceCommitArg    bool

	// local variables
	commitMessageCommit string

	commitArgs = []string{
		"message\tCommit message",
	}
)

// commitCmd represents the commit command
var commitCmd = &cobra.Command{
	Use:       "commit message [flags]",
	Short:     "Commit workflow",
	Long:      "Commit work in current branch",
	Args:      cobra.ExactArgs(1),
	ValidArgs: commitArgs,
	PreRun:    commitPreRunCommand,
	Run:       commitCommand,
}

func commitPreRunCommand(cmd *cobra.Command, args []string) {
	helper.WelcomeDisplay()
	RootConfig = helper.NewConfig()
	RootRepo = helper.NewRepo(RootConfig)

	log.Debug("pre run commit")
	helper.SpinStartDisplay("Verifications - commit...")

	// Get commit message
	messageCommitArg = args[0]

	if !forceCommitArg && !RootRepo.HasCurrentWorkflow {
		helper.SpinStopDisplay("fail")
		log.Warningln("We are not in a current workflow.")
		log.Warningln("You can force the commit/push by running the following command")
		log.Warningln("#> " + RootConfig.ScriptName + " commit [-a, -n, -f] message")
		os.Exit(1)
	}

	// Define pre commit message
	preMessageCommit := RootRepo.CurrentWorkflowData.Commit
	log.Debugf("preMessageCommit: %v\n", preMessageCommit)
	activeBranch := RootRepo.CurrentWorkflowData.Branch
	log.Debugf("activeBranch: %v\n", activeBranch)

	// Ensure standard is correct (if enforced)
	commitMessageCommit = fmt.Sprintf("%s%s", preMessageCommit, messageCommitArg)
	if !helper.TestStandard(commitMessageCommit, RootConfig.CommitExpr, activeBranch, RootConfig.BranchExpr, RootConfig.EnforceStandard) {
		helper.SpinStopDisplay("fail")
		log.Fatalln("Standard not respected")
	}
	log.Debugf("commitMessageCommit: %v\n", commitMessageCommit)

	helper.SpinUpdateDisplay("Verifications")
	helper.SpinStopDisplay("success")
}

func commitCommand(cmd *cobra.Command, args []string) {
	log.Debug("run commit")
	helper.SpinStartDisplay("Git operations")

	// git add all files
	if allFilesCommitArg {
		helper.SpinUpdateDisplay("Git add all files")
		helper.RepoAddAllFiles(RootConfig.CommitIgnorePatternsCompiled)
	}

	// git commit
	helper.SpinUpdateDisplay("Git commit")
	helper.RepoCommit(commitMessageCommit, RootConfig.CommitIgnorePatternsCompiled)

	// git push
	if !noPushCommitArg {
		helper.SpinUpdateDisplay("Git push")
		helper.RepoPush(RootRepo.PublicAuthKey, RootRepo.CurrentWorkflowData.Branch)
	}

	helper.SpinUpdateDisplay("Git operations")
	helper.SpinStopDisplay("success")

	if !noPushCommitArg {
		helper.SpinSideNoteDisplay("git push origin")
	}

	// Say GoodBye
	helper.ByeByeDisplay()
}

func init() {
	rootCmd.AddCommand(commitCmd)

	commitCmd.Flags().BoolVarP(&noPushCommitArg, "no-push", "n", false, "Activate option to avoid pushing commits")
	commitCmd.Flags().BoolVarP(&allFilesCommitArg, "all-files", "a", false, "specify the merge request number")
	commitCmd.Flags().BoolVarP(&forceCommitArg, "force-commit", "f", false, "force the commit if we are no in a workflow")

	commitCmd.Flags().SortFlags = false
}
