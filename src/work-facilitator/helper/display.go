package helper

import (
	c "spirit-dev/work-facilitator/work-facilitator/common"
	"strconv"

	"github.com/go-git/go-git/v5"
	"github.com/pterm/pterm"
	log "github.com/sirupsen/logrus"
)

var (
	thisSpinner *pterm.SpinnerPrinter
)

func WelcomeDisplay() string {

	scriptName, appName := QuickConfig()

	newHeader := pterm.HeaderPrinter{
		TextStyle:       pterm.NewStyle(pterm.FgBlack),
		BackgroundStyle: pterm.NewStyle(pterm.BgBlue),
		Margin:          15,
	}

	newHeader.Println(scriptName + " / " + appName)

	return ""
}

func ByeByeDisplay() {
	Addline("Thanks for using\n")
}

func VersionDisplay(version string) {
	Addline("Version : " + pterm.LightBlue(version))
}

func Addline(text string) {
	pterm.DefaultBasicText.Print(text)
}

func SpinSideNoteDisplay(text string) {
	Addline(pterm.Yellow(" > ") + pterm.Cyan(text) + "\n")
}

func SpinStartDisplay(text string) {
	if log.GetLevel().String() == log.InfoLevel.String() {
		thisSpinner, _ = pterm.DefaultSpinner.Start(text)
	}
}

func SpinUpdateDisplay(text string) {
	if log.GetLevel().String() == log.InfoLevel.String() {
		thisSpinner.UpdateText(text)
	}
}

func SpinStopDisplay(resolved string) {
	if log.GetLevel().String() == log.InfoLevel.String() {
		if resolved == "info" {
			thisSpinner.Info()
		} else if resolved == "success" {
			thisSpinner.Success()
		} else if resolved == "warning" {
			thisSpinner.Warning()
		} else if resolved == "fail" {
			thisSpinner.Fail()
		}
	}
}

func ShowSummary(cfg c.Config, wf c.Workflow) {
	dt, title := "", ""

	if cfg.Ticketing == c.JIRA {
		title = "branch type\ncommit type\ntitle\nticket\nbranch\ncommit\nref branch"
		dt = wf.BranchType + "\n" + wf.CommitType + "\n" + wf.Title + "\n" + wf.Ticket + "\n" + wf.Branch + "\n" + wf.Commit + "\n" + wf.RefBranch
	}
	if cfg.Ticketing == c.GITLAB {
		title = "branch type\ncommit type\ntitle\nissue\nbranch\ncommit\nref branch"
		dt = wf.BranchType + "\n" + wf.CommitType + "\n" + wf.Title + "\n" + strconv.Itoa(wf.Issue) + "\n" + wf.Branch + "\n" + wf.Commit + "\n" + wf.RefBranch
	}

	panels := pterm.Panels{
		{
			{Data: pterm.DefaultBox.WithRightPadding(10).WithLeftPadding(10).Sprintf(wf.Branch)},
		},
		{
			{Data: pterm.Cyan(title)},
			{Data: dt},
		},
	}

	pterm.DefaultPanel.WithPanels(panels).WithPadding(15).Render()
}

func ShowBox(status git.Status) {
	pterm.DefaultBox.WithRightPadding(5).WithLeftPadding(5).Println("changes")
	Addline(status.String() + "\n")
}

func UpgradeV2ToV3() bool {
	// Show an interactive confirmation dialog and get the result.
	result, _ := pterm.DefaultInteractiveConfirm.Show()

	// Print a blank line for better readability.
	pterm.Println()

	return result
}
