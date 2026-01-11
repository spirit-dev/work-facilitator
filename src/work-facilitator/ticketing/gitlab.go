package ticketing

import (
	c "spirit-dev/work-facilitator/work-facilitator/common"
	"spirit-dev/work-facilitator/work-facilitator/helper"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

var (
	glabClient *gitlab.Client
)

func ClientGlab(cfg c.GlabConfig) {
	gl, err := gitlab.NewClient(cfg.Token, gitlab.WithBaseURL(cfg.BaseUrl))
	if err != nil {
		helper.SpinStopDisplay("fail")
		log.Fatalf("Failed to create client: %v", err)
	}
	glabClient = gl
}

func GetGlabIssue(key int, pid string) *gitlab.MergeRequest {

	log.Debugf("key: %v\n", key)

	pjt, _, err := glabClient.Projects.GetProject(pid, &gitlab.GetProjectOptions{})
	if err != nil {
		helper.SpinStopDisplay("fail")
		log.Fatalln("Gitlab project not found for pid " + pid + " - " + err.Error())
	}

	mr, _, err := glabClient.MergeRequests.GetMergeRequest(pjt.ID, key, &gitlab.GetMergeRequestsOptions{})
	if err != nil {
		helper.SpinStopDisplay("fail")
		log.Fatalln("Gitlab mr not found for key " + strconv.Itoa(key) + " - " + err.Error())
	}

	return mr
}
