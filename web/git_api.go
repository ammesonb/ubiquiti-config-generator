package web

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"

	"github.com/charmbracelet/log"

	"github.com/ammesonb/ubiquiti-config-generator/config"
	"github.com/ammesonb/ubiquiti-config-generator/db"
)

var (
	GREEN_CHECK = "&#9989;"
	RED_CROSS   = "&#10060;"
	WARNING     = "&#x26a0;"
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

func makeGitRequest(client *http.Client, what string, jwt string, accessToken string, url, method string, body map[string]any) (*http.Response, error) {
	var encoded []byte = nil
	var err error

	if body != nil {
		encoded, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to create %s body: %v", what, err)
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(encoded))
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %v", what, err)
	}

	addGitHubHeaders(req, jwt, accessToken)

	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request for %s: %v", what, err)
	}

	return response, nil
}

func makeJWT(cfg *config.Config) (string, error) {
	keyfile, err := os.Open(cfg.Git.PrivateKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to open keyfile: %v", err)
	}

	encoded, err := io.ReadAll(keyfile)
	if err != nil {
		return "", fmt.Errorf("unable to read from keyfile: %v", err)
	}

	block, _ := pem.Decode(encoded)
	x509Encoded := block.Bytes
	privateKey, _ := x509.ParseECPrivateKey(x509Encoded)

	t := jwt.New(jwt.SigningMethodRS256)
	t = jwt.NewWithClaims(jwt.SigningMethodRS256,
		jwt.MapClaims{
			"iat": time.Now().Unix(),
			"exp": time.Now().Unix() + 600,
			"iss": cfg.Git.AppId,
		})
	return t.SignedString(privateKey)
}

type installationsResponse struct {
	Installations []installation `json:"installations"`
}

type installation struct {
	AppID           int32  `json:"app_id"`
	AccessTokensURL string `json:"access_tokens_url"`
}

type accessTokenResponse struct {
	AccessToken string `json:"token"`
}

func getAccessToken(client *http.Client, appID int32, jwt string) (string, error) {
	response, err := makeGitRequest(
		client,
		"installation",
		jwt,
		"",
		"https://api.github.com/app/installations",
		"GET",
		nil,
	)
	if err != nil {
		return "", err
	}

	installationBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read installations body: %v", err)
	}

	var appInstalls installationsResponse
	if err = json.Unmarshal(installationBody, &appInstalls); err != nil {
		fmt.Println(string(installationBody))
		return "", fmt.Errorf("failed to parse installation response: %v", err)
	} else if len(appInstalls.Installations) == 0 {
		return "", fmt.Errorf("response did not return any installations; need at least one")
	}

	accessTokenURL := ""
	for _, install := range appInstalls.Installations {
		if install.AppID == appID {
			accessTokenURL = install.AccessTokensURL
		}
	}

	if accessTokenURL == "" {
		fmt.Println(appInstalls.Installations)
		return "", fmt.Errorf("could not find an installation matching AppID %d", appID)
	}

	response, err = makeGitRequest(
		client,
		"access token",
		jwt,
		"",
		appInstalls.Installations[0].AccessTokensURL,
		"GET",
		nil,
	)
	if err != nil {
		return "", err
	}

	tokenBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read token response body: %v", err)
	}

	var tokenResponse accessTokenResponse
	if err = json.Unmarshal(tokenBody, &tokenResponse); err != nil {
		fmt.Println(tokenBody)
		return "", fmt.Errorf("failed to parse access token response: %v", err)
	}

	return tokenResponse.AccessToken, nil
}

func createCheck(client *http.Client, logDB *gorm.DB, request checkSuiteRequest, accessToken string) error {
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
		logDB.Create(&db.CheckLog{
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
			logDB.Create(&db.CheckLog{
				Revision:  request.CheckSuite.HeadSHA,
				Status:    db.StatusFailure,
				Timestamp: time.Now(),
				Message:   fmt.Sprintf("Failed to update commit status: %v", err),
			})
		}
	} else {
		logDB.Create(&db.CheckLog{
			Revision:  request.CheckSuite.HeadSHA,
			Status:    db.StatusInfo,
			Timestamp: time.Now(),
			Message:   "Check run scheduled",
		})
	}

	return nil
}

func updateCheck(
	client *http.Client,
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

	if response.StatusCode != 200 {
		return fmt.Errorf("failed to update check %d with status code %d: %s", request.ID, response.StatusCode, string(body))
	}

	return nil
}

func failCheck(
	client *http.Client,
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
	client *http.Client,
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

	if response.StatusCode != 201 {
		return fmt.Errorf("failed to update commit %s to state %s, got HTTP %d: %s", revision, status, response.StatusCode, string(body))
	}

	return nil
}

func upsertComment() {
}

func sendGitPost() {
}
