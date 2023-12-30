package web

import (
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
)

func internalServerError(w http.ResponseWriter, log *log.Logger, message string, err error) {
	w.WriteHeader(500)
	log.Errorf("%s: %v", message, err)
	_, _ = w.Write([]byte("Internal server error: " + strings.ToLower(message)))
}

func badRequest(w http.ResponseWriter, log *log.Logger, message string) {
	w.WriteHeader(400)
	log.Warnf(message)
	_, _ = w.Write([]byte("Bad request: " + strings.ToLower(message)))
}
