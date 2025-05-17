package web

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net/http"
	"time"

	"github.com/ammesonb/ubiquiti-config-generator/services/configuration"
	"github.com/ammesonb/ubiquiti-config-generator/services/db"
	"github.com/ammesonb/ubiquiti-config-generator/services/filesystem"
	"github.com/ammesonb/ubiquiti-config-generator/services/github"
	"github.com/ammesonb/ubiquiti-config-generator/services/github_client"

	"github.com/charmbracelet/log"

	"github.com/gorilla/mux"
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

type webServer struct {
	configService configuration.Service
	fsService     filesystem.Service
	dbService     db.Service
	githubService github.Service
	logger        *log.Logger
}

func NewWebServer(logger *log.Logger, configService configuration.Service, fsService filesystem.Service, dbService db.Service, githubService github.Service) *webServer {
	return &webServer{
		configService: configService,
		fsService:     fsService,
		dbService:     dbService,
		githubService: githubService,
		logger:        logger,
	}
}

func (server *webServer) Start(ctx context.Context) error {
	server.logger.Debug("Initializing web server")

	r := mux.NewRouter()
	server.addStaticRoutes(r)
	server.addLogsRoutes(r)
	server.addGitWebhookRoutes(r)

	srv := server.listenAndServe(ctx, r)

	// Block until we receive our signal.
	<-ctx.Done()

	server.logger.Warn("Shutting down webhook server")
	return srv.Shutdown(ctx)
}

func (server *webServer) addStaticRoutes(r *mux.Router) {
	staticRouter := r.PathPrefix("/static/").Subrouter()
	staticRouter.PathPrefix("/").
		Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static")))).
		Methods("GET")
}

func (server *webServer) addLogsRoutes(r *mux.Router) {
	loggingConfig := server.configService.GetLoggingConfig()
	logsRouter := r.PathPrefix("/logs/").Subrouter()
	logsRouter.Use(
		func(next http.Handler) http.Handler {
			return basicAuthMiddleware(next, loggingConfig.User, loggingConfig.Password)
		},
	)
}

func (server *webServer) addGitWebhookRoutes(r *mux.Router) {
	gitConfig := server.configService.GetGitConfig()
	gitRouter := r.Path(gitConfig.WebhookRoute).
		Methods("POST").Subrouter()
	gitRouter.Use(
		func(h http.Handler) http.Handler {
			return github.NewAuthMiddleware(
				server.configService,
				server.logger,
				github_client.NewAPIClient(server.configService, server.fsService),
			).AccessTokenMiddleware(h)
		},
	)
	gitRouter.Handle("/", github.NewWebhookListener(server.configService, server.dbService, server.githubService, server.logger))
}

func (server *webServer) listenAndServe(ctx context.Context, r *mux.Router) *http.Server {
	listenConfig := server.configService.GetGitConfig()

	srv := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf("%s:%s", listenConfig.ListenIP, listenConfig.WebhookPort),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			server.logger.Fatalf("Error from web server listen/serve: %v", err)
		}
	}()

	return srv
}
