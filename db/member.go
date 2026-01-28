// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

// WorkspaceMember is a workspace_users row joined with the member's user, so responses can show who the member is without a second lookup.
type WorkspaceMember struct {
	WorkspaceId Id
	UserId      Id
	Name        string
	Email       string
	Role        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// WorkspaceMemberRepository is the interface for workspace membership CRUD.
type WorkspaceMemberRepository interface {
	Create(workspaceId Id, userId Id, role string) error
	GetByWorkspaceAndUser(workspaceId Id, userId Id) (*WorkspaceMember, error)
	ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*WorkspaceMember, error)
	CountByWorkspaceId(workspaceId Id) (int64, error)
	UpdateRole(workspaceId Id, userId Id, role string) error
	Delete(workspaceId Id, userId Id) error
}

// WorkspaceMemberRepositoryPostgres implements WorkspaceMemberRepository for PostgreSQL
type WorkspaceMemberRepositoryPostgres struct {
	db *sql.DB
}

// NewWorkspaceMemberRepository returns the workspace membership repository for the current driver
func NewWorkspaceMemberRepository(db *sql.DB) WorkspaceMemberRepository {
	return &WorkspaceMemberRepositoryPostgres{db: db}
}

// scanWorkspaceMember maps a row into a workspace member. Column order must match the select list.
func scanWorkspaceMember(m *WorkspaceMember) []any {
	return []any{
		&m.WorkspaceId,
		&m.UserId,
		&m.Name,
		&m.Email,
		&m.Role,
		&m.CreatedAt,
		&m.UpdatedAt,
	}
}

// --- Postgres ---

// Create adds a user to a workspace with the given role
func (r *WorkspaceMemberRepositoryPostgres) Create(workspaceId Id, userId Id, role string) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		`INSERT INTO workspace_users (id, workspace_id, user_id, role)
		VALUES ($1, $2, $3, $4)`,
		id.String(),
		workspaceId.String(),
		userId.String(),
		role,
	)

	return err
}

// GetByWorkspaceAndUser returns a membership, or nil when the user is not a member
func (r *WorkspaceMemberRepositoryPostgres) GetByWorkspaceAndUser(workspaceId Id, userId Id) (*WorkspaceMember, error) {
	return QueryOne(
		r.db,
		`SELECT
			wu.workspace_id, u.id, u.name, u.email, wu.role, wu.created_at, wu.updated_at
		FROM workspace_users wu
		INNER JOIN users u ON u.id = wu.user_id
		WHERE wu.workspace_id = $1 AND wu.user_id = $2`,
		scanWorkspaceMember,
		workspaceId.String(),
		userId.String(),
	)
}

// ListByWorkspaceId returns a page of members, owners first
func (r *WorkspaceMemberRepositoryPostgres) ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*WorkspaceMember, error) {
	return QueryMany(
		r.db,
		`SELECT
			wu.workspace_id, u.id, u.name, u.email, wu.role, wu.created_at, wu.updated_at
		FROM workspace_users wu
		INNER JOIN users u ON u.id = wu.user_id
		WHERE wu.workspace_id = $1
		ORDER BY wu.created_at ASC
		LIMIT $2 OFFSET $3`,
		scanWorkspaceMember,
		workspaceId.String(),
		limit,
		offset,
	)
}

// CountByWorkspaceId returns how many members a workspace has
func (r *WorkspaceMemberRepositoryPostgres) CountByWorkspaceId(workspaceId Id) (int64, error) {
	var total int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM workspace_users WHERE workspace_id = $1`,
		workspaceId.String(),
	).Scan(&total)

	return total, err
}

// UpdateRole changes a member's role in a workspace
func (r *WorkspaceMemberRepositoryPostgres) UpdateRole(workspaceId Id, userId Id, role string) error {
	_, err := r.db.Exec(
		`UPDATE workspace_users
		SET role = $1, updated_at = $2
		WHERE workspace_id = $3 AND user_id = $4`,
		role,
		time.Now().UTC(),
		workspaceId.String(),
		userId.String(),
	)

	return err
}

// Delete removes a member from a workspace
func (r *WorkspaceMemberRepositoryPostgres) Delete(workspaceId Id, userId Id) error {
	_, err := r.db.Exec(
		`DELETE FROM workspace_users WHERE workspace_id = $1 AND user_id = $2`,
		workspaceId.String(),
		userId.String(),
	)

	return err
}
