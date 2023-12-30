package web

import (
	"net/http"

	"github.com/ammesonb/ubiquiti-config-generator/config"
	"github.com/ammesonb/ubiquiti-config-generator/logger"
)

type checkSuite struct {
	AfterSHA   string `json:"after"`
	BeforeSHA  string `json:"before"`
	Conclusion string `json:"conclusion"`
	HeadBranch string `json:"head_branch"`
	HeadCommit string `json:"head_commit"`
	HeadSHA    string `json:"head_sha"`
	Status     string `json:"status"`
}

type repository struct {
	Name           string `json:"name"`
	URL            string `json:"URL"`
	CloneURL       string `json:"clone_url"`
	StatusesURL    string `json:"statuses_url"`
	DeploymentsURL string `json:"deployments_url"`

	App struct {
		ExternalURL string `json:"external_url"`
	} `json:"app"`
}

type checkSuiteRequest struct {
	CheckSuite checkSuite `json:"check_suite"`
	Repository repository `json:"repository"`
}

// ProcessGitCheckSuite will create new check runs and update their statuses as appropriate
func ProcessGitCheckSuite(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	log := logger.DefaultLogger()
	var body []byte
	if _, err := r.Body.Read(body); err != nil {
		internalServerError(w, log, "Failed to read check suite request body", err)
		return
	}

	if !validateGitBody(w, r, cfg.Git.WebhookSecret, log, "check suite", body) {
		return
	}

	form := r.Form
	if !form.Has("action") {
		badRequest(w, log, "Check suite did not specify action")
		return
	}

	action := form.Get("action")

	if action == "completed" {
		log.Info("Maybe should check deployment here - unsure why this is needed")
	} else if action != "requested" && action != "rerequested" {
		log.Infof("Ignoring check suite action %s", action)
		return
	}
}
