package web

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"

	"github.com/charmbracelet/log"
)

func validateGitBody(w http.ResponseWriter, r *http.Request, secret string, log *log.Logger, event string, body []byte) bool {
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
