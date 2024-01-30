package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gorm.io/gorm"

	"github.com/ammesonb/ubiquiti-config-generator/config"
	"github.com/ammesonb/ubiquiti-config-generator/console_logger"
	"github.com/ammesonb/ubiquiti-config-generator/db"
)

// ProcessGitCheckRun will handle a requested check run and validate the new configuration
func ProcessGitCheckRun(
	w http.ResponseWriter,
	r *http.Request,
	client *http.Client,
	logDB *gorm.DB,
	cfg *config.Config,
	accessToken string,
) {
	log := console_logger.DefaultLogger()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		internalServerError(w, log, "Failed to read check run request body", err)
		return
	}

	if !validateGitWebhookBody(w, r, cfg.Git.WebhookSecret, log, "check run", body) {
		return
	}

	form := r.Form
	if !form.Has("action") {
		badRequest(w, log, "Check run did not specify action")
		return
	}

	if !validateAction(w, r, "check run", log, []string{"created", "requested", "rerequested"}) {
		return
	}

	var checkrun checkRunRequest
	if err = json.Unmarshal(body, &checkrun); err != nil {
		fmt.Println(body)
		log.Errorf("Failed to parse check run request: %v", err)
		return
	}

	revision := checkrun.CheckRun.CheckSuite.HeadSHA

	// Mark check as in progress
	if err = updateCheck(
		client,
		checkrun.CheckRun,
		accessToken,
		db.StatusInProgress,
		map[string]any{
			"started_at": time.Now().Format(time.RFC3339),
		},
	); err != nil {
		logDB.Create(&db.CheckLog{
			Revision:  revision,
			Status:    db.StatusFailure,
			Timestamp: time.Now(),
			Message:   "Failed to set check run to in progress",
		})
		log.Error(err)
		return
	}

	// Start by cloning the repository
	logDB.Create(&db.CheckLog{
		Revision:  revision,
		Status:    db.StatusInfo,
		Timestamp: time.Now(),
		Message: fmt.Sprintf(
			"Cloning repo %s using branch %s",
			checkrun.CheckRun.CheckSuite.Repository.Name,
			checkrun.CheckRun.CheckSuite.HeadBranch,
		),
	})

	repositoryDirectory, err := cloneRepo(
		checkrun.CheckRun.CheckSuite.Repository.CloneURL,
		checkrun.CheckRun.CheckSuite.HeadBranch,
	)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("Cloned repository in %s", repositoryDirectory)

	// TODO: analyze config.yaml in repository directory and repo to identify devices that need to have configs validated
	// TODO: load/parse configs specified for those devices
	//       - steps above needed for deployment too
	// TODO: validate those configs
	// TODO: pull prod configs for devices
	// TODO: update check + post comment with diff
}
