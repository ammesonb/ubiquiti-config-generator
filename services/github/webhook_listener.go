package github

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ammesonb/ubiquiti-config-generator/services/configuration"
	"github.com/ammesonb/ubiquiti-config-generator/services/db"
	"github.com/charmbracelet/log"
	"github.com/google/go-github/v70/github"
)

type webhookListener struct {
	config        configuration.Service
	githubService Service
	logger        *log.Logger
	dbService     db.Service
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
	case *github.PushEvent:
		// If not on configured branch, do nothing
		// Ref is formatted like refs/heads/main
		if strings.Count(event.GetRef(), "/") < 2 || strings.SplitN(event.GetRef(), "/", 3)[2] != listener.config.GetGitConfig().PrimaryBranch {
			listener.logger.Infof("Push event received on non-configured branch %s, doing nothing", event.GetRef())
			w.WriteHeader(http.StatusOK)
			return
		}
		listener.logger.Info("Push event received on configured branch, creating deployment")
		deployment, beforeSHA, afterSHA, err := listener.githubService.CreateDeployment(
			event.GetRepo().GetOwner().GetLogin(),
			event.GetRepo().GetName(),
			event.GetBefore(),
			event.GetHeadCommit().GetSHA(),
			r.Context(),
		)

		if err != nil {
			listener.logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			// TODO: remove this logger since superfluous
			listener.dbService.GetDB().Create(&db.Deployment{
				ID:           int(deployment.GetID()),
				FromRevision: beforeSHA,
				ToRevision:   afterSHA,
				Status:       db.StatusPending,
				StartedAt:    time.Now(),
			})
			listener.logger.Infof("Deployment %d created successfully", deployment.GetID())
		}

		// TODO: handle database records for deployment status
		// TODO: should we be writing response codes?
	default:
		listener.logger.Errorf("unhandled event type: %s", github.WebHookType(r))
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func NewWebhookListener(config configuration.Service, dbService db.Service, githubService Service, logger *log.Logger) http.Handler {
	return &webhookListener{
		config:        config,
		dbService:     dbService,
		githubService: githubService,
		logger:        logger,
	}
}
