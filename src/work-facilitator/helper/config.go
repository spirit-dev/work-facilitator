package helper

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	c "spirit-dev/work-facilitator/work-facilitator/common"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

func NewConfig() c.Config {

	initConfig()

	logLevel := viper.GetString("global.log_level")
	initLogging(logLevel)

	branchContent := viper.GetString("global.branch_content")
	branchContentSplit := strings.Split(branchContent, "|")
	branchContentJoin := strings.Join(branchContentSplit, ",")

	commitType := viper.GetString("global.commit_content")
	commitTypeSplit := strings.Split(commitType, "|")
	commitTypeSplit = append(commitTypeSplit, "notGiven")
	commitTypeJoin := strings.Join(commitTypeSplit, ",")
	typeMapping := viper.GetString("global.type_mapping")

	ticketingJiraEnabled := viper.GetBool("ticketing.jira.enabled")
	ticketingJiraServer := viper.GetString("ticketing.jira.server")
	ticketingJiraUsername := viper.GetString("ticketing.jira.username")
	ticketingJiraPassword := viper.GetString("ticketing.jira.password")
	ticketingGlabEnabled := viper.GetBool("ticketing.gitlab.enabled")
	ticketingGlabServer := viper.GetString("ticketing.gitlab.server")
	ticketingGlabToken := viper.GetString("ticketing.gitlab.token")

	ticketing := ""
	if ticketingJiraEnabled && ticketingGlabEnabled {
		log.Fatalln("Both Jira and Gitlab defined. Unknow behaviour")
	}
	if ticketingJiraEnabled {
		ticketing = c.JIRA
	} else if ticketingGlabEnabled {
		ticketing = c.GITLAB
	} else {
		log.Warningln("No ticketing system defined. Unknow behaviour")
	}

	sshKeyId := viper.GetString("global.ssh_key_id")
	hasSshKeyId := false
	if sshKeyId != "" {
		hasSshKeyId = true
	}

	conf := c.Config{
		LogLevel:              logLevel,
		AppName:               viper.GetString("global.app_name"),
		ScriptName:            viper.GetString("global.script_name"),
		Version:               viper.GetString("global.version"),
		EnforceStandard:       viper.GetBool("global.enforce_standard"),
		DefaultBranch:         viper.GetString("global.default_branch"),
		BranchContent:         branchContentSplit,
		BranchContentStr:      "[" + branchContentJoin + "]",
		BranchExpr:            viper.GetString("global.branch_expr"),
		CommitType:            commitTypeSplit,
		CommitTypeStr:         "[" + commitTypeJoin + "]",
		CommitExpr:            viper.GetString("global.commit_expr"),
		TypeMapping:           typeMapping,
		Ticketing:             ticketing,
		TicketingJiraEnabled:  ticketingJiraEnabled,
		TicketingJiraServer:   ticketingJiraServer,
		TicketingJiraUsername: ticketingJiraUsername,
		TicketingJiraPassword: ticketingJiraPassword,
		TicketingGlabEnabled:  ticketingGlabEnabled,
		TicketingGlabServer:   ticketingGlabServer,
		TicketingGlabToken:    ticketingGlabToken,
		HasSshKeyId:           hasSshKeyId,
		SshKeyId:              sshKeyId,
	}

	return conf
}

func QuickConfig() (string, string) {

	initConfig()
	appName := viper.GetString("global.app_name")
	scriptName := viper.GetString("global.script_name")
	return scriptName, appName
}

func initLogging(logLvl string) {

	// Setup logging
	if logLvl == log.TraceLevel.String() {
		log.SetLevel(log.TraceLevel)
		log.Traceln("log level: trace")
	} else if logLvl == log.DebugLevel.String() {
		log.SetLevel(log.DebugLevel)
		log.Debugln("log level: debug")
	} else if logLvl == log.InfoLevel.String() {
		log.SetLevel(log.InfoLevel)
	} else if logLvl == log.WarnLevel.String() {
		log.SetLevel(log.WarnLevel)
	} else if logLvl == log.ErrorLevel.String() {
		log.SetLevel(log.ErrorLevel)
	} else if logLvl == log.FatalLevel.String() {
		log.SetLevel(log.FatalLevel)
	} else if logLvl == log.PanicLevel.String() {
		log.SetLevel(log.PanicLevel)
	}
}

func initConfig() {
	viper.SetConfigName(".workflow") // name of config file (without extension)
	viper.SetConfigType("ini")       // REQUIRED if the config file does not have the extension in the name
	// Search config in home directory with name ".cobra" (without extension).
	viper.AddConfigPath(homeDir())                                                            // TODO re-enable the home dir as a first place to look at
	viper.AddConfigPath("/home/jbordat/Documents/Projects/scripts/work-facilitator/config")   // path to look for the config file in
	viper.AddConfigPath("/home/jbordat/Documents/projects/scripts/work-facilitator/config")   // path to look for the config file in
	viper.AddConfigPath("/Users/bordaje1/Documents/Projects/scripts/work-facilitator/config") // path to look for the config file in
	viper.AddConfigPath("/code/config")                                                       // call multiple times to add many search paths

	err1 := viper.ReadInConfig() // Find and read the config file
	if err1 != nil {             // Handle errors reading the config file
		log.Fatalln("fatal error config file: %w", err1)
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("wf") // will be uppercased automatically
}

func homeDir() string {

	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		SpinStopDisplay("fail")
		fmt.Println(err)
		os.Exit(1)
	}

	return home
}

func BuildBranchTypeConfig() []string {
	return NewConfig().BranchContent
}

func CurrentPath() string {
	path, _ := os.Getwd()
	// path = path + "/../../repo_test" // TODO comment this out
	// path = "/Users/bordaje1/Documents/Projects/devops-shared-platform/armature/mobile/mobile-armature-definition"
	log.Debugln("Current path: " + path)

	return path
}

func CleanGlabString(txt string) string {
	t1 := strings.Replace(txt, "Draft: Resolve ", "", -1)
	t2 := strings.Replace(t1, `"`, "", -1)
	return t2
}

func CleanString(text string) string {

	// Cleanup any non word character
	// Typically will transform
	// from		test8&*weqwe()
	// to		test8__weqwe__
	re := regexp.MustCompile(`\W`)
	cleaned := strings.Join(re.Split(text, -1), "_")
	log.Debugln("Cleaned title from: " + text + " -- to : " + cleaned)

	return cleaned
}

func DefineCommit(branchType string, typeMapping string) string {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(typeMapping), &jsonMap)

	log.Debugln("Commit type mapping: " + typeMapping)
	commit := jsonMap[branchType]

	if commit == nil {
		SpinStopDisplay("fail")
		log.Fatalln("No commit type found in mapping for: " + branchType)
	} else {
		log.Debugln("Commit type found in mapping : " + branchType + " -> " + commit.(string))
	}

	return commit.(string)
}

func TestStandard(commit, commitExpr, branch, branchExpr string, standardEnforced bool) bool {
	ok := true
	if standardEnforced {

		// Test commit pattern
		testCommit := fmt.Sprintf("%stest_message", commit)
		re := regexp.MustCompile(commitExpr)
		if !re.MatchString(testCommit) {
			log.Warningln(`Sorry, the commit message does not comply the standard
Please review:
 regex : ` + commitExpr + `
 commit: ` + commit)
			ok = false
		}

		// Test branch pattern
		re = regexp.MustCompile(branchExpr)
		if !re.MatchString(branch) {
			log.Warningln(`Sorry, the branch pattern does not comply the standard
Please review:
 regex : ` + branchExpr + `
 commit: ` + branch)
			ok = false
		}
	}
	return ok
}
