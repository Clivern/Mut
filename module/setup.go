// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"errors"
	"fmt"
	"time"

	"github.com/clivern/mut/db"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var (
	ErrPlatformAlreadyInstalled = errors.New("platform already installed")
	ErrFailedCompleteSetup      = errors.New("failed complete setup")
)

// Setup does the initial platform install and config.
type Setup struct {
	ConfigRepository db.ConfigRepository
	UserRepository   db.UserRepository
}

// SetupRequest is the body you send when running setup.
type SetupRequest struct {
	PlatformEmail string `json:"platformEmail" validate:"required,email,min=4,max=60" label:"Platform Email"`
}

// NewSetup creates a Setup with the given repositories.
func NewSetup(config db.ConfigRepository, users db.UserRepository) *Setup {
	return &Setup{
		ConfigRepository: config,
		UserRepository:   users,
	}
}

// IsInstalled checks whether the platform has been installed.
func (s *Setup) IsInstalled() bool {
	config, err := s.ConfigRepository.Get("is_installed")
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to check install status")
		return false
	}

	return config != nil
}

// Install performs the initial platform installation with the provided request.
func (s *Setup) Install(req *SetupRequest) error {
	if s.IsInstalled() {
		log.Warn().
			Msg("Setup requested on an installed platform")
		return ErrPlatformAlreadyInstalled
	}

	err := s.UserRepository.Create(&db.User{
		Name:            req.PlatformEmail,
		Email:           req.PlatformEmail,
		Provider:        db.UserProviderLocal,
		Role:            db.UserRoleAdmin,
		IsActive:        true,
		IsEmailVerified: true,
		LastLoginAt:     time.Now().UTC(),
		Language:        db.UserLanguageEN,
		Theme:           db.UserThemeDefault,
	})
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to create admin user")
		return fmt.Errorf("%w: create admin user: %v", ErrFailedCompleteSetup, err)
	}

	configs := []struct {
		key   string
		value string
	}{
		{"is_installed", "1"},
		{"platform_name", viper.GetString("app.name")},
		{"platform_url", viper.GetString("app.url")},
		{"platform_email", req.PlatformEmail},
		{"platform_description", "AI Powered Kubernetes operator for anomaly detection and resource optimization."},
		{"maintenance_mode", "0"},
	}

	for _, config := range configs {
		err = s.ConfigRepository.Create(config.key, config.value)
		if err != nil {
			log.Error().
				Err(err).
				Str("key", config.key).
				Msg("Failed to create setup config")
			return fmt.Errorf("%w: create %s config: %v", ErrFailedCompleteSetup, config.key, err)
		}
	}

	log.Info().
		Msg("Platform installed")

	return nil
}
