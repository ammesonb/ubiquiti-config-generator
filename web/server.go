package web

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"

	"github.com/ammesonb/ubiquiti-config-generator/config"
)

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

// StartWebhookServer spins up a new server that responds to GitHub webhooks and also provides check/deployment status logs
func StartWebhookServer(cfg *config.Config, shutdownChannel chan os.Signal) {
	r := mux.NewRouter()
	r.PathPrefix("/static/").
		Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static")))).
		Methods("GET")

	// Also need check_run, push, and deployment
	r.Path("/").
		Methods("POST").
		HeadersRegexp("X-Hub-Signature-256", "^sha256=[a-fA-F0-9]{32}$").
		Headers("X-GitHub-Event", "check_suite").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ProcessGitCheckSuite(w, r, cfg) })

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
			log.Println(err)
		}
	}()

	// Block until we receive our signal.
	<-shutdownChannel

	// Wait 5 seconds - arbitrary, could use other contextual signals if needed
	// e.g. graphql/DB closing
	ctx, cancel := context.WithTimeout(context.Background(), 5)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	_ = srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)
}
