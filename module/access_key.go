// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/clivern/mut/db"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// Access key permissions. A workspace access key may only carry these.
const (
	AccessKeyPermissionWorkspaceRead = "workspace:read"
	AccessKeyPermissionMembersRead   = "members:read"
	AccessKeyPermissionClustersRead  = "clusters:read"
	AccessKeyPermissionClustersWrite = "clusters:write"
)

// Workspace management permissions. Only owners and admins hold them, and access keys never do.
const (
	PermissionMembersManage    = "members:manage"
	PermissionAccessKeysManage = "accessKeys:manage"
	PermissionClustersManage   = "clusters:manage"
)

// AccessKeyPrefix starts every workspace access key so leaked keys are easy to spot.
const AccessKeyPrefix = "mut_"

var (
	ErrAccessKeyNotFound = errors.New("access key not found")
	ErrAccessKeyInvalid  = errors.New("access key invalid")
)

// AccessKey manages workspace access keys. Only the hash of each secret is stored.
type AccessKey struct {
	AccessKeyRepository db.AccessKeyRepository
	WorkspaceRepository db.WorkspaceRepository
}

// CreateAccessKeyRequest is the body for creating a workspace access key.
type CreateAccessKeyRequest struct {
	Name        string   `json:"name" validate:"required,min=2,max=60" label:"Name"`
	ExpiresAt   string   `json:"expiresAt" validate:"omitempty,max=64" label:"Expires at"`
	Permissions []string `json:"permissions" validate:"required,min=1,dive,oneof=workspace:read members:read clusters:read clusters:write" label:"Permissions"`
}

// AccessKeyResponse is a workspace access key shaped for API responses.
type AccessKeyResponse struct {
	Id          db.Id    `json:"id"`
	WorkspaceId db.Id    `json:"workspaceId"`
	Name        string   `json:"name"`
	MaskedKey   string   `json:"maskedKey"`
	Permissions []string `json:"permissions"`
	ExpiresAt   *string  `json:"expiresAt"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
	Key         string   `json:"key,omitempty"`
}

// ListAccessKeysResult is a page of workspace access keys.
type ListAccessKeysResult struct {
	AccessKeys []*AccessKeyResponse
	Total      int64
}

// NewAccessKey creates an access key module with the given repositories.
func NewAccessKey(accessKeys db.AccessKeyRepository, workspaces db.WorkspaceRepository) *AccessKey {
	return &AccessKey{
		AccessKeyRepository: accessKeys,
		WorkspaceRepository: workspaces,
	}
}

// toAccessKeyResponse shapes an access key row for API responses. An empty key hides the secret.
func toAccessKeyResponse(item *db.AccessKey, key string) *AccessKeyResponse {
	return &AccessKeyResponse{
		Id:          item.Id,
		WorkspaceId: item.WorkspaceId,
		Name:        item.Name,
		MaskedKey:   AccessKeyPrefix + "****" + item.KeySuffix,
		Permissions: strings.Split(item.Permissions, ","),
		CreatedAt:   item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   item.UpdatedAt.UTC().Format(time.RFC3339),
		Key:         key,
		ExpiresAt: lo.TernaryF(
			item.ExpiresAt != nil,
			func() *string { return new(item.ExpiresAt.UTC().Format(time.RFC3339)) },
			func() *string { return nil },
		),
	}
}

// hashAccessKey returns the SHA-256 hex digest that is stored in place of the secret.
func hashAccessKey(key string) string {
	sum := sha256.Sum256([]byte(key))

	return hex.EncodeToString(sum[:])
}

// parseAccessKeyExpiresAt parses the optional RFC 3339 expiry. An empty value means the key never expires.
func parseAccessKeyExpiresAt(value string) (*time.Time, error) {
	if lo.IsEmpty(value) {
		return nil, nil
	}

	expiresAt, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, ErrInvalidExpiresAt
	}

	return &expiresAt, nil
}

// generateAccessKey returns a new random access key secret.
func generateAccessKey() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("generate access key: %w", err)
	}

	return AccessKeyPrefix + hex.EncodeToString(bytes), nil
}

// CreateAccessKey creates an access key for a workspace. The raw key is returned once and never stored.
func (a *AccessKey) CreateAccessKey(workspaceId db.Id, user *db.User, req *CreateAccessKeyRequest) (*AccessKeyResponse, error) {
	expiresAt, err := parseAccessKeyExpiresAt(req.ExpiresAt)
	if err != nil {
		return nil, err
	}

	key, err := generateAccessKey()
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to generate access key")
		return nil, err
	}

	item := &db.AccessKey{
		WorkspaceId: workspaceId,
		Name:        req.Name,
		KeyHash:     hashAccessKey(key),
		KeySuffix:   key[len(key)-4:],
		Permissions: strings.Join(req.Permissions, ","),
		ExpiresAt:   expiresAt,
	}

	err = a.AccessKeyRepository.Create(item)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to create access key")
		return nil, err
	}

	log.Info().
		Str("userId", user.Id.String()).
		Str("workspaceId", workspaceId.String()).
		Str("accessKeyId", item.Id.String()).
		Msg("Access key created")

	return toAccessKeyResponse(item, key), nil
}

// ListAccessKeys returns a page of a workspace's access keys (masked, never the secret).
func (a *AccessKey) ListAccessKeys(workspaceId db.Id, user *db.User, limit, offset int) (*ListAccessKeysResult, error) {
	total, err := a.AccessKeyRepository.CountByWorkspaceId(workspaceId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to count access keys")
		return nil, err
	}

	keys, err := a.AccessKeyRepository.ListByWorkspaceId(workspaceId, limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to list access keys")
		return nil, err
	}

	list := make([]*AccessKeyResponse, 0, len(keys))
	for _, item := range keys {
		list = append(list, toAccessKeyResponse(item, ""))
	}

	return &ListAccessKeysResult{AccessKeys: list, Total: total}, nil
}

// GetAccessKey returns one access key of a workspace (never the secret).
func (a *AccessKey) GetAccessKey(workspaceId db.Id, user *db.User, id db.Id) (*AccessKeyResponse, error) {
	item, err := a.AccessKeyRepository.GetById(id)
	if err != nil {
		log.Error().
			Err(err).
			Str("accessKeyId", id.String()).
			Msg("Failed to get access key")
		return nil, err
	}
	if item == nil || item.WorkspaceId != workspaceId {
		return nil, ErrAccessKeyNotFound
	}

	return toAccessKeyResponse(item, ""), nil
}

// DeleteAccessKey revokes a workspace access key. Its secret stops working immediately.
func (a *AccessKey) DeleteAccessKey(workspaceId db.Id, user *db.User, id db.Id) error {
	item, err := a.AccessKeyRepository.GetById(id)
	if err != nil {
		log.Error().
			Err(err).
			Str("accessKeyId", id.String()).
			Msg("Failed to get access key")
		return err
	}
	if item == nil || item.WorkspaceId != workspaceId {
		return ErrAccessKeyNotFound
	}

	err = a.AccessKeyRepository.Delete(item.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("accessKeyId", item.Id.String()).
			Msg("Failed to delete access key")
		return err
	}

	log.Info().
		Str("userId", user.Id.String()).
		Str("workspaceId", workspaceId.String()).
		Str("accessKeyId", item.Id.String()).
		Msg("Access key deleted")

	return nil
}

// AuthenticateAccessKey returns the unexpired access key that matches the raw secret.
func (a *AccessKey) AuthenticateAccessKey(key string) (*db.AccessKey, error) {
	item, err := a.AccessKeyRepository.GetByKeyHash(hashAccessKey(key))
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to get access key by hash")
		return nil, err
	}
	if item == nil {
		return nil, ErrAccessKeyInvalid
	}

	return item, nil
}
