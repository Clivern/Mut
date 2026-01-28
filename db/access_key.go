// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

// AccessKey is the DB row for a workspace access key.
type AccessKey struct {
	Id          Id
	WorkspaceId Id
	Name        string
	KeyHash     string
	KeySuffix   string
	Permissions string
	ExpiresAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// AccessKeyRepository is the interface for workspace access key CRUD.
type AccessKeyRepository interface {
	Create(accessKey *AccessKey) error
	GetById(id Id) (*AccessKey, error)
	GetByKeyHash(keyHash string) (*AccessKey, error)
	ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*AccessKey, error)
	CountByWorkspaceId(workspaceId Id) (int64, error)
	Delete(id Id) error
}

// AccessKeyRepositoryPostgres implements AccessKeyRepository for PostgreSQL
type AccessKeyRepositoryPostgres struct {
	db *sql.DB
}

// NewAccessKeyRepository returns the access key repository for the current driver
func NewAccessKeyRepository(db *sql.DB) AccessKeyRepository {
	return &AccessKeyRepositoryPostgres{db: db}
}

// scanAccessKey maps a row into an access key. Column order must match the select list.
func scanAccessKey(k *AccessKey) []any {
	return []any{
		&k.Id,
		&k.WorkspaceId,
		&k.Name,
		&k.KeyHash,
		&k.KeySuffix,
		&k.Permissions,
		&k.ExpiresAt,
		&k.CreatedAt,
		&k.UpdatedAt,
	}
}

// --- Postgres ---

// Create inserts an access key and sets its Id and timestamps
func (r *AccessKeyRepositoryPostgres) Create(accessKey *AccessKey) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	accessKey.Id = id

	err = r.db.QueryRow(
		`INSERT INTO access_keys
		(id, workspace_id, name, key_hash, key_suffix, permissions, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at`,
		accessKey.Id.String(),
		accessKey.WorkspaceId.String(),
		accessKey.Name,
		accessKey.KeyHash,
		accessKey.KeySuffix,
		accessKey.Permissions,
		accessKey.ExpiresAt,
	).Scan(
		&accessKey.CreatedAt,
		&accessKey.UpdatedAt,
	)

	return err
}

// GetById returns an access key by Id, or nil when it does not exist
func (r *AccessKeyRepositoryPostgres) GetById(id Id) (*AccessKey, error) {
	return QueryOne(
		r.db,
		`SELECT
			id, workspace_id, name, key_hash, key_suffix, permissions, expires_at, created_at, updated_at
		FROM access_keys
		WHERE id = $1`,
		scanAccessKey,
		id.String(),
	)
}

// GetByKeyHash returns an unexpired access key by the hash of its secret
func (r *AccessKeyRepositoryPostgres) GetByKeyHash(keyHash string) (*AccessKey, error) {
	return QueryOne(
		r.db,
		`SELECT
			id, workspace_id, name, key_hash, key_suffix, permissions, expires_at, created_at, updated_at
		FROM access_keys
		WHERE key_hash = $1 AND (expires_at IS NULL OR expires_at > $2)`,
		scanAccessKey,
		keyHash,
		time.Now().UTC(),
	)
}

// ListByWorkspaceId returns a page of access keys for a workspace, newest first
func (r *AccessKeyRepositoryPostgres) ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*AccessKey, error) {
	return QueryMany(
		r.db,
		`SELECT
			id, workspace_id, name, key_hash, key_suffix, permissions, expires_at, created_at, updated_at
		FROM access_keys
		WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		scanAccessKey,
		workspaceId.String(),
		limit,
		offset,
	)
}

// CountByWorkspaceId returns how many access keys a workspace has
func (r *AccessKeyRepositoryPostgres) CountByWorkspaceId(workspaceId Id) (int64, error) {
	var total int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM access_keys WHERE workspace_id = $1`,
		workspaceId.String(),
	).Scan(&total)

	return total, err
}

// Delete removes an access key
func (r *AccessKeyRepositoryPostgres) Delete(id Id) error {
	_, err := r.db.Exec(
		`DELETE FROM access_keys WHERE id = $1`,
		id.String(),
	)

	return err
}
