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

	"github.com/valyala/fasttemplate"
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

	// Load commit ignore patterns (with defaults)
	commitIgnorePatterns := viper.GetStringSlice("global.commit_ignore_patterns")
	if len(commitIgnorePatterns) == 0 {
		commitIgnorePatterns = []string{
			c.DefaultCommitIgnorePattern1,
			c.DefaultCommitIgnorePattern2,
		}
	}
	// Compile and validate all patterns
	var commitIgnorePatternsCompiled []*regexp.Regexp
	for _, pattern := range commitIgnorePatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			log.Fatalln("Invalid regex pattern in commit_ignore_patterns: " + pattern)
		}
		commitIgnorePatternsCompiled = append(commitIgnorePatternsCompiled, re)
	}

	// Load uncommitted files detection behavior (with default)
	uncommittedFilesDetection := viper.GetString("global.uncommitted_files_detection")
	if uncommittedFilesDetection == "" {
		uncommittedFilesDetection = "fatal" // Default to fatal mode
	}
	// Validate the value
	validDetectionModes := map[string]bool{
		"disabled":    true,
		"warning":     true,
		"fatal":       true,
		"interactive": true,
	}
	if !validDetectionModes[uncommittedFilesDetection] {
		log.Warningln("Invalid uncommitted_files_detection value: " + uncommittedFilesDetection + ". Using 'fatal' as default.")
		uncommittedFilesDetection = "fatal"
	}

	// Load AI configuration
	aiEnabled := viper.GetBool("ai.enabled")
	aiProvider := viper.GetString("ai.provider")
	if aiProvider == "" {
		aiProvider = "openai" // Default provider
	}
	aiAPIKey := viper.GetString("ai.api_key")
	// Support environment variable references (e.g., $OPENAI_API_KEY)
	if strings.HasPrefix(aiAPIKey, "$") {
		envVar := strings.TrimPrefix(aiAPIKey, "$")
		aiAPIKey = os.Getenv(envVar)
	}
	aiModel := viper.GetString("ai.model")
	aiMaxTokens := viper.GetInt("ai.max_tokens")
	if aiMaxTokens == 0 {
		aiMaxTokens = 1024 // Default max tokens
	}
	aiTemperature := viper.GetFloat64("ai.temperature")
	if aiTemperature == 0 {
		aiTemperature = 0.7 // Default temperature
	}
	aiTimeout := viper.GetInt("ai.timeout")
	if aiTimeout == 0 {
		aiTimeout = 30 // Default 30 seconds
	}
	aiExcludePatterns := viper.GetStringSlice("ai.exclude_patterns")

	// Vertex AI specific configuration
	aiGoogleProjectID := viper.GetString("ai.google_project_id")
	aiGoogleLocation := viper.GetString("ai.google_location")
	if aiGoogleLocation == "" {
		aiGoogleLocation = "us-central1" // Default location
	}
	aiGoogleServiceAccountKey := viper.GetString("ai.google_service_account_key")
	// Support environment variable references
	if strings.HasPrefix(aiGoogleServiceAccountKey, "$") {
		envVar := strings.TrimPrefix(aiGoogleServiceAccountKey, "$")
		aiGoogleServiceAccountKey = os.Getenv(envVar)
	}

	conf := c.Config{
		LogLevel:                     logLevel,
		AppName:                      viper.GetString("global.app_name"),
		ScriptName:                   viper.GetString("global.script_name"),
		Version:                      viper.GetString("global.version"),
		EnforceStandard:              viper.GetBool("global.enforce_standard"),
		DefaultBranch:                viper.GetString("global.default_branch"),
		BranchContent:                branchContentSplit,
		BranchContentStr:             "[" + branchContentJoin + "]",
		BranchExpr:                   viper.GetString("global.branch_expr"),
		BranchTemplate:               viper.GetString("global.branch_template"),
		BranchSeparator:              viper.GetString("global.branch_separator"),
		CommitType:                   commitTypeSplit,
		CommitTypeStr:                "[" + commitTypeJoin + "]",
		CommitExpr:                   viper.GetString("global.commit_expr"),
		CommitTemplate:               viper.GetString("global.commit_template"),
		TypeMapping:                  typeMapping,
		Ticketing:                    ticketing,
		TicketingJiraEnabled:         ticketingJiraEnabled,
		TicketingJiraServer:          ticketingJiraServer,
		TicketingJiraUsername:        ticketingJiraUsername,
		TicketingJiraPassword:        ticketingJiraPassword,
		TicketingGlabEnabled:         ticketingGlabEnabled,
		TicketingGlabServer:          ticketingGlabServer,
		TicketingGlabToken:           ticketingGlabToken,
		HasSshKeyId:                  hasSshKeyId,
		SshKeyId:                     sshKeyId,
		CommitIgnorePatterns:         commitIgnorePatterns,
		CommitIgnorePatternsCompiled: commitIgnorePatternsCompiled,
		UncommittedFilesDetection:    uncommittedFilesDetection,
		AIEnabled:                    aiEnabled,
		AIProvider:                   aiProvider,
		AIAPIKey:                     aiAPIKey,
		AIModel:                      aiModel,
		AIMaxTokens:                  aiMaxTokens,
		AITemperature:                aiTemperature,
		AITimeout:                    aiTimeout,
		AIExcludePatterns:            aiExcludePatterns,
		AIGoogleProjectID:            aiGoogleProjectID,
		AIGoogleLocation:             aiGoogleLocation,
		AIGoogleServiceAccountKey:    aiGoogleServiceAccountKey,
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
	viper.SetConfigType("yaml")      // REQUIRED if the config file does not have the extension in the name
	// Search config in home directory with name ".cobra" (without extension).
	// viper.AddConfigPath(homeDir()) // TODO re-enable the home dir as a first place to look at
	// viper.AddConfigPath("/home/jbordat/Documents/Projects/scripts/work-facilitator/config") // path to look for the config file in
	viper.AddConfigPath("/home/jbordat/Documents/projects/scripts/work-facilitator/config") // path to look for the config file in
	// viper.AddConfigPath("/Users/bordaje1/Documents/Projects/scripts/work-facilitator/config") // path to look for the config file in
	// viper.AddConfigPath("/home/jbordat/Documents/tmp/work-facilitator/config")
	// viper.AddConfigPath("/code/config") // call multiple times to add many search paths

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
	// path = path + "/home/jbordat/Documents/projects/perso/test" // TODO comment this out
	log.Debugln("Current path: " + path)

	return path
}

func CleanGlabString(txt string) string {
	t1 := strings.Replace(txt, "Draft: Resolve ", "", -1)
	t2 := strings.Replace(t1, `"`, "", -1)
	return t2
}

func CleanString(input, separator string) string {
	var result strings.Builder
	lastWasSeparator := false

	for _, char := range input {
		if (char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') {
			result.WriteRune(char)
			lastWasSeparator = false
		} else if !lastWasSeparator && result.Len() > 0 {
			// Add separator only if previous char wasn't a separator
			// and we're not at the start
			result.WriteString(separator)
			lastWasSeparator = true
		}
	}

	// Remove trailing separator if present
	resultStr := result.String()
	if len(resultStr) > 0 && len(separator) > 0 && strings.HasSuffix(resultStr, separator) {
		resultStr = resultStr[:len(resultStr)-len(separator)]
	}

	return resultStr
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

func Template(template string, m map[string]interface{}) string {

	t := fasttemplate.New(template, "{{", "}}")
	s := t.ExecuteString(m)

	return s
}
