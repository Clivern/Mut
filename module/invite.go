// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/clivern/mut/conf"
	"github.com/clivern/mut/db"
	"github.com/clivern/mut/pkg/util"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/spf13/viper"
)

const (
	InviteStatusPending  = "pending"
	InviteStatusAccepted = "accepted"
	InviteStatusExpired  = "expired"
)

var (
	ErrInviteNotFound         = errors.New("invite not found")
	ErrPendingInviteExists    = errors.New("pending invite already exists")
	ErrUserAlreadyInWorkspace = errors.New("user already in workspace")
	ErrFailedCreateInvite     = errors.New("failed create invite")
	ErrFailedListInvites      = errors.New("failed list invites")
	ErrFailedGetInvite        = errors.New("failed get invite")
	ErrFailedDeleteInvite     = errors.New("failed delete invite")
	ErrFailedAttachInvites    = errors.New("failed attach invites")
)

// InviteMailer sends invite emails.
type InviteMailer interface {
	SendInviteEmail(to, inviteLink, platformName string) error
}

// CreateInviteRequest is the body for inviting an email address to a workspace.
type CreateInviteRequest struct {
	Email string `json:"email" validate:"required,email,min=4,max=255" label:"Email"`
	Role  string `json:"role" validate:"required,oneof=admin regular readonly" label:"Role"`
}

// InviteResponse is an invite shaped for API responses.
type InviteResponse struct {
	Id            db.Id   `json:"id"`
	Email         string  `json:"email"`
	Role          string  `json:"role"`
	Token         string  `json:"token,omitempty"`
	Status        string  `json:"status"`
	InviterUserId db.Id   `json:"inviterUserId"`
	WorkspaceId   db.Id   `json:"workspaceId"`
	ExpiresAt     string  `json:"expiresAt"`
	AcceptedAt    *string `json:"acceptedAt,omitempty"`
	CreatedAt     string  `json:"createdAt"`
	UpdatedAt     string  `json:"updatedAt"`
}

// ListInvitesResult is a page of invites for a workspace.
type ListInvitesResult struct {
	Invites []*InviteResponse
	Total   int64
}

// Invite manages workspace invites sent to email addresses.
type Invite struct {
	UserInviteRepository      db.UserInviteRepository
	UserRepository            db.UserRepository
	WorkspaceRepository       db.WorkspaceRepository
	WorkspaceMemberRepository db.WorkspaceMemberRepository
	Mailer                    InviteMailer
}

// NewInvite creates an invite module with the given repositories and mailer.
func NewInvite(
	invites db.UserInviteRepository,
	users db.UserRepository,
	workspaces db.WorkspaceRepository,
	members db.WorkspaceMemberRepository,
	mailer InviteMailer,
) *Invite {
	return &Invite{
		UserInviteRepository:      invites,
		UserRepository:            users,
		WorkspaceRepository:       workspaces,
		WorkspaceMemberRepository: members,
		Mailer:                    mailer,
	}
}

// toInviteResponse shapes an invite row for API responses. An empty token hides it.
func toInviteResponse(inv *db.UserInvite, token string) *InviteResponse {
	return &InviteResponse{
		Id:            inv.Id,
		Email:         inv.Email,
		Role:          inv.Role,
		Token:         token,
		Status:        inv.Status,
		InviterUserId: inv.InviterUserId,
		WorkspaceId:   inv.WorkspaceId,
		ExpiresAt:     inv.ExpiresAt.UTC().Format(time.RFC3339),
		CreatedAt:     inv.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:     inv.UpdatedAt.UTC().Format(time.RFC3339),
		AcceptedAt: lo.TernaryF(
			inv.AcceptedAt != nil,
			func() *string { return new(inv.AcceptedAt.UTC().Format(time.RFC3339)) },
			func() *string { return nil },
		),
	}
}

// CreateInvite invites an email address to a workspace and sends the sign-in email.
func (i *Invite) CreateInvite(workspaceId db.Id, user *db.User, req *CreateInviteRequest) (*InviteResponse, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	existingUser, err := i.UserRepository.GetByEmail(req.Email)
	if err != nil {
		log.Error().
			Err(err).
			Str("email", req.Email).
			Msg("Failed to get user by email")
		return nil, fmt.Errorf("%w: %v", ErrFailedCreateInvite, err)
	}

	if existingUser != nil {
		isMember, err := i.isWorkspaceMember(workspaceId, existingUser.Id)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrFailedCreateInvite, err)
		}
		if isMember {
			return nil, ErrUserAlreadyInWorkspace
		}
	}

	count, err := i.UserInviteRepository.CountPendingByEmailInWorkspace(workspaceId, req.Email)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("email", req.Email).
			Msg("Failed to count pending invites")
		return nil, fmt.Errorf("%w: %v", ErrFailedCreateInvite, err)
	}
	if count > 0 {
		return nil, ErrPendingInviteExists
	}

	token, err := uuid.NewRandom()
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to generate invite token")
		return nil, fmt.Errorf("%w: %v", ErrFailedCreateInvite, err)
	}

	invite := &db.UserInvite{
		Email:         req.Email,
		Role:          req.Role,
		Token:         token.String(),
		Status:        InviteStatusPending,
		InviterUserId: user.Id,
		WorkspaceId:   workspaceId,
		ExpiresAt:     time.Now().UTC().Add(conf.InviteExpiry),
	}

	err = i.UserInviteRepository.Create(invite)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("email", req.Email).
			Msg("Failed to create invite")
		return nil, fmt.Errorf("%w: %v", ErrFailedCreateInvite, err)
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("inviteId", invite.Id.String()).
		Str("userId", user.Id.String()).
		Str("email", req.Email).
		Str("role", req.Role).
		Msg("Workspace invite created")

	err = i.Mailer.SendInviteEmail(req.Email, util.AppURL("/login"), viper.GetString("app.name"))
	if err != nil {
		log.Warn().
			Err(err).
			Str("userId", user.Id.String()).
			Str("email", req.Email).
			Msg("Failed to send invite email; invite was created")
	}

	if existingUser != nil {
		err = i.AttachPending(existingUser)
		if err != nil {
			log.Error().
				Err(err).
				Str("userId", existingUser.Id.String()).
				Str("email", req.Email).
				Msg("Failed to attach workspace invite")
		} else {
			invite.Status = InviteStatusAccepted
		}
	}

	return toInviteResponse(invite, invite.Token), nil
}

// ListInvites returns a page of a workspace's invites. Only owners and admins may list them.
func (i *Invite) ListInvites(workspaceId db.Id, user *db.User, limit, offset int) (*ListInvitesResult, error) {
	i.markExpired()

	total, err := i.UserInviteRepository.CountByWorkspaceId(workspaceId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to count invites")
		return nil, fmt.Errorf("%w: %v", ErrFailedListInvites, err)
	}

	invites, err := i.UserInviteRepository.ListByWorkspaceId(workspaceId, limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to list invites")
		return nil, fmt.Errorf("%w: %v", ErrFailedListInvites, err)
	}

	list := make([]*InviteResponse, 0, len(invites))
	for _, inv := range invites {
		list = append(list, toInviteResponse(inv, inv.Token))
	}

	return &ListInvitesResult{Invites: list, Total: total}, nil
}

// GetInvite returns one invite of a workspace. Only owners and admins may read it.
func (i *Invite) GetInvite(workspaceId db.Id, user *db.User, inviteId db.Id) (*InviteResponse, error) {
	i.markExpired()

	invite, err := i.UserInviteRepository.GetById(inviteId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("inviteId", inviteId.String()).
			Msg("Failed to get invite")
		return nil, fmt.Errorf("%w: %v", ErrFailedGetInvite, err)
	}
	if invite == nil || invite.WorkspaceId != workspaceId {
		return nil, ErrInviteNotFound
	}

	return toInviteResponse(invite, ""), nil
}

// DeleteInvite revokes an invite of a workspace. Only owners and admins may revoke it.
func (i *Invite) DeleteInvite(workspaceId db.Id, user *db.User, inviteId db.Id) error {
	invite, err := i.UserInviteRepository.GetById(inviteId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("inviteId", inviteId.String()).
			Msg("Failed to get invite")
		return fmt.Errorf("%w: %v", ErrFailedDeleteInvite, err)
	}
	if invite == nil || invite.WorkspaceId != workspaceId {
		return ErrInviteNotFound
	}

	err = i.UserInviteRepository.Delete(inviteId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("inviteId", inviteId.String()).
			Msg("Failed to delete invite")
		return fmt.Errorf("%w: %v", ErrFailedDeleteInvite, err)
	}

	log.Info().
		Str("userId", user.Id.String()).
		Str("workspaceId", workspaceId.String()).
		Str("inviteId", inviteId.String()).
		Msg("Workspace invite deleted")

	return nil
}

// AttachPending adds the user to every pending workspace invite sent to their email.
func (i *Invite) AttachPending(user *db.User) error {
	i.markExpired()

	invites, err := i.UserInviteRepository.ListPendingByEmail(user.Email)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Msg("Failed to list pending invites")
		return fmt.Errorf("%w: %v", ErrFailedAttachInvites, err)
	}

	now := time.Now().UTC()
	for _, invite := range invites {
		isMember, err := i.isWorkspaceMember(invite.WorkspaceId, user.Id)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrFailedAttachInvites, err)
		}

		if !isMember {
			err = i.WorkspaceMemberRepository.Create(invite.WorkspaceId, user.Id, invite.Role)
			if err != nil {
				log.Error().
					Err(err).
					Str("userId", user.Id.String()).
					Str("workspaceId", invite.WorkspaceId.String()).
					Msg("Failed to add workspace member")
				return fmt.Errorf("%w: %v", ErrFailedAttachInvites, err)
			}
		}

		err = i.UserInviteRepository.UpdateStatus(invite.Id, InviteStatusAccepted, &now)
		if err != nil {
			log.Error().
				Err(err).
				Str("userId", user.Id.String()).
				Str("inviteId", invite.Id.String()).
				Msg("Failed to accept invite")
			return fmt.Errorf("%w: %v", ErrFailedAttachInvites, err)
		}

		log.Info().
			Str("inviteId", invite.Id.String()).
			Str("workspaceId", invite.WorkspaceId.String()).
			Str("userId", user.Id.String()).
			Msg("Workspace invite attached")
	}

	return nil
}

// isWorkspaceMember reports whether the user already has a membership in the workspace.
func (i *Invite) isWorkspaceMember(workspaceId db.Id, userId db.Id) (bool, error) {
	membership, err := i.WorkspaceMemberRepository.GetByWorkspaceAndUser(workspaceId, userId)
	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("userId", userId.String()).
			Msg("Failed to get workspace membership")
		return false, err
	}

	return membership != nil, nil
}

// markExpired flags pending invites past their expiry. A failure is logged and does not stop the caller.
func (i *Invite) markExpired() {
	count, err := i.UserInviteRepository.MarkExpiredAsExpired()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to mark expired invites")
		return
	}

	if count > 0 {
		log.Info().
			Int64("count", count).
			Msg("Expired invites marked")
	}
}
