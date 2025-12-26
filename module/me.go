// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"time"

	"github.com/clivern/mut/db"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// Me resolves session and API key principals for /me.
type Me struct {
	APIKeyRepository db.APIKeyRepository
	UserRepository   db.UserRepository
}

// MeAPIKeyOwner is the user who owns an API key.
type MeAPIKeyOwner struct {
	Id    db.Id  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

// MeAPIKeyInfo is API key metadata returned by /me.
type MeAPIKeyInfo struct {
	Id        db.Id   `json:"id"`
	Name      string  `json:"name"`
	ExpiresAt *string `json:"expiresAt,omitempty"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
}

// MeAPIKeyResponse is returned when /me is called with X-API-Key.
type MeAPIKeyResponse struct {
	PrincipalType string        `json:"principalType"`
	Owner         MeAPIKeyOwner `json:"owner"`
	APIKey        MeAPIKeyInfo  `json:"apiKey"`
}

// MeSessionResponse is returned when /me is called with a session cookie.
type MeSessionResponse struct {
	PrincipalType string        `json:"principalType"`
	User          MeAPIKeyOwner `json:"user"`
}

// NewMe creates a /me module with the given repositories.
func NewMe(apiKeys db.APIKeyRepository, users db.UserRepository) *Me {
	return &Me{
		APIKeyRepository: apiKeys,
		UserRepository:   users,
	}
}

// GetBySession returns the session user.
func (m *Me) GetBySession(user *db.User) *MeSessionResponse {
	return &MeSessionResponse{
		PrincipalType: "session",
		User: MeAPIKeyOwner{
			Id:    user.Id,
			Email: user.Email,
			Name:  user.Name,
			Role:  user.Role,
		},
	}
}

// GetByAPIKey returns the API key owner and key metadata.
func (m *Me) GetByAPIKey(key string) (*MeAPIKeyResponse, error) {
	apiKey, err := m.APIKeyRepository.GetByKey(key)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to get API key")
		return nil, err
	}
	if apiKey == nil {
		return nil, ErrAPIKeyNotFound
	}

	owner, err := m.UserRepository.GetById(apiKey.UserId)
	if err != nil {
		log.Error().
			Err(err).
			Str("apiKeyId", apiKey.Id.String()).
			Msg("Failed to get API key owner")
		return nil, err
	}
	if owner == nil {
		return nil, ErrAPIKeyNotFound
	}

	return &MeAPIKeyResponse{
		PrincipalType: "api_key",
		Owner: MeAPIKeyOwner{
			Id:    owner.Id,
			Email: owner.Email,
			Name:  owner.Name,
			Role:  owner.Role,
		},
		APIKey: MeAPIKeyInfo{
			Id:        apiKey.Id,
			Name:      apiKey.Name,
			CreatedAt: apiKey.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt: apiKey.UpdatedAt.UTC().Format(time.RFC3339),
			ExpiresAt: lo.TernaryF(
				apiKey.ExpiresAt != nil,
				func() *string { return new(apiKey.ExpiresAt.UTC().Format(time.RFC3339)) },
				func() *string { return nil },
			),
		},
	}, nil
}
