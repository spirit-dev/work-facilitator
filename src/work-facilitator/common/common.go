package common

import (
	"regexp"

	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

type Config struct {
	LogLevel        string
	AppName         string
	ScriptName      string
	Version         string
	EnforceStandard bool
	DefaultBranch   string

	BranchContent    []string
	BranchContentStr string
	BranchExpr       string
	BranchTemplate   string
	BranchSeparator  string
	CommitType       []string
	CommitTypeStr    string
	CommitExpr       string
	CommitTemplate   string
	TypeMapping      string

	Ticketing string

	TicketingJiraEnabled  bool
	TicketingJiraServer   string
	TicketingJiraUsername string
	TicketingJiraPassword string

	TicketingGlabEnabled bool
	TicketingGlabServer  string
	TicketingGlabToken   string

	HasSshKeyId bool
	SshKeyId    string

	CommitIgnorePatterns         []string
	CommitIgnorePatternsCompiled []*regexp.Regexp

	// UncommittedFilesDetection defines behavior when uncommitted files are detected during 'end' command
	// Options: "disabled", "warning", "fatal", "interactive"
	UncommittedFilesDetection string

	// AI configuration
	AIEnabled                 bool
	AIProvider                string
	AIAPIKey                  string
	AIModel                   string
	AIMaxTokens               int
	AITemperature             float64
	AITimeout                 int
	AIExcludePatterns         []string
	AIGoogleProjectID         string
	AIGoogleLocation          string
	AIGoogleServiceAccountKey string
}

type Workflow struct {
	CurrentWork string
	BranchType  string
	CommitType  string
	Issue       int
	Ticket      string
	Title       string
	Commit      string
	RefBranch   string
	Branch      string
}

type Repo struct {
	BasePath            string
	OriginUrl           string
	BrowserUrl          string
	Namespace           string
	Name                string
	FName               string
	DefaultBranch       string
	Worklist            []string
	CurrentWorkflowName string
	HasCurrentWorkflow  bool
	CurrentWorkflowData Workflow
	PublicAuthKey       *ssh.PublicKeys
	Separator           string
}

type JiraConfig struct {
	Server   string
	Username string
	Password string
}

type GlabConfig struct {
	BaseUrl string
	Token   string
}

const (
	JIRA           = "JIRA"
	GITLAB         = "GITLAB"
	NOTGIVEN       = "notGiven"
	NOTGIVENBRANCH = "notGivenBranch"
	GOMASTER       = "go_master"

	// Default commit ignore patterns (regex)
	DefaultCommitIgnorePattern1 = `out\.ya?ml$`
	DefaultCommitIgnorePattern2 = `out\d+\.ya?ml$`
)
