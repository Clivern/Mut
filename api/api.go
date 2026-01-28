// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"github.com/clivern/mut/db"
	"github.com/clivern/mut/module"
	"github.com/clivern/mut/pkg/github"
	"github.com/clivern/mut/pkg/google"
	"github.com/clivern/mut/pkg/llm"
	"github.com/clivern/mut/pkg/resend"
	"github.com/clivern/mut/pkg/util"

	"github.com/spf13/viper"
)

// Instance is the process-wide API wired at server startup.
var Instance *API

// API holds shared application modules for HTTP handlers.
type API struct {
	Auth      *module.Auth
	Profile   *module.Profile
	Setup     *module.Setup
	Settings  *module.Settings
	APIKey    *module.APIKey
	Me        *module.Me
	User      *module.User
	Workspace *module.Workspace
	// WorkspaceMembers manages who belongs to a workspace and their roles
	WorkspaceMembers *module.WorkspaceMembers
	// AccessKey manages workspace access keys
	AccessKey *module.AccessKey
	// Invite manages workspace invites sent to email addresses
	Invite *module.Invite
	// Cluster serves read-only Kubernetes state for the connected cluster
	Cluster *module.Cluster
	// Chat serves the workspace assistant that answers questions about clusters
	Chat   *module.Chat
	GitHub *github.OAuth
	Google *google.OAuth
}

// New wires repositories and modules once for the process.
func New() *API {
	conn := db.GetDB()

	users := db.NewUserRepository(conn)
	sessions := db.NewSessionRepository(conn)
	configs := db.NewConfigRepository(conn)
	apiKeys := db.NewAPIKeyRepository(conn)
	workspaces := db.NewWorkspaceRepository(conn)
	workspaceMembers := db.NewWorkspaceMemberRepository(conn)
	invites := db.NewUserInviteRepository(conn)
	accessKeys := db.NewAccessKeyRepository(conn)

	cluster := module.NewCluster(db.NewClusterRepository(conn), db.NewClusterStateRepository(conn), db.NewWorkspaceRepository(conn))

	a := &API{
		Auth:             module.NewAuth(users, sessions, configs),
		Profile:          module.NewProfile(users),
		Setup:            module.NewSetup(configs, users),
		Settings:         module.NewSettings(configs),
		APIKey:           module.NewAPIKey(apiKeys),
		Me:               module.NewMe(apiKeys, users),
		User:             module.NewUser(users, sessions, resend.NewMailer()),
		Workspace:        module.NewWorkspace(workspaces),
		WorkspaceMembers: module.NewWorkspaceMembers(workspaceMembers, workspaces, users),
		AccessKey:        module.NewAccessKey(accessKeys, workspaces),
		Invite:           module.NewInvite(invites, users, workspaces, workspaceMembers, resend.NewMailer()),
		Cluster:          cluster,
		Chat:             module.NewChat(db.NewChatRepository(conn), module.NewChatTools(cluster), func() module.ChatModel { return llm.NewChatClient() }),
		GitHub: github.NewOAuth(github.OAuthConfig{
			ClientID:     viper.GetString("app.oauth.github.client_id"),
			ClientSecret: viper.GetString("app.oauth.github.client_secret"),
			RedirectURL:  util.AppURL("/api/v1/public/action/oauth/github/callback"),
			Scopes:       []string{"read:user", "user:email"},
		}),
		Google: google.NewOAuth(google.OAuthConfig{
			ClientID:     viper.GetString("app.oauth.google.client_id"),
			ClientSecret: viper.GetString("app.oauth.google.client_secret"),
			RedirectURL:  util.AppURL("/api/v1/public/action/oauth/google/callback"),
			Scopes:       []string{"openid", "email", "profile"},
		}),
	}

	Instance = a

	return a
}
