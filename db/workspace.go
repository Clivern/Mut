// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

// Workspace role constants
const (
	WorkspaceRoleOwner    = "owner"
	WorkspaceRoleAdmin    = "admin"
	WorkspaceRoleRegular  = "regular"
	WorkspaceRoleReadOnly = "readonly"
)

// Workspace is the DB row for a workspace.
type Workspace struct {
	Id           Id
	Name         string
	Role         string
	MembersCount int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// WorkspaceRepository is the interface for workspace CRUD and membership lookups.
type WorkspaceRepository interface {
	CreateWithOwner(workspace *Workspace, ownerId Id) error
	GetById(id Id) (*Workspace, error)
	GetForUser(id Id, userId Id) (*Workspace, error)
	ListByUser(userId Id, limit, offset int) ([]*Workspace, error)
	CountByUser(userId Id) (int64, error)
	Update(workspace *Workspace) error
	Delete(id Id) error
}

// WorkspaceMeta is extra metadata for a workspace.
type WorkspaceMeta struct {
	Id          Id
	Key         string
	Value       string
	WorkspaceId Id
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// WorkspaceMetaRepository is the interface for workspace metadata CRUD.
type WorkspaceMetaRepository interface {
	Create(id Id, key, value string) error
	Get(id Id, key string) (*WorkspaceMeta, error)
	Update(id Id, key, value string) error
	Delete(id Id, key string) error
	ListByWorkspace(id Id) ([]*WorkspaceMeta, error)
	Upsert(id Id, key, value string) error
}

// WorkspaceMetaRepositoryPostgres implements WorkspaceMetaRepository for PostgreSQL
type WorkspaceMetaRepositoryPostgres struct {
	db *sql.DB
}

// WorkspaceRepositoryPostgres implements WorkspaceRepository for PostgreSQL
type WorkspaceRepositoryPostgres struct {
	db *sql.DB
}

// NewWorkspaceRepository returns the workspace repository for the current driver
func NewWorkspaceRepository(db *sql.DB) WorkspaceRepository {
	return &WorkspaceRepositoryPostgres{db: db}
}

// NewWorkspaceMetaRepository returns the repository for the current driver
func NewWorkspaceMetaRepository(db *sql.DB) WorkspaceMetaRepository {
	return &WorkspaceMetaRepositoryPostgres{db: db}
}

// scanWorkspaceMeta maps a row into workspace metadata. Column order must match the select list.
func scanWorkspaceMeta(meta *WorkspaceMeta) []any {
	return []any{
		&meta.Id,
		&meta.Key,
		&meta.Value,
		&meta.WorkspaceId,
		&meta.CreatedAt,
		&meta.UpdatedAt,
	}
}

// scanWorkspace maps a row into a workspace. Column order must match the select list.
func scanWorkspace(workspace *Workspace) []any {
	return []any{
		&workspace.Id,
		&workspace.Name,
		&workspace.Role,
		&workspace.MembersCount,
		&workspace.CreatedAt,
		&workspace.UpdatedAt,
	}
}

// --- Postgres ---

// CreateWithOwner inserts a workspace and makes ownerId its owner, in one statement
func (r *WorkspaceRepositoryPostgres) CreateWithOwner(workspace *Workspace, ownerId Id) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	membershipId, err := NewId()
	if err != nil {
		return err
	}

	workspace.Id = id

	_, err = r.db.Exec(
		`WITH created AS (
			INSERT INTO workspaces (id, name, created_at, updated_at)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		)
		INSERT INTO workspace_users (id, workspace_id, user_id, role)
		SELECT $5::uuid, created.id, $6::uuid, $7::varchar FROM created`,
		workspace.Id.String(),
		workspace.Name,
		workspace.CreatedAt,
		workspace.UpdatedAt,
		membershipId.String(),
		ownerId.String(),
		WorkspaceRoleOwner,
	)

	return err
}

// GetById returns a workspace without a role, for requests authorized by an access key
func (r *WorkspaceRepositoryPostgres) GetById(id Id) (*Workspace, error) {
	return QueryOne(
		r.db,
		`SELECT
			w.id, w.name, '',
			(SELECT COUNT(*) FROM workspace_users m WHERE m.workspace_id = w.id),
			w.created_at, w.updated_at
		FROM workspaces w
		WHERE w.id = $1`,
		scanWorkspace,
		id.String(),
	)
}

// GetForUser returns a workspace with the user's role, or nil when the user is not a member
func (r *WorkspaceRepositoryPostgres) GetForUser(id Id, userId Id) (*Workspace, error) {
	return QueryOne(
		r.db,
		`SELECT
			w.id, w.name, wu.role,
			(SELECT COUNT(*) FROM workspace_users m WHERE m.workspace_id = w.id),
			w.created_at, w.updated_at
		FROM workspaces w
		INNER JOIN workspace_users wu ON wu.workspace_id = w.id
		WHERE w.id = $1 AND wu.user_id = $2`,
		scanWorkspace,
		id.String(),
		userId.String(),
	)
}

// ListByUser returns the workspaces a user belongs to, newest first
func (r *WorkspaceRepositoryPostgres) ListByUser(userId Id, limit, offset int) ([]*Workspace, error) {
	return QueryMany(
		r.db,
		`SELECT
			w.id, w.name, wu.role,
			(SELECT COUNT(*) FROM workspace_users m WHERE m.workspace_id = w.id),
			w.created_at, w.updated_at
		FROM workspaces w
		INNER JOIN workspace_users wu ON wu.workspace_id = w.id
		WHERE wu.user_id = $1
		ORDER BY w.created_at DESC
		LIMIT $2 OFFSET $3`,
		scanWorkspace,
		userId.String(),
		limit,
		offset,
	)
}

// CountByUser returns how many workspaces a user belongs to
func (r *WorkspaceRepositoryPostgres) CountByUser(userId Id) (int64, error) {
	var total int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM workspace_users WHERE user_id = $1`,
		userId.String(),
	).Scan(&total)

	return total, err
}

// Update saves the workspace name
func (r *WorkspaceRepositoryPostgres) Update(workspace *Workspace) error {
	_, err := r.db.Exec(
		`UPDATE workspaces
		SET name = $1, updated_at = $2
		WHERE id = $3`,
		workspace.Name,
		workspace.UpdatedAt,
		workspace.Id.String(),
	)

	return err
}

// Delete removes a workspace and its memberships
func (r *WorkspaceRepositoryPostgres) Delete(id Id) error {
	_, err := r.db.Exec(
		`DELETE FROM workspaces WHERE id = $1`,
		id.String(),
	)

	return err
}

// Create inserts a workspace meta
func (r *WorkspaceMetaRepositoryPostgres) Create(id Id, key, value string) error {
	metaId, err := NewId()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		`INSERT INTO workspaces_meta (id, workspace_id, key, value)
		VALUES ($1, $2, $3, $4)`,
		metaId.String(),
		id.String(),
		key,
		value,
	)

	return err
}

// Get returns a workspace meta
func (r *WorkspaceMetaRepositoryPostgres) Get(id Id, key string) (*WorkspaceMeta, error) {
	return QueryOne(
		r.db,
		`SELECT id, key, value, workspace_id, created_at, updated_at
		FROM workspaces_meta
		WHERE workspace_id = $1 AND key = $2`,
		scanWorkspaceMeta,
		id.String(),
		key,
	)
}

// Update updates a workspace meta
func (r *WorkspaceMetaRepositoryPostgres) Update(id Id, key, value string) error {
	_, err := r.db.Exec(
		`UPDATE workspaces_meta
		SET
			value = $1,
			updated_at = $2
		WHERE workspace_id = $3 AND key = $4`,
		value,
		time.Now().UTC(),
		id.String(),
		key,
	)

	return err
}

// Delete removes a workspace meta
func (r *WorkspaceMetaRepositoryPostgres) Delete(id Id, key string) error {
	_, err := r.db.Exec(
		`DELETE FROM workspaces_meta WHERE workspace_id = $1 AND key = $2`,
		id.String(),
		key,
	)

	return err
}

// ListByWorkspace returns a list of workspace meta
func (r *WorkspaceMetaRepositoryPostgres) ListByWorkspace(id Id) ([]*WorkspaceMeta, error) {
	return QueryMany(
		r.db,
		`SELECT id, key, value, workspace_id, created_at, updated_at
		FROM workspaces_meta
		WHERE workspace_id = $1
		ORDER BY key`,
		scanWorkspaceMeta,
		id.String(),
	)
}

// Upsert upserts a workspace meta
func (r *WorkspaceMetaRepositoryPostgres) Upsert(id Id, key, value string) error {
	existing, err := r.Get(id, key)
	if err != nil {
		return err
	}
	if existing == nil {
		return r.Create(id, key, value)
	}

	return r.Update(id, key, value)
}
