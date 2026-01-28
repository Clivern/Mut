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
)

var (
	ErrWorkspaceNotFound  = errors.New("workspace not found")
	ErrWorkspaceForbidden = errors.New("workspace action forbidden")
)

// WorkspaceRequest is the body for creating or renaming a workspace.
type WorkspaceRequest struct {
	Name string `json:"name" validate:"required,min=2,max=60" label:"Workspace Name"`
}

// WorkspaceList is a page of workspaces for a user.
type WorkspaceList struct {
	Workspaces []*db.Workspace
	Total      int64
}

// Workspace manages workspaces and checks membership before each action.
type Workspace struct {
	WorkspaceRepository db.WorkspaceRepository
}

// NewWorkspace creates a Workspace module with the given repository.
func NewWorkspace(workspaces db.WorkspaceRepository) *Workspace {
	return &Workspace{
		WorkspaceRepository: workspaces,
	}
}

// CreateWorkspace creates a workspace and makes the user its owner.
func (w *Workspace) CreateWorkspace(userId db.Id, req *WorkspaceRequest) (*db.Workspace, error) {
	now := time.Now().UTC()
	workspace := &db.Workspace{
		Name:      req.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := w.WorkspaceRepository.CreateWithOwner(workspace, userId)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", userId.String()).
			Msg("Failed to create workspace")
		return nil, fmt.Errorf("create workspace: %w", err)
	}

	log.Info().
		Str("userId", userId.String()).
		Str("workspaceId", workspace.Id.String()).
		Msg("Workspace created")

	workspace.Role = db.WorkspaceRoleOwner
	workspace.MembersCount = 1

	return workspace, nil
}

// ListWorkspaces returns a page of the workspaces the user belongs to.
func (w *Workspace) ListWorkspaces(userId db.Id, limit, offset int) (*WorkspaceList, error) {
	workspaces, err := w.WorkspaceRepository.ListByUser(userId, limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", userId.String()).
			Msg("Failed to list workspaces")
		return nil, fmt.Errorf("list workspaces: %w", err)
	}

	total, err := w.WorkspaceRepository.CountByUser(userId)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", userId.String()).
			Msg("Failed to count workspaces")
		return nil, fmt.Errorf("count workspaces: %w", err)
	}

	return &WorkspaceList{
		Workspaces: workspaces,
		Total:      total,
	}, nil
}

// GetWorkspace returns a workspace the user is a member of.
func (w *Workspace) GetWorkspace(userId db.Id, workspaceId db.Id) (*db.Workspace, error) {
	workspace, err := w.WorkspaceRepository.GetForUser(workspaceId, userId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to get workspace")
		return nil, fmt.Errorf("get workspace: %w", err)
	}

	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	return workspace, nil
}

// GetWorkspaceForAccessKey returns a workspace for a request authorized by a workspace access key.
func (w *Workspace) GetWorkspaceForAccessKey(workspaceId db.Id) (*db.Workspace, error) {
	workspace, err := w.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to get workspace")
		return nil, fmt.Errorf("get workspace: %w", err)
	}

	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	return workspace, nil
}

// UpdateWorkspace renames a workspace. Only owners and admins of the workspace may rename it.
func (w *Workspace) UpdateWorkspace(userId db.Id, workspaceId db.Id, req *WorkspaceRequest) (*db.Workspace, error) {
	workspace, err := w.GetWorkspace(userId, workspaceId)
	if err != nil {
		return nil, err
	}

	if !(workspace.Role == db.WorkspaceRoleOwner || workspace.Role == db.WorkspaceRoleAdmin) {
		return nil, ErrWorkspaceForbidden
	}

	workspace.Name = req.Name
	workspace.UpdatedAt = time.Now().UTC()

	err = w.WorkspaceRepository.Update(workspace)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to update workspace")
		return nil, fmt.Errorf("update workspace: %w", err)
	}

	return workspace, nil
}

// DeleteWorkspace removes a workspace. Only owners and admins of the workspace may delete it.
func (w *Workspace) DeleteWorkspace(userId db.Id, workspaceId db.Id) error {
	workspace, err := w.GetWorkspace(userId, workspaceId)
	if err != nil {
		return err
	}

	if !(workspace.Role == db.WorkspaceRoleOwner || workspace.Role == db.WorkspaceRoleAdmin) {
		return ErrWorkspaceForbidden
	}

	err = w.WorkspaceRepository.Delete(workspace.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to delete workspace")
		return fmt.Errorf("delete workspace: %w", err)
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Msg("Workspace deleted")

	return nil
}
