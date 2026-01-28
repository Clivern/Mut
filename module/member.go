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
	ErrWorkspaceMemberNotFound = errors.New("workspace member not found")
	ErrWorkspaceOwnerProtected = errors.New("workspace owner cannot be changed or removed")
)

// UpdateWorkspaceMemberRequest is the body for changing a member's role.
type UpdateWorkspaceMemberRequest struct {
	Role string `json:"role" validate:"required,oneof=admin regular readonly" label:"Role"`
}

// WorkspaceMemberResponse is a workspace member shaped for API responses.
type WorkspaceMemberResponse struct {
	UserId    db.Id  `json:"userId"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// ListWorkspaceMembersResult is a page of workspace members.
type ListWorkspaceMembersResult struct {
	Members []*WorkspaceMemberResponse
	Total   int64
}

// WorkspaceMembers manages who belongs to a workspace and with which role.
type WorkspaceMembers struct {
	WorkspaceMemberRepository db.WorkspaceMemberRepository
	WorkspaceRepository       db.WorkspaceRepository
	UserRepository            db.UserRepository
}

// NewWorkspaceMembers creates a workspace membership module with the given repositories.
func NewWorkspaceMembers(
	members db.WorkspaceMemberRepository,
	workspaces db.WorkspaceRepository,
	users db.UserRepository,
) *WorkspaceMembers {
	return &WorkspaceMembers{
		WorkspaceMemberRepository: members,
		WorkspaceRepository:       workspaces,
		UserRepository:            users,
	}
}

// toWorkspaceMemberResponse shapes a membership row for API responses.
func toWorkspaceMemberResponse(item *db.WorkspaceMember) *WorkspaceMemberResponse {
	return &WorkspaceMemberResponse{
		UserId:    item.UserId,
		Name:      item.Name,
		Email:     item.Email,
		Role:      item.Role,
		CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// ListMembers returns a page of a workspace's members.
func (m *WorkspaceMembers) ListMembers(workspaceId db.Id, user *db.User, limit, offset int) (*ListWorkspaceMembersResult, error) {
	if user != nil {
		membership, err := m.WorkspaceRepository.GetForUser(workspaceId, user.Id)
		if err != nil {
			log.Error().
				Err(err).
				Str("userId", user.Id.String()).
				Str("workspaceId", workspaceId.String()).
				Msg("Failed to get workspace membership")
			return nil, err
		}
		if membership == nil {
			return nil, ErrWorkspaceNotFound
		}
	}

	total, err := m.WorkspaceMemberRepository.CountByWorkspaceId(workspaceId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to count workspace members")
		return nil, err
	}

	members, err := m.WorkspaceMemberRepository.ListByWorkspaceId(workspaceId, limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to list workspace members")
		return nil, err
	}

	list := make([]*WorkspaceMemberResponse, 0, len(members))
	for _, item := range members {
		list = append(list, toWorkspaceMemberResponse(item))
	}

	return &ListWorkspaceMembersResult{Members: list, Total: total}, nil
}

// UpdateMemberRole changes the role of a member. The owner's role cannot be changed.
func (m *WorkspaceMembers) UpdateMemberRole(workspaceId db.Id, user *db.User, memberUserId db.Id, req *UpdateWorkspaceMemberRequest) (*WorkspaceMemberResponse, error) {
	member, err := m.WorkspaceMemberRepository.GetByWorkspaceAndUser(workspaceId, memberUserId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to get workspace member")
		return nil, err
	}
	if member == nil {
		return nil, ErrWorkspaceMemberNotFound
	}
	if member.Role == db.WorkspaceRoleOwner {
		return nil, ErrWorkspaceOwnerProtected
	}

	err = m.WorkspaceMemberRepository.UpdateRole(workspaceId, memberUserId, req.Role)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("memberUserId", memberUserId.String()).
			Msg("Failed to update workspace member role")
		return nil, fmt.Errorf("update workspace member role: %w", err)
	}

	log.Info().
		Str("userId", user.Id.String()).
		Str("workspaceId", workspaceId.String()).
		Str("memberUserId", memberUserId.String()).
		Str("role", req.Role).
		Msg("Workspace member role updated")

	updated, err := m.WorkspaceMemberRepository.GetByWorkspaceAndUser(workspaceId, memberUserId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to get updated workspace member")
		return nil, err
	}

	return toWorkspaceMemberResponse(updated), nil
}

// RemoveMember removes a member from the workspace. The owner cannot be removed.
func (m *WorkspaceMembers) RemoveMember(workspaceId db.Id, user *db.User, memberUserId db.Id) error {
	member, err := m.WorkspaceMemberRepository.GetByWorkspaceAndUser(workspaceId, memberUserId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to get workspace member")
		return err
	}
	if member == nil {
		return ErrWorkspaceMemberNotFound
	}
	if member.Role == db.WorkspaceRoleOwner {
		return ErrWorkspaceOwnerProtected
	}

	err = m.WorkspaceMemberRepository.Delete(workspaceId, memberUserId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("memberUserId", memberUserId.String()).
			Msg("Failed to remove workspace member")
		return fmt.Errorf("remove workspace member: %w", err)
	}

	log.Info().
		Str("userId", user.Id.String()).
		Str("workspaceId", workspaceId.String()).
		Str("memberUserId", memberUserId.String()).
		Msg("Workspace member removed")

	return nil
}
