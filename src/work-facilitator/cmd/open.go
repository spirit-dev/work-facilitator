/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"spirit-dev/work-facilitator/work-facilitator/helper"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

// openCmd represents the open command
var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Open the corresponding repo in brwoser",
	Long:  "Open your browser for the corresponding repo",
	Run:   openCommand,
}

func openCommand(cmd *cobra.Command, args []string) {
	helper.WelcomeDisplay()
	RootConfig = helper.NewConfig()
	RootRepo = helper.NewRepo(RootConfig)

	log.Debug("run end")

	url := RootRepo.BrowserUrl
	// browser.OpenURL(url)
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}

	// Say GoodBye
	helper.ByeByeDisplay()
}

func init() {
	rootCmd.AddCommand(openCmd)
}
