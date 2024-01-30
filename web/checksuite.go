package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gorm.io/gorm"

	"github.com/ammesonb/ubiquiti-config-generator/config"
	"github.com/ammesonb/ubiquiti-config-generator/console_logger"
	"github.com/ammesonb/ubiquiti-config-generator/db"
)

// ProcessGitCheckSuite will create new check runs and update their statuses as appropriate
func ProcessGitCheckSuite(
	w http.ResponseWriter,
	r *http.Request,
	client *http.Client,
	logDB *gorm.DB,
	cfg *config.Config,
	accessToken string,
) {
	log := console_logger.DefaultLogger()
	var body []byte
	if _, err := r.Body.Read(body); err != nil {
		internalServerError(w, log, "Failed to read check suite request body", err)
		return
	}

	if !validateGitWebhookBody(w, r, cfg.Git.WebhookSecret, log, "check suite", body) {
		return
	}

	if !validateAction(w, r, "check run", log, []string{"requested", "rerequested"}) {
		return
	}

	form := r.Form
	action := form.Get("action")

	if action == "completed" {
		log.Info("Maybe should check deployment here - unsure why this is needed")
	}

	var request checkSuiteRequest
	if err := json.Unmarshal(body, &request); err != nil {
		fmt.Println(body)
		log.Fatalf("Failed to parse request: %v", err)
		return
	}

	check := &db.CommitCheck{
		Revision:  request.CheckSuite.HeadSHA,
		Status:    "pending",
		StartedAt: time.Now(),
	}

	exists, err := db.Exists(logDB, db.CommitCheck{}, "Revision", check.Revision)
	if err != nil {
		log.Errorf("Error when checking if commit check already exists: %v", err)
	} else if exists {
		log.Warnf("Check already added to DB for revision: %s", check.Revision)
	} else {
		logDB.Create(&check)
		if err = createCheck(client, logDB, request, accessToken); err != nil {
			log.Errorf("Failed creating check: %v", err)
		} else {
			log.Info("Successfully created check")
		}
	}
}
