// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"fmt"
	"strings"

	"github.com/clivern/mut/db"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// UpdateSettingsRequest is the body for updating app settings.
type UpdateSettingsRequest struct {
	PlatformName        string `json:"platformName" validate:"required,max=100" label:"Platform Name"`
	PlatformURL         string `json:"platformURL" validate:"required,url,max=255" label:"Platform URL"`
	PlatformEmail       string `json:"platformEmail" validate:"required,email,max=255" label:"Platform Email"`
	PlatformDescription string `json:"platformDescription" validate:"max=1024" label:"Platform Description"`
	MaintenanceMode     bool   `json:"maintenanceMode" label:"Maintenance Mode"`
}

// Settings manages app settings.
type Settings struct {
	ConfigRepository db.ConfigRepository
}

// NewSettings creates a settings module with the given config repo.
func NewSettings(config db.ConfigRepository) *Settings {
	return &Settings{ConfigRepository: config}
}

// Update persists the provided settings to the config repository.
func (s *Settings) Update(req *UpdateSettingsRequest) error {
	updates := map[string]string{
		"platform_name":        strings.TrimSpace(req.PlatformName),
		"platform_url":         strings.TrimSpace(req.PlatformURL),
		"platform_email":       strings.TrimSpace(req.PlatformEmail),
		"platform_description": strings.TrimSpace(req.PlatformDescription),
		"maintenance_mode":     lo.Ternary(req.MaintenanceMode, "1", "0"),
	}

	for key, value := range updates {
		err := s.ConfigRepository.Update(key, value)
		if err != nil {
			log.Error().
				Err(err).
				Str("key", key).
				Msg("Failed to update setting")
			return fmt.Errorf("failed to update config %s: %w", key, err)
		}
	}

	log.Info().
		Bool("maintenanceMode", req.MaintenanceMode).
		Msg("Settings updated")

	return nil
}

// GetSettings retrieves the current settings from the config repository.
func (s *Settings) GetSettings() (map[string]string, error) {
	keys := []string{
		"platform_name",
		"platform_url",
		"platform_email",
		"platform_description",
		"maintenance_mode",
	}

	settings := make(map[string]string, len(keys))

	for _, key := range keys {
		config, err := s.ConfigRepository.Get(key)
		if err != nil {
			log.Error().
				Err(err).
				Str("key", key).
				Msg("Failed to get setting")
			return nil, err
		}

		settings[key] = lo.TernaryF(config != nil, func() string { return config.Value }, func() string { return "" })
	}

	return settings, nil
}
