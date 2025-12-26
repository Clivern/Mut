// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"github.com/clivern/mut/db"
	"github.com/clivern/mut/module"
)

// Instance is the process-wide API wired at server startup.
var Instance *API

// API holds shared application modules for HTTP handlers.
type API struct {
	Auth     *module.Auth
	Profile  *module.Profile
	Setup    *module.Setup
	Settings *module.Settings
	APIKey   *module.APIKey
	Me       *module.Me
	User     *module.User
}

// New wires repositories and modules once for the process.
func New() *API {
	conn := db.GetDB()

	users := db.NewUserRepository(conn)
	sessions := db.NewSessionRepository(conn)
	configs := db.NewConfigRepository(conn)
	apiKeys := db.NewAPIKeyRepository(conn)

	a := &API{
		Auth:     module.NewAuth(users, sessions, configs),
		Profile:  module.NewProfile(users),
		Setup:    module.NewSetup(configs, users),
		Settings: module.NewSettings(configs),
		APIKey:   module.NewAPIKey(apiKeys),
		Me:       module.NewMe(apiKeys, users),
		User:     module.NewUser(users, sessions),
	}

	Instance = a

	return a
}
