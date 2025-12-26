// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package core

import (
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/clivern/mut/api"
	"github.com/clivern/mut/db"
	"github.com/clivern/mut/locale"
	"github.com/clivern/mut/middleware"
	"github.com/clivern/mut/module"
	"github.com/clivern/mut/pkg/util"

	"github.com/go-chi/chi/v5"
	cmid "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Setup creates and configures the HTTP server
func SetupServer(Static embed.FS, a *api.API) http.Handler {
	r := chi.NewRouter()

	r.Use(cmid.Recoverer)
	r.Use(cmid.Timeout(
		time.Duration(viper.GetInt("app.timeout")) * time.Second,
	))
	r.Use(middleware.ClientIP(viper.GetString("app.trusted_proxies")))
	r.Use(middleware.RequestId)
	r.Use(middleware.AppContext)
	r.Use(middleware.PrometheusMiddleware)
	r.Use(middleware.Logger)
	r.Use(middleware.RequestSizeLimit(
		int64(10 * 1024 * 1024),
	))
	r.Use(middleware.Auth())
	r.Use(middleware.ValidateParams)
	r.Get("/favicon.ico", func(w http.ResponseWriter, _ *http.Request) { // suppress browser favicon requests
		w.WriteHeader(http.StatusNoContent)
	})

	r.Group(func(r chi.Router) { // public — no auth required
		r.Get("/api/v1/public/_health", a.HealthAction)                                                  // liveness probe
		r.Get("/api/v1/public/_ready", a.ReadyAction)                                                    // readiness probe (db, deps)
		r.With(middleware.RateLimit(5, time.Minute)).Post("/api/v1/public/action/setup", a.SetupAction)  // initial app setup
		r.Get("/api/v1/public/action/setup/status", a.SetupStatusAction)                                 // setup completion status
		r.With(middleware.RateLimit(10, time.Minute)).Post("/api/v1/public/action/login", a.LoginAction) // user login
		r.Post("/api/v1/public/action/logout", a.LogoutAction)                                           // user logout
	})
	r.Get("/api/v1/me", a.GetMeAction) // current authenticated user
	r.Group(func(r chi.Router) {       // user profile
		r.Use(middleware.Protect(middleware.Config{Roles: []string{db.UserRoleAdmin, db.UserRoleRegular, db.UserRoleReadOnly}}))
		r.Get("/api/v1/action/profile", a.GetProfileAction)    // get user profile
		r.Put("/api/v1/action/profile", a.UpdateProfileAction) // update user profile
	})
	r.Group(func(r chi.Router) { // app settings — admin only
		r.Use(middleware.Protect(middleware.Config{Roles: []string{db.UserRoleAdmin}}))
		r.Put("/api/v1/action/settings", a.UpdateSettingsAction) // update app settings
		r.Get("/api/v1/action/settings", a.GetSettingsAction)    // get app settings
	})
	r.Group(func(r chi.Router) { // user API keys
		r.Use(middleware.Protect(middleware.Config{Roles: []string{db.UserRoleAdmin, db.UserRoleRegular}}))
		r.Post("/api/v1/apiKeys", a.CreateUserAPIKeyAction)              // create user API key
		r.Get("/api/v1/apiKeys", a.ListUserAPIKeysAction)                // list user API keys
		r.Get("/api/v1/apiKeys/{apiKeyId}", a.GetUserAPIKeyAction)       // get user API key
		r.Delete("/api/v1/apiKeys/{apiKeyId}", a.DeleteUserAPIKeyAction) // delete user API key
	})
	r.Group(func(r chi.Router) { // users — admin only
		r.Use(middleware.Protect(middleware.Config{Roles: []string{db.UserRoleAdmin}}))
		r.Post("/api/v1/users", a.CreateUserAction)            // create user
		r.Get("/api/v1/users", a.ListUsersAction)              // list users
		r.Get("/api/v1/users/{userId}", a.GetUserAction)       // get user
		r.Put("/api/v1/users/{userId}", a.UpdateUserAction)    // update user
		r.Delete("/api/v1/users/{userId}", a.DeleteUserAction) // delete user
	})

	r.With(middleware.BasicAuth(viper.GetString("app.metrics.username"), viper.GetString("app.metrics.secret"))).Get("/api/v1/public/_metrics", promhttp.Handler().ServeHTTP) // Prometheus metrics

	dist, err := fs.Sub(Static, "web/dist")
	if err != nil {
		panic(fmt.Sprintf(
			"Error while accessing dist files: %s",
			err.Error(),
		))
	}

	r.Handle("/assets/*", http.StripPrefix("/", http.FileServer(http.FS(dist)))) // static frontend assets
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {                    // SPA fallback — serve index.html
		if strings.HasPrefix(r.URL.Path, "/api/") { // an unknown API route is a JSON 404, never the dashboard
			util.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "route_not_found"),
			})
			return
		}

		indexFile, err := dist.Open("index.html")
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		defer indexFile.Close()

		stat, err := indexFile.Stat()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeContent(w, r, "index.html", stat.ModTime(), indexFile.(io.ReadSeeker))
	})

	return r
}

// Run starts the HTTP server with graceful shutdown support
func RunServer(Static embed.FS) error {
	err := db.InitDB(ReadWriteDatabase(), ReadOnlyDatabase()...)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	a := api.New()
	handler := SetupServer(Static, a)

	err = module.StartBus()
	if err != nil {
		return fmt.Errorf("failed to start nats bus: %w", err)
	}

	defer func() {
		err := db.CloseDB()
		if err != nil {
			log.Error().
				Err(err).
				Msg("Error closing database connection")
		}
	}()

	defer module.StopBus()

	timeout := time.Duration(viper.GetInt("app.timeout")) * time.Second

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", strconv.Itoa(viper.GetInt("app.port"))),
		Handler:           handler,
		MaxHeaderBytes:    230 * 1024,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadTimeout:       timeout + 5*time.Second,
		WriteTimeout:      timeout + 5*time.Second,
	}

	serr := make(chan error, 1)

	go func() {
		log.Info().
			Int("port", viper.GetInt("app.port")).
			Msg("Starting HTTP server")

		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			serr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serr:
		return fmt.Errorf("server error: %w", err)
	case sig := <-quit:
		log.Info().
			Str("signal", sig.String()).
			Msg("Received shutdown signal")

		shutdownTimeout := 30 * time.Second

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		log.Info().
			Dur("timeout", shutdownTimeout).
			Msg("Gracefully shutting down server")

		err := srv.Shutdown(ctx)
		if err != nil {
			return fmt.Errorf("server forced to shutdown: %w", err)
		}

		log.Info().
			Msg("Server shutdown complete")
	}

	return nil
}
