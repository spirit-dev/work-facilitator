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
	forcePause     bool

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

	// Check for uncommitted files during validation phase
	if RootConfig.UncommittedFilesDetection != "disabled" && !forcePause {
		log.Debug("Checking for uncommitted files...")
		uncommittedFiles, hasUncommitted, err := helper.RepoCheckUncommittedFiles()
		if err != nil {
			helper.SpinStopDisplay("fail")
			log.Fatalln("Error checking repository status:", err)
		}

		if hasUncommitted {
			helper.SpinStopDisplay("warning")
			helper.DisplayUncommittedFiles(uncommittedFiles)

			switch RootConfig.UncommittedFilesDetection {
			case "fatal":
				log.Fatalln("Uncommitted files detected. Please commit or stash changes before pausing workflow.")
			case "warning":
				helper.SpinSideNoteDisplay("Warning: Uncommitted files detected")
				helper.SpinStartDisplay("Verifications - pause...")
			case "interactive":
				// Prompt user for confirmation
				if !helper.PromptUserConfirmation("Continue with uncommitted files?") {
					log.Fatalln("Operation cancelled by user")
				}
				helper.SpinStartDisplay("Verifications - pause...")
			}
		}
	} else if forcePause && RootConfig.UncommittedFilesDetection != "disabled" {
		log.Debug("Force flag set, skipping uncommitted files check...")
		helper.SpinSideNoteDisplay("Warning: Skipping uncommitted files check (force mode)")
	}

	helper.SpinUpdateDisplay("Verifications")
	helper.SpinStopDisplay("success")
}

func pauseCommand(cmd *cobra.Command, args []string) {
	log.Debug("run pause")
	helper.SpinStartDisplay("Git operations")

	// Re-check for uncommitted files in case files changed between PreRun and Run
	if RootConfig.UncommittedFilesDetection == "fatal" && !forcePause {
		log.Debug("Re-checking for uncommitted files before git operations...")
		uncommittedFiles, hasUncommitted, err := helper.RepoCheckUncommittedFiles()
		if err != nil {
			helper.SpinStopDisplay("fail")
			log.Fatalln("Error checking repository status:", err)
		}

		if hasUncommitted {
			helper.SpinStopDisplay("fail")
			helper.DisplayUncommittedFiles(uncommittedFiles)
			log.Fatalln("Uncommitted files detected. Aborting workflow pause.")
		}
	}

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
	pauseCmd.Flags().BoolVarP(&forcePause, "force", "f", false, "Force pause workflow, skip uncommitted files check")

}
