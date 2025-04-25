package web

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
	"github.com/ammesonb/ubiquiti-config-generator/mocks"
	"github.com/ammesonb/ubiquiti-config-generator/services/db"

	"github.com/google/go-github/v70/github"

	"github.com/charmbracelet/log"
)

var (
	GREEN_CHECK = "&#9989;"
	RED_CROSS   = "&#10060;"
	WARNING     = "&#x26a0;"
)

var (
	errMarshalBody   = "failed to create %s body"
	errCreateRequest = "failed to create %s request for %s"
	errDoRequest     = "failed to do %s request for %s"
)

func validateGitWebhookBody(w http.ResponseWriter, r *http.Request, secret string, log *log.Logger, event string, body []byte) bool {
	hash := hmac.New(sha256.New, []byte(secret))

	if _, err := hash.Write(body); err != nil {
		internalServerError(w, log, "Failed to hash "+event+" body", err)
		return false
	}

	// Encode the output of the hash as a string then convert back to byte for comparison
	checksum := []byte(hex.EncodeToString(hash.Sum(nil)))

	if subtle.ConstantTimeCompare(checksum, []byte(r.Header.Get("X-Hub-Signature-256"))) != 1 {
		badRequest(w, log, "Body signature for "+event+" does not match header")
		return false
	}

	return true
}

func validateAction(w http.ResponseWriter, r *http.Request, what string, log *log.Logger, actions []string) bool {
	form := r.Form
	if !form.Has("action") {
		badRequest(w, log, what+" did not specify action")
		return false
	}

	formAction := form.Get("action")

	client := github.NewClient(nil)
	client.WithAuthToken("")

	client.Checks.CreateCheckRun(context.Background(), "willnorris", "ubiquiti-config-generator", github.CreateCheckRunOptions{})

	for _, action := range actions {
		if action == formAction {
			return true
		}
	}

	log.Infof("Ignoring %s action %s", what, formAction)
	return false
}

func addGitHubHeaders(req *http.Request, jwt string, accessToken string) {
	req.Header.Add("accept", "application/vnd.github.v3+json")
	if jwt != "" {
		req.Header.Add("Authorization", "Bearer "+jwt)
	} else if accessToken != "" {
		req.Header.Add("Authorization", "token "+jwt)
	}
}

func makeGitRequest(client mocks.WebClient, what string, jwt string, accessToken string, url, method string, body map[string]any) (*http.Response, error) {
	var encoded []byte = nil
	var err error

	if body != nil {
		encoded, err = json.Marshal(body)
		if err != nil {
			return nil, errors.ErrWithCtxParent(errMarshalBody, what, err)
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(encoded))
	if err != nil {
		return nil, errors.ErrWithVarCtxParent(errCreateRequest, err, method, what)
	}

	addGitHubHeaders(req, jwt, accessToken)

	response, err := client.Do(req)
	if err != nil {
		return nil, errors.ErrWithCtxParent(errDoRequest, what, err)
	}

	return response, nil
}

func createCheck(client mocks.WebClient, request checkSuiteRequest, accessToken string, dbService db.Service) error {
	response, err := makeGitRequest(
		client,
		"check run request",
		"",
		accessToken,
		request.Repository.URL+"/check_runs",
		"POST",
		map[string]any{
			"name":        "configuration-validator",
			"head_sha":    request.CheckSuite.HeadSHA,
			"details_url": request.Repository.App.ExternalURL + "/checks/" + request.CheckSuite.HeadSHA,
		},
	)
	if err != nil {
		return err
	}

	if response.StatusCode != 201 {
		dbService.GetDB().Create(&db.CheckLog{
			Revision:  request.CheckSuite.HeadSHA,
			Status:    db.StatusFailure,
			Timestamp: time.Now(),
			Message:   "Check run scheduled",
		})

		if err = setCommitStatus(
			client,
			accessToken,
			request.Repository.StatusesURL,
			request.CheckSuite.HeadSHA,
			db.StatusFailure,
			"Failed to schedule check run",
		); err != nil {
			dbService.GetDB().Create(&db.CheckLog{
				Revision:  request.CheckSuite.HeadSHA,
				Status:    db.StatusFailure,
				Timestamp: time.Now(),
				Message:   fmt.Sprintf("Failed to update commit status: %v", err),
			})
		}
	} else {
		dbService.GetDB().Create(&db.CheckLog{
			Revision:  request.CheckSuite.HeadSHA,
			Status:    db.StatusInfo,
			Timestamp: time.Now(),
			Message:   "Check run scheduled",
		})
	}

	return nil
}

func updateCheck(
	client mocks.WebClient,
	request checkRun,
	accessToken string,
	status string,
	extra map[string]any,
) error {
	extra["status"] = status
	response, err := makeGitRequest(
		client,
		"update check",
		"",
		accessToken,
		request.URL,
		"PATCH",
		extra,
	)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(response.Body)

	if err != nil {
		return errors.ErrWithParent("failed to read check update response body", err)
	} else if response.StatusCode != 200 {
		return errors.ErrWithVarCtx("failed to update check %d with status code %d: %s", request.ID, response.StatusCode, string(body))
	}

	return nil
}

func failCheck(
	client mocks.WebClient,
	request checkRun,
	accessToken string,
	extra map[string]any,
	err error,
) error {
	extra["completed_at"] = time.Now().Format(time.RFC3339)
	extra["conclusion"] = db.StatusFailure
	extra["output"] = map[string]any{
		"title":   "Configuration Validator",
		"summary": fmt.Sprintf("%s %v", RED_CROSS, "Check failed"),
		"text":    err.Error(),
	}

	return updateCheck(
		client,
		request,
		accessToken,
		db.StatusFailure,
		extra,
	)
}

func setCommitStatus(
	client mocks.WebClient,
	accessToken string,
	url string,
	revision,
	status,
	description string,
) error {
	response, err := makeGitRequest(
		client,
		"update commit status",
		"",
		accessToken,
		strings.ReplaceAll(url, "{sha}", revision),
		"POST",
		map[string]any{
			"state":       status,
			"description": description,
			"context":     "configuration-validator",
		},
	)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return errors.ErrWithParent("failed to read commit status response body", err)
	} else if response.StatusCode != 201 {
		return errors.ErrWithVarCtx(
			"failed to update commit %s to state %s, got HTTP %d: %s",
			revision,
			status,
			response.StatusCode,
			string(body),
		)
	}

	return nil
}

func upsertComment() {
}

func sendGitPost() {
}
