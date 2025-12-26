// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"errors"
	"fmt"
	"time"

	"github.com/clivern/mut/db"
	"github.com/clivern/mut/pkg/util"

	"github.com/rs/zerolog/log"
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
	PlatformName  string `json:"platformName" validate:"required,min=2,max=50" label:"Platform Name"`
	PlatformURL   string `json:"platformURL" validate:"required,url,min=4,max=255" label:"Platform URL"`
	PlatformEmail string `json:"platformEmail" validate:"required,email,min=4,max=60" label:"Platform Email"`
	AdminName     string `json:"adminName" validate:"required,min=2,max=60" label:"Admin Name"`
	AdminEmail    string `json:"adminEmail" validate:"required,email,min=4,max=60" label:"Admin Email"`
	AdminPassword string `json:"adminPassword" validate:"required,strong_password,min=8,max=60" label:"Admin Password"`
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

	password, err := util.HashPassword(req.AdminPassword)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to hash admin password")
		return fmt.Errorf("%w: hash admin password: %v", ErrFailedCompleteSetup, err)
	}

	err = s.UserRepository.Create(&db.User{
		Name:        req.AdminName,
		Email:       req.AdminEmail,
		Password:    password,
		Role:        db.UserRoleAdmin,
		IsActive:    true,
		LastLoginAt: time.Now().UTC(),
		Language:    db.UserLanguageEN,
		Theme:       db.UserThemeDefault,
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
		{"platform_name", req.PlatformName},
		{"platform_url", req.PlatformURL},
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
