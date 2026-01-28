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
		r.Get("/api/v1/public/_health", a.HealthAction)                                   // liveness probe
		r.Get("/api/v1/public/_ready", a.ReadyAction)                                     // readiness probe (db, deps)
		r.Post("/api/v1/public/action/setup", a.SetupAction)                              // initial app setup
		r.Get("/api/v1/public/action/setup/status", a.SetupStatusAction)                  // setup completion status
		r.Post("/api/v1/public/action/logout", a.LogoutAction)                            // user logout
		r.Get("/api/v1/public/action/oauth/github", a.GitHubOAuthStartAction)             // start GitHub sign-in
		r.Get("/api/v1/public/action/oauth/github/callback", a.GitHubOAuthCallbackAction) // GitHub sign-in callback
		r.Get("/api/v1/public/action/oauth/google", a.GoogleOAuthStartAction)             // start Google sign-in
		r.Get("/api/v1/public/action/oauth/google/callback", a.GoogleOAuthCallbackAction) // Google sign-in callback
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

	r.With(middleware.Protect(middleware.Config{
		Roles:     []string{db.UserRoleAdmin, db.UserRoleRegular, db.UserRoleReadOnly},
		Workspace: true,
		Perm:      module.AccessKeyPermissionClustersRead,
	})).Get("/api/v1/workspaces/{workspaceId}/clusters/{clusterId}/state", a.GetClusterStateAction) // cluster state snapshot, by member or access key

	r.With(middleware.Protect(middleware.Config{
		Perm: module.AccessKeyPermissionClustersWrite,
	})).Put("/api/v1/workspaces/{workspaceId}/clusters/{clusterId}/state", a.SaveClusterStateAction) // push cluster state, by access key from an in-cluster agent

	r.Group(func(r chi.Router) { // assistant chats are private to the user who holds them
		r.Use(middleware.Protect(middleware.Config{
			Roles:     []string{db.UserRoleAdmin, db.UserRoleRegular},
			Workspace: true,
		}))
		r.Get("/api/v1/workspaces/{workspaceId}/chats", a.ListChatsAction)                          // list the user's chats
		r.Post("/api/v1/workspaces/{workspaceId}/chats", a.CreateChatAction)                        // start a chat
		r.Get("/api/v1/workspaces/{workspaceId}/chats/{chatId}", a.GetChatAction)                   // get a chat with its messages
		r.Delete("/api/v1/workspaces/{workspaceId}/chats/{chatId}", a.DeleteChatAction)             // delete a chat
		r.Post("/api/v1/workspaces/{workspaceId}/chats/{chatId}/messages", a.SendChatMessageAction) // ask in a chat
	})

	r.Group(func(r chi.Router) { // workspaces — any signed-in user, limited to workspaces they belong to
		r.Use(middleware.Protect(middleware.Config{Roles: []string{db.UserRoleAdmin, db.UserRoleRegular, db.UserRoleReadOnly}}))
		r.Post("/api/v1/workspaces", a.CreateWorkspaceAction)                 // create workspace
		r.Get("/api/v1/workspaces", a.ListWorkspacesAction)                   // list user workspaces
		r.Put("/api/v1/workspaces/{workspaceId}", a.UpdateWorkspaceAction)    // update workspace
		r.Delete("/api/v1/workspaces/{workspaceId}", a.DeleteWorkspaceAction) // delete workspace
	})

	r.With(middleware.Protect(middleware.Config{
		Roles:     []string{db.UserRoleAdmin, db.UserRoleRegular, db.UserRoleReadOnly},
		Workspace: true,
		Perm:      module.AccessKeyPermissionWorkspaceRead,
	})).Get("/api/v1/workspaces/{workspaceId}", a.GetWorkspaceAction) // get workspace, by member or access key
	r.With(middleware.Protect(middleware.Config{
		Roles:     []string{db.UserRoleAdmin, db.UserRoleRegular, db.UserRoleReadOnly},
		Workspace: true,
		Perm:      module.AccessKeyPermissionMembersRead,
	})).Get("/api/v1/workspaces/{workspaceId}/members", a.ListWorkspaceMembersAction) // list members, by member or access key

	// Workspace management: owners and admins only, enforced per route by its permission.
	r.With(managePerm(module.PermissionMembersManage)).Post("/api/v1/workspaces/{workspaceId}/invites", a.CreateInviteAction)                             // invite a user by email
	r.With(managePerm(module.PermissionMembersManage)).Get("/api/v1/workspaces/{workspaceId}/invites", a.ListInvitesAction)                               // list invites
	r.With(managePerm(module.PermissionMembersManage)).Get("/api/v1/workspaces/{workspaceId}/invites/{inviteId}", a.GetInviteAction)                      // get invite
	r.With(managePerm(module.PermissionMembersManage)).Delete("/api/v1/workspaces/{workspaceId}/invites/{inviteId}", a.DeleteInviteAction)                // revoke invite
	r.With(managePerm(module.PermissionMembersManage)).Put("/api/v1/workspaces/{workspaceId}/members/{memberUserId}", a.UpdateWorkspaceMemberAction)      // update member role
	r.With(managePerm(module.PermissionMembersManage)).Delete("/api/v1/workspaces/{workspaceId}/members/{memberUserId}", a.RemoveWorkspaceMemberAction)   // remove member
	r.With(managePerm(module.PermissionAccessKeysManage)).Post("/api/v1/workspaces/{workspaceId}/accessKeys", a.CreateWorkspaceAccessKeyAction)           // create access key
	r.With(managePerm(module.PermissionAccessKeysManage)).Get("/api/v1/workspaces/{workspaceId}/accessKeys", a.ListWorkspaceAccessKeysAction)             // list access keys
	r.With(managePerm(module.PermissionAccessKeysManage)).Get("/api/v1/workspaces/{workspaceId}/accessKeys/{keyId}", a.GetWorkspaceAccessKeyAction)       // get access key
	r.With(managePerm(module.PermissionAccessKeysManage)).Delete("/api/v1/workspaces/{workspaceId}/accessKeys/{keyId}", a.DeleteWorkspaceAccessKeyAction) // revoke access key
	r.With(managePerm(module.PermissionClustersManage)).Post("/api/v1/workspaces/{workspaceId}/clusters", a.CreateClusterAction)                          // register a cluster
	r.With(managePerm(module.PermissionClustersManage)).Get("/api/v1/workspaces/{workspaceId}/clusters", a.ListClustersAction)                            // list registered clusters
	r.With(managePerm(module.PermissionClustersManage)).Get("/api/v1/workspaces/{workspaceId}/clusters/{clusterId}", a.GetClusterDetailAction)            // get a registered cluster
	r.With(managePerm(module.PermissionClustersManage)).Put("/api/v1/workspaces/{workspaceId}/clusters/{clusterId}", a.UpdateClusterAction)               // update a registered cluster
	r.With(managePerm(module.PermissionClustersManage)).Delete("/api/v1/workspaces/{workspaceId}/clusters/{clusterId}", a.DeleteClusterAction)            // remove a registered cluster

	r.With(middleware.BasicAuth(
		viper.GetString("app.metrics.username"),
		viper.GetString("app.metrics.secret"),
	)).Get("/api/v1/public/_metrics", promhttp.Handler().ServeHTTP) // Prometheus metrics

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

// managePerm guards a workspace management route with a permission that only owners and admins hold.
func managePerm(permission string) func(http.Handler) http.Handler {
	return middleware.Protect(middleware.Config{
		Roles:     []string{db.UserRoleAdmin, db.UserRoleRegular, db.UserRoleReadOnly},
		Workspace: true,
		Perm:      permission,
	})
}
