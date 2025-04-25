package github

import (
	"fmt"
	"net/http"

	"github.com/ammesonb/ubiquiti-config-generator/services/configuration"
	"github.com/charmbracelet/log"
	"github.com/google/go-github/v70/github"
)

type webhookListener struct {
	config configuration.Service
	logger *log.Logger
}

func (listener *webhookListener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, []byte(listener.config.GetGitConfig().WebhookSecret))
	if err != nil {
		listener.logger.Error(err)
		return
	}
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		listener.logger.Error(err)
		return
	}

	switch event := event.(type) {
	case *github.CheckSuiteEvent:
		fmt.Println(event)
		return
	default:
		listener.logger.Errorf("unhandled event type: %s", github.WebHookType(r))
	}
}

func NewWebhookListener(config configuration.Service, logger *log.Logger) http.Handler {
	return &webhookListener{
		config: config,
		logger: logger,
	}
}
