// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

// UserInvite is the DB row for a workspace invite sent to an email address.
type UserInvite struct {
	Id            Id
	Email         string
	Role          string
	Token         string
	Status        string
	InviterUserId Id
	WorkspaceId   Id
	ExpiresAt     time.Time
	AcceptedAt    *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// UserInviteRepository is the interface for workspace invite CRUD.
type UserInviteRepository interface {
	Create(invite *UserInvite) error
	GetById(id Id) (*UserInvite, error)
	ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*UserInvite, error)
	ListPendingByEmail(email string) ([]*UserInvite, error)
	UpdateStatus(id Id, status string, acceptedAt *time.Time) error
	MarkExpiredAsExpired() (int64, error)
	Delete(id Id) error
	CountByWorkspaceId(workspaceId Id) (int64, error)
	CountPendingByEmailInWorkspace(workspaceId Id, email string) (int64, error)
}

// UserInviteRepositoryPostgres implements UserInviteRepository for PostgreSQL
type UserInviteRepositoryPostgres struct {
	db *sql.DB
}

// NewUserInviteRepository returns the workspace invite repository for the current driver
func NewUserInviteRepository(db *sql.DB) UserInviteRepository {
	return &UserInviteRepositoryPostgres{db: db}
}

// scanUserInvite maps a row into a workspace invite. Column order must match the select list.
func scanUserInvite(u *UserInvite) []any {
	return []any{
		&u.Id,
		&u.Email,
		&u.Role,
		&u.Token,
		&u.Status,
		&u.InviterUserId,
		&u.WorkspaceId,
		&u.ExpiresAt,
		&u.AcceptedAt,
		&u.CreatedAt,
		&u.UpdatedAt,
	}
}

// --- Postgres ---

// Create inserts an invite
func (r *UserInviteRepositoryPostgres) Create(invite *UserInvite) error {
	id, err := NewId()
	if err != nil {
		return err
	}

	invite.Id = id

	return r.db.QueryRow(
		`INSERT INTO user_invites
		(id, email, role, token, status, inviter_user_id, workspace_id, expires_at, accepted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at`,
		invite.Id.String(),
		invite.Email,
		invite.Role,
		invite.Token,
		invite.Status,
		invite.InviterUserId.String(),
		invite.WorkspaceId.String(),
		invite.ExpiresAt,
		invite.AcceptedAt,
	).Scan(&invite.CreatedAt, &invite.UpdatedAt)
}

// GetById returns an invite by Id, or nil when it does not exist
func (r *UserInviteRepositoryPostgres) GetById(id Id) (*UserInvite, error) {
	return QueryOne(
		r.db,
		`SELECT
			id, email, role, token, status, inviter_user_id, workspace_id,
			expires_at, accepted_at, created_at, updated_at
		FROM user_invites
		WHERE id = $1`,
		scanUserInvite,
		id.String(),
	)
}

// ListByWorkspaceId returns a page of invites for a workspace, newest first
func (r *UserInviteRepositoryPostgres) ListByWorkspaceId(workspaceId Id, limit, offset int) ([]*UserInvite, error) {
	return QueryMany(
		r.db,
		`SELECT
			id, email, role, token, status, inviter_user_id, workspace_id,
			expires_at, accepted_at, created_at, updated_at
		FROM user_invites
		WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		scanUserInvite,
		workspaceId.String(),
		limit,
		offset,
	)
}

// ListPendingByEmail returns pending, unexpired invites for an email address, oldest first
func (r *UserInviteRepositoryPostgres) ListPendingByEmail(email string) ([]*UserInvite, error) {
	return QueryMany(
		r.db,
		`SELECT
			id, email, role, token, status, inviter_user_id, workspace_id,
			expires_at, accepted_at, created_at, updated_at
		FROM user_invites
		WHERE LOWER(email) = LOWER($1) AND status = 'pending' AND expires_at > $2
		ORDER BY created_at ASC`,
		scanUserInvite,
		email,
		time.Now().UTC(),
	)
}

// UpdateStatus sets the status of an invite, and when it was accepted
func (r *UserInviteRepositoryPostgres) UpdateStatus(id Id, status string, acceptedAt *time.Time) error {
	_, err := r.db.Exec(
		`UPDATE user_invites
		SET
			status = $1,
			accepted_at = $2,
			updated_at = $3
		WHERE id = $4`,
		status,
		acceptedAt,
		time.Now().UTC(),
		id.String(),
	)

	return err
}

// MarkExpiredAsExpired sets pending invites past their expiry to expired. Returns how many changed.
func (r *UserInviteRepositoryPostgres) MarkExpiredAsExpired() (int64, error) {
	result, err := r.db.Exec(
		`UPDATE user_invites
		SET
			status = 'expired',
			updated_at = $1
		WHERE status = 'pending' AND expires_at < $2`,
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Delete removes an invite
func (r *UserInviteRepositoryPostgres) Delete(id Id) error {
	_, err := r.db.Exec(
		`DELETE FROM user_invites WHERE id = $1`,
		id.String(),
	)

	return err
}

// CountByWorkspaceId returns how many invites a workspace has
func (r *UserInviteRepositoryPostgres) CountByWorkspaceId(workspaceId Id) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM user_invites
		WHERE workspace_id = $1`,
		workspaceId.String(),
	).Scan(&count)

	return count, err
}

// CountPendingByEmailInWorkspace returns pending, unexpired invites for an email in a workspace
func (r *UserInviteRepositoryPostgres) CountPendingByEmailInWorkspace(workspaceId Id, email string) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*)
		FROM user_invites
		WHERE workspace_id = $1 AND email = $2 AND status = 'pending' AND expires_at > $3`,
		workspaceId.String(),
		email,
		time.Now().UTC(),
	).Scan(&count)

	return count, err
}
