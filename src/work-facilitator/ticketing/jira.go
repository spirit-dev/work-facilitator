package ticketing

import (
	c "spirit-dev/work-facilitator/work-facilitator/common"
	"spirit-dev/work-facilitator/work-facilitator/helper"

	jira "github.com/andygrunwald/go-jira"
	log "github.com/sirupsen/logrus"
)

var (
	jiraClient *jira.Client
)

func ClientJira(cfg c.JiraConfig) {

	bt := jira.BasicAuthTransport{
		Username: cfg.Username,
		Password: cfg.Password,
	}
	jiraClient, _ = jira.NewClient(bt.Client(), cfg.Server)
}

func GetJiraIssue(key string) *jira.Issue {

	log.Debugf("key: %v\n", key)

	issue, _, err := jiraClient.Issue.Get(key, nil)
	if err != nil {
		helper.SpinStopDisplay("fail")
		log.Fatalln("Jira issue for key : " + key + " - " + err.Error())
	}

	return issue
}
