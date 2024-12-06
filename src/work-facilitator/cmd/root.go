/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"spirit-dev/work-facilitator/work-facilitator/common"

	"github.com/spf13/cobra"
)

// var (
// 	// _          = helper.WelcomeDisplay()
// 	RootConfig = helper.NewConfig()
// 	RootRepo   = helper.NewRepo(RootConfig)
// )

var (
	// _          = helper.WelcomeDisplay()
	// RootConfig = helper.NewConfig()
	// RootRepo   = helper.NewRepo(RootConfig)
	RootConfig common.Config
	RootRepo   common.Repo
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "work-facilitator",
	Short: "Easily manage you git repo based on JIRA / GitLab issues",
	Long: `
You can manage you git repository workloaf based on your company JIRA or GitLab issues.
The goal is to match the semantic release, branch naming strategy, commit message compliance without having to type things every time.
We wanna be lazy and error prone.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Execute Cmd
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}

	// // Say GoodBye
	// helper.ByeByeDisplay()
}

func init() {

	// rootCmd.CompletionOptions.DisableDefaultCmd = true
	// rootCmd.CompletionOptions.HiddenDefaultCmd = true

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.work-facilitator.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
