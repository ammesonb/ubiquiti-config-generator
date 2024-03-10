package web

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/log"

	rcontext "github.com/gorilla/context"
	"github.com/gorilla/mux"

	"github.com/ammesonb/ubiquiti-config-generator/config"
	"github.com/ammesonb/ubiquiti-config-generator/db"
)

var GIT_ACCESS_TOKEN_CONTEXT = "git_access_token"

func basicAuthMiddleware(next http.Handler, username, password string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()

		if !ok ||
			subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 ||
			subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="logs"`)
			w.WriteHeader(401)
			_, _ = w.Write([]byte("Unauthorised.\n"))
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func githubAccessTokenMiddleware(next http.Handler, logger *log.Logger, cfg *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		client := &http.Client{}
		jwt, err := makeJWT(cfg)
		if err != nil {
			logger.Errorf("Failed making JWT: %v", err)
			return
		}

		accessToken, err := getAccessToken(client, cfg.Git.AppID, jwt)
		if err != nil {
			logger.Errorf("Failed getting access token: %v", err)
			return
		}

		rcontext.Set(r, GIT_ACCESS_TOKEN_CONTEXT, accessToken)
		next.ServeHTTP(w, r)
	})
}

// StartWebhookServer spins up a new server that responds to GitHub webhooks and also provides check/deployment status logs
func StartWebhookServer(logger *log.Logger, cfg *config.Config, shutdownChannel chan os.Signal) {
	logDB, err := db.OpenDB(cfg.Logging.DBName)
	if err != nil {
		logger.Fatalf("Failed to open DB connection: %v", err)
		return
	}

	logger.Debug("Initializing web server")

	r := mux.NewRouter()
	staticRouter := r.PathPrefix("/static/").Subrouter()
	staticRouter.PathPrefix("/").
		Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static")))).
		Methods("GET")

	// TODO: Also need check_run, push, and deployment
	gitRouter := r.Path("/").
		Methods("POST").Subrouter()

	gitRouter.Use(
		func(next http.Handler) http.Handler {
			return githubAccessTokenMiddleware(next, logger, cfg)
		},
	)

	gitRouter.
		Path("/").
		HeadersRegexp("X-Hub-Signature-256", "^sha256=[a-fA-F0-9]{32}$").
		Headers("X-GitHub-Event", "check_suite").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ProcessGitCheckSuite(w, r, &http.Client{}, logDB, cfg, rcontext.Get(r, GIT_ACCESS_TOKEN_CONTEXT).(string))
		})

	srv := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf("%s:%d", cfg.Git.ListenIP, cfg.Git.WebhookPort),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.Fatalf("Error from web server listen/serve: %v", err)
		}
	}()

	// Block until we receive our signal.
	<-shutdownChannel

	// Wait 5 seconds - arbitrary, could use other contextual signals if needed
	// e.g. DB closing
	ctx, cancel := context.WithTimeout(context.Background(), 5)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	_ = srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Warn("shutting down")
}
