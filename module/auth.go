// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/clivern/mut/conf"
	"github.com/clivern/mut/db"
	"github.com/clivern/mut/pkg/util"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/spf13/viper"
)

var (
	ErrAccountNotActive       = errors.New("account not active")
	ErrMaintenanceModeEnabled = errors.New("maintenance mode enabled")
	ErrFailedCheckMaintenance = errors.New("failed check maintenance")
	ErrFailedCreateSession    = errors.New("failed create session")

	ErrUserNotFound           = errors.New("user not found")
	ErrUserEmailAlreadyExists = errors.New("user with this email already exists")

	ErrFailedRevokeSessions = errors.New("failed revoke sessions")
)

// LoginResult is returned on a successful login.
type LoginResult struct {
	User          *db.User
	Session       *db.Session
	CookieOptions *util.CookieOptions
}

// OAuthIdentity is a verified identity from an OAuth provider.
type OAuthIdentity struct {
	Provider       string
	ProviderUserID string
	Email          string
	Name           string
}

// Auth handles user authentication and sessions.
type Auth struct {
	UserRepository    db.UserRepository
	SessionRepository db.SessionRepository
	ConfigRepository  db.ConfigRepository
}

// NewAuth creates an Auth module with the given repositories.
func NewAuth(users db.UserRepository, sessions db.SessionRepository, config db.ConfigRepository) *Auth {
	return &Auth{
		UserRepository:    users,
		SessionRepository: sessions,
		ConfigRepository:  config,
	}
}

// LoginWithOAuth finds or creates a user from an OAuth identity and opens a session.
func (a *Auth) LoginWithOAuth(ctx context.Context, identity *OAuthIdentity) (*LoginResult, error) {
	user, err := a.UserRepository.GetByProvider(identity.Provider, identity.ProviderUserID)
	if err != nil {
		log.Error().
			Err(err).
			Str("provider", identity.Provider).
			Msg("Failed to get user by provider")
		return nil, fmt.Errorf("get user by provider: %w", err)
	}

	foundByProvider := user != nil

	if user == nil {
		user, err = a.UserRepository.GetByEmail(identity.Email)
		if err != nil {
			log.Error().
				Err(err).
				Str("provider", identity.Provider).
				Msg("Failed to get user by email")
			return nil, fmt.Errorf("get user by email: %w", err)
		}
	}

	if user != nil && !foundByProvider && (user.Provider == db.UserProviderLocal || lo.IsEmpty(user.Provider)) {
		providerUserID := identity.ProviderUserID
		user.Provider = identity.Provider
		user.ProviderUserId = &providerUserID
		user.IsEmailVerified = true
		if lo.IsNotEmpty(identity.Name) {
			user.Name = identity.Name
		}

		err = a.UserRepository.Update(user)
		if err != nil {
			log.Error().
				Err(err).
				Str("userId", user.Id.String()).
				Str("provider", identity.Provider).
				Msg("Failed to convert local user to oauth")
			return nil, fmt.Errorf("convert local user to oauth: %w", err)
		}

		log.Info().
			Str("userId", user.Id.String()).
			Str("provider", identity.Provider).
			Msg("Local user converted to oauth")
	}

	if user == nil {
		pwd, err := util.GenerateSecureToken(20)
		if err != nil {
			log.Error().
				Err(err).
				Str("provider", identity.Provider).
				Msg("Failed to generate oauth password")
			return nil, fmt.Errorf("generate oauth password: %w", err)
		}

		hpass, err := util.HashPassword(pwd)
		if err != nil {
			log.Error().
				Err(err).
				Str("provider", identity.Provider).
				Msg("Failed to hash oauth password")
			return nil, fmt.Errorf("hash oauth password: %w", err)
		}

		now := time.Now().UTC()
		providerUserID := identity.ProviderUserID
		user = &db.User{
			Name:            GetDisplayName(identity),
			Email:           identity.Email,
			Password:        hpass,
			Provider:        identity.Provider,
			ProviderUserId:  &providerUserID,
			Role:            db.UserRoleRegular,
			IsActive:        true,
			IsEmailVerified: true,
			LastLoginAt:     now,
			Language:        db.UserLanguageEN,
			Theme:           db.UserThemeDefault,
		}

		err = a.UserRepository.Create(user)
		if err != nil {
			log.Error().
				Err(err).
				Str("provider", identity.Provider).
				Msg("Failed to create oauth user")
			return nil, fmt.Errorf("create oauth user: %w", err)
		}

		log.Info().
			Str("userId", user.Id.String()).
			Str("provider", identity.Provider).
			Msg("OAuth user created")
	}

	if !user.IsActive {
		log.Warn().
			Str("userId", user.Id.String()).
			Msg("OAuth login rejected for inactive user")
		return nil, ErrAccountNotActive
	}

	maintenance, err := a.ConfigRepository.Get("maintenance_mode")
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Msg("Failed to check maintenance mode")
		return nil, fmt.Errorf("%w: %v", ErrFailedCheckMaintenance, err)
	}
	if maintenance != nil && maintenance.Value == "1" && user.Role != db.UserRoleAdmin {
		log.Info().
			Str("userId", user.Id.String()).
			Msg("Login blocked by maintenance mode")
		return nil, ErrMaintenanceModeEnabled
	}

	err = a.UserRepository.UpdateLastLogin(user.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Msg("Failed to update last login")
		return nil, fmt.Errorf("update last login: %w", err)
	}

	token, err := util.GenerateSecureToken(32)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Msg("Failed to generate session token")
		return nil, fmt.Errorf("%w: %v", ErrFailedCreateSession, err)
	}

	session := &db.Session{
		Token:     token,
		UserId:    user.Id,
		ExpiresAt: time.Now().UTC().Add(conf.DefaultSessionDuration),
	}

	if appCtx, ok := AppCtxFrom(ctx); ok {
		if lo.IsNotEmpty(appCtx.IPAddress) {
			session.IPAddress = &appCtx.IPAddress
		}
		if lo.IsNotEmpty(appCtx.UserAgent) {
			session.UserAgent = &appCtx.UserAgent
		}
	}

	err = a.SessionRepository.Create(session)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Msg("Failed to create session")
		return nil, fmt.Errorf("%w: %v", ErrFailedCreateSession, err)
	}

	log.Info().
		Str("userId", user.Id.String()).
		Str("sessionId", session.Id.String()).
		Msg("User logged in")

	cookieOptions := lo.Ternary(
		strings.HasPrefix(viper.GetString("app.url"), "https://"),
		util.SecureCookieOptions(),
		util.DefaultCookieOptions(),
	)

	return &LoginResult{
		User:          user,
		Session:       session,
		CookieOptions: cookieOptions,
	}, nil
}

// GetDisplayName prefers the provider name and falls back to the email local part.
func GetDisplayName(identity *OAuthIdentity) string {
	name := identity.Name
	if lo.IsEmpty(name) {
		name, _, _ = strings.Cut(identity.Email, "@")
	}

	return name
}

// LogoutByToken revokes all sessions of the user that owns the session token.
func (a *Auth) LogoutByToken(token string) error {
	if lo.IsEmpty(token) {
		return nil
	}

	session, err := a.SessionRepository.GetByToken(token)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to get session")
		return fmt.Errorf("%w: %v", ErrFailedRevokeSessions, err)
	}
	if session == nil {
		return nil
	}

	return a.Logout(session.UserId)
}

// Logout revokes all sessions for a user.
func (a *Auth) Logout(userId db.Id) error {
	count, err := a.SessionRepository.DeleteExpired()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to delete expired sessions")
	} else if count > 0 {
		log.Info().
			Int64("count", count).
			Msg("Expired sessions deleted")
	}

	err = a.SessionRepository.DeleteByUserId(userId)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", userId.String()).
			Msg("Failed to revoke user sessions")
		return fmt.Errorf("%w: %v", ErrFailedRevokeSessions, err)
	}

	log.Info().
		Str("userId", userId.String()).
		Msg("User logged out")

	return nil
}
