// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"errors"
	"strings"

	"github.com/clivern/mut/db"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

var (
	ErrPermissionPrincipalMissing = errors.New("permission principal missing")
	ErrPermissionWorkspaceMissing = errors.New("permission workspace missing")
)

// Perm checks workspace permissions for users and workspace access keys.
type Perm struct {
	WorkspaceRepository db.WorkspaceRepository
	user                *db.User
	accessKey           *db.AccessKey
	workspace           *db.Workspace
}

// NewPerm creates a permission checker with the given workspace repository.
func NewPerm(workspaces db.WorkspaceRepository) *Perm {
	return &Perm{
		WorkspaceRepository: workspaces,
	}
}

// WithUser sets the user principal. Mutually exclusive with WithAccessKey.
func (p *Perm) WithUser(user *db.User) *Perm {
	p.user = user
	p.accessKey = nil

	return p
}

// WithAccessKey sets the workspace access key principal. Mutually exclusive with WithUser.
func (p *Perm) WithAccessKey(accessKey *db.AccessKey) *Perm {
	p.accessKey = accessKey
	p.user = nil

	return p
}

// WithWorkspace sets the workspace being accessed.
func (p *Perm) WithWorkspace(workspace *db.Workspace) *Perm {
	p.workspace = workspace

	return p
}

// WithWorkspaceId loads the workspace by id from the repository.
func (p *Perm) WithWorkspaceId(workspaceId db.Id) *Perm {
	workspace, _ := p.WorkspaceRepository.GetById(workspaceId)
	p.workspace = workspace

	return p
}

// Can reports whether the configured principal may perform the given permission on the workspace.
func (p *Perm) Can(permission string) (bool, error) {
	if p.workspace == nil {
		log.Error().
			Str("permission", permission).
			Msg("Permission check without a workspace")
		return false, ErrPermissionWorkspaceMissing
	}

	switch {
	case p.user != nil:
		return p.CanAsUser(permission)
	case p.accessKey != nil:
		return p.CanAsAccessKey(permission)
	default:
		log.Error().
			Str("workspaceId", p.workspace.Id.String()).
			Str("permission", permission).
			Msg("Permission check without a principal")
		return false, ErrPermissionPrincipalMissing
	}
}

// CanAsUser checks if the user has the given permission.
func (p *Perm) CanAsUser(permission string) (bool, error) {
	membership, err := p.WorkspaceRepository.GetForUser(p.workspace.Id, p.user.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", p.workspace.Id.String()).
			Str("userId", p.user.Id.String()).
			Msg("Failed to get workspace membership")
		return false, err
	}
	if membership == nil {
		return false, nil
	}

	switch permission {
	case AccessKeyPermissionWorkspaceRead, AccessKeyPermissionMembersRead, AccessKeyPermissionClustersRead:
		return true, nil
	case PermissionMembersManage, PermissionAccessKeysManage, PermissionClustersManage:
		return membership.Role == db.WorkspaceRoleOwner || membership.Role == db.WorkspaceRoleAdmin, nil
	default:
		return false, nil
	}
}

// CanAsAccessKey checks if the access key has the given permission.
func (p *Perm) CanAsAccessKey(permission string) (bool, error) {
	if p.accessKey.WorkspaceId != p.workspace.Id {
		return false, nil
	}

	return lo.Contains(strings.Split(p.accessKey.Permissions, ","), permission), nil
}
