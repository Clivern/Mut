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
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrAccountNotActive       = errors.New("account not active")
	ErrMaintenanceModeEnabled = errors.New("maintenance mode enabled")
	ErrFailedCheckMaintenance = errors.New("failed check maintenance")
	ErrFailedCreateSession    = errors.New("failed create session")

	ErrUserNotFound           = errors.New("user not found")
	ErrUserEmailAlreadyExists = errors.New("user with this email already exists")

	ErrFailedRevokeSessions = errors.New("failed revoke sessions")
)

// LoginRequest is the body for a password login.
type LoginRequest struct {
	Email      string `json:"email" validate:"required,email" label:"Email"`
	Password   string `json:"password" validate:"required" label:"Password"`
	RememberMe bool   `json:"rememberMe" label:"Remember Me"`
}

// LoginResult is returned on a successful login.
type LoginResult struct {
	User          *db.User
	Session       *db.Session
	CookieOptions *util.CookieOptions
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

// Login authenticates a user by email and password and creates a session.
func (a *Auth) Login(ctx context.Context, req *LoginRequest) (*LoginResult, error) {
	user, err := a.UserRepository.GetByEmail(req.Email)
	if err != nil {
		log.Error().
			Err(err).
			Str("email", req.Email).
			Msg("Failed to get user")
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil || !util.ComparePassword(user.Password, req.Password) {
		log.Info().
			Str("email", req.Email).
			Msg("Login rejected for invalid credentials")
		return nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		log.Warn().
			Str("userId", user.Id.String()).
			Msg("Login rejected for inactive user")
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

	duration := lo.Ternary(req.RememberMe, conf.RememberMeSessionDuration, conf.DefaultSessionDuration)
	session := &db.Session{
		Token:     token,
		UserId:    user.Id,
		ExpiresAt: time.Now().UTC().Add(duration),
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
	cookieOptions.MaxAge = lo.Ternary(req.RememberMe, int(conf.RememberMeSessionDuration/time.Second), 0)

	return &LoginResult{
		User:          user,
		Session:       session,
		CookieOptions: cookieOptions,
	}, nil
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
