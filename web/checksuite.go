package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ammesonb/ubiquiti-config-generator/services/configuration"
	"github.com/charmbracelet/log"

	"github.com/ammesonb/ubiquiti-config-generator/services/db"
)

// ProcessGitCheckSuite will create new check runs and update their statuses as appropriate
func ProcessGitCheckSuite(
	w http.ResponseWriter,
	r *http.Request,
	client *http.Client,
	configService configuration.Service,
	dbService db.Service,
	logger *log.Logger,
	accessToken string,
) {
	gitCfg := configService.GetGitConfig()
	var body []byte
	if _, err := r.Body.Read(body); err != nil {
		internalServerError(w, logger, "Failed to read check suite request body", err)
		return
	}

	if !validateGitWebhookBody(w, r, gitCfg.WebhookSecret, logger, "check suite", body) {
		return
	}

	if !validateAction(w, r, "check run", logger, []string{"requested", "rerequested"}) {
		return
	}

	form := r.Form
	action := form.Get("action")

	if action == "completed" {
		logger.Info("Maybe should check deployment here - unsure why this is needed")
	}

	var request checkSuiteRequest
	if err := json.Unmarshal(body, &request); err != nil {
		fmt.Println(string(body))
		log.Fatalf("Failed to parse request: %v", err)
		return
	}

	ensureDBCommitCheck(client, request, accessToken, dbService, logger)
}

func ensureDBCommitCheck(client *http.Client, request checkSuiteRequest, accessToken string, dbService db.Service, logger *log.Logger) {
	check := &db.CommitCheck{
		Revision:  request.CheckSuite.HeadSHA,
		Status:    "pending",
		StartedAt: time.Now(),
	}

	exists, err := dbService.Exists(db.CommitCheck{}, "Revision", check.Revision)
	if err != nil {
		logger.Errorf("Error when checking if commit check already exists: %v", err)
	} else if exists {
		logger.Warnf("Check already added to DB for revision: %s", check.Revision)
	} else {
		if err = createCheck(client, request, accessToken, dbService); err != nil {
			logger.Errorf("Failed creating check: %v", err)
		} else {
			dbService.GetDB().Create(&check)
			logger.Info("Successfully created check")
		}
	}
}
