package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ammesonb/ubiquiti-config-generator/config"
	"github.com/ammesonb/ubiquiti-config-generator/services/configuration"
	"github.com/ammesonb/ubiquiti-config-generator/services/db"
	"github.com/charmbracelet/log"
)

// ProcessGitCheckRun will handle a requested check run and validate the new configuration
func ProcessGitCheckRun(
	w http.ResponseWriter,
	r *http.Request,
	client *http.Client,
	cfg *configuration.Config,
	devices map[string]*configuration.DeviceConfig,
	accessToken string,
	dbService db.Service,
	logger *log.Logger,
) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		internalServerError(w, logger, "Failed to read check run request body", err)
		return
	}

	if !validateGitWebhookBody(w, r, cfg.Git.WebhookSecret, logger, "check run", body) {
		return
	}

	form := r.Form
	if !form.Has("action") {
		badRequest(w, logger, "Check run did not specify action")
		return
	}

	if !validateAction(w, r, "check run", logger, []string{"created", "requested", "rerequested"}) {
		return
	}

	var checkrun checkRunRequest
	if err = json.Unmarshal(body, &checkrun); err != nil {
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
		dbService.GetDB().Create(&db.CheckLog{
			Revision:  revision,
			Status:    db.StatusFailure,
			Timestamp: time.Now(),
			Message:   "Failed to set check run to in progress",
		})
		logger.Error(err)
		return
	}

	// Start by cloning the repository
	dbService.GetDB().Create(&db.CheckLog{
		Revision:  revision,
		Status:    db.StatusInfo,
		Timestamp: time.Now(),
		Message: fmt.Sprintf(
			"Cloning repo %s using branch %s",
			checkrun.CheckRun.CheckSuite.Repository.Name,
			checkrun.CheckRun.CheckSuite.HeadBranch,
		),
	})

	repositoryDirectory, head, err := cloneRepo(
		checkrun.CheckRun.CheckSuite.Repository.CloneURL,
		checkrun.CheckRun.CheckSuite.HeadBranch,
	)
	if err != nil {
		logger.Error(err)
		return
	}

	logger.Infof("Cloned repository in %s", repositoryDirectory)

	changes, err := getChangedFiles(repositoryDirectory, head)
	if err != nil {
		logger.Error(err)
		return
	}

	for deviceName, device := range devices {
		if config.DeviceFilesChanged(device, changes) {
			// TODO: actually do things here
			fmt.Printf("Got changes for device %s", deviceName)
		}
	}

	// TODO: for each device:
	// TODO:   - parse and load files
	// TODO:   - perform validations
	// TODO: pull prod configs for devices
	// TODO: update check + post comment with diff
}
