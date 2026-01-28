// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/clivern/mut/db"
	"github.com/clivern/mut/locale"
	"github.com/clivern/mut/module"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// CreateInviteAction invites an email address to a workspace. Owners and admins may do this.
func (a *API) CreateInviteAction(w http.ResponseWriter, r *http.Request) {
	workspaceId := db.Id(chi.URLParam(r, "workspaceId"))

	var req module.CreateInviteRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	user, _ := a.GetUser(r)

	invite, err := a.Invite.CreateInvite(workspaceId, user, &req)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrWorkspaceNotFound, http.StatusNotFound, locale.TR(r, "workspace_not_found")},
			{module.ErrWorkspaceForbidden, http.StatusForbidden, locale.TR(r, "workspace_forbidden")},
			{module.ErrInviteNotFound, http.StatusNotFound, locale.TR(r, "invite_not_found")},
			{module.ErrUserAlreadyInWorkspace, http.StatusConflict, locale.TR(r, "user_already_workspace_member")},
			{module.ErrPendingInviteExists, http.StatusConflict, locale.TR(r, "pending_invite_already_exists")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to create invite")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_create_invite"),
		})
		return
	}

	a.WriteJSON(w, http.StatusCreated, invite)
}

// ListInvitesAction returns a page of a workspace's invites. Owners and admins may do this.
func (a *API) ListInvitesAction(w http.ResponseWriter, r *http.Request) {
	workspaceId := db.Id(chi.URLParam(r, "workspaceId"))

	limit, offset := a.ParsePagination(r)

	user, _ := a.GetUser(r)

	result, err := a.Invite.ListInvites(workspaceId, user, limit, offset)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrWorkspaceNotFound, http.StatusNotFound, locale.TR(r, "workspace_not_found")},
			{module.ErrWorkspaceForbidden, http.StatusForbidden, locale.TR(r, "workspace_forbidden")},
			{module.ErrInviteNotFound, http.StatusNotFound, locale.TR(r, "invite_not_found")},
			{module.ErrUserAlreadyInWorkspace, http.StatusConflict, locale.TR(r, "user_already_workspace_member")},
			{module.ErrPendingInviteExists, http.StatusConflict, locale.TR(r, "pending_invite_already_exists")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to list invites")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_list_invites"),
		})
		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"invites": result.Invites,
		"_meta": map[string]any{
			"limit":  limit,
			"offset": offset,
			"total":  result.Total,
		},
	})
}

// GetInviteAction returns one of a workspace's invites. Owners and admins may do this.
func (a *API) GetInviteAction(w http.ResponseWriter, r *http.Request) {
	workspaceId := db.Id(chi.URLParam(r, "workspaceId"))
	inviteId := db.Id(chi.URLParam(r, "inviteId"))

	user, _ := a.GetUser(r)

	invite, err := a.Invite.GetInvite(workspaceId, user, inviteId)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrWorkspaceNotFound, http.StatusNotFound, locale.TR(r, "workspace_not_found")},
			{module.ErrWorkspaceForbidden, http.StatusForbidden, locale.TR(r, "workspace_forbidden")},
			{module.ErrInviteNotFound, http.StatusNotFound, locale.TR(r, "invite_not_found")},
			{module.ErrUserAlreadyInWorkspace, http.StatusConflict, locale.TR(r, "user_already_workspace_member")},
			{module.ErrPendingInviteExists, http.StatusConflict, locale.TR(r, "pending_invite_already_exists")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("inviteId", inviteId.String()).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to get invite")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_get_invite"),
		})
		return
	}

	a.WriteJSON(w, http.StatusOK, invite)
}

// DeleteInviteAction revokes one of a workspace's invites. Owners and admins may do this.
func (a *API) DeleteInviteAction(w http.ResponseWriter, r *http.Request) {
	workspaceId := db.Id(chi.URLParam(r, "workspaceId"))
	inviteId := db.Id(chi.URLParam(r, "inviteId"))

	user, _ := a.GetUser(r)

	err := a.Invite.DeleteInvite(workspaceId, user, inviteId)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrWorkspaceNotFound, http.StatusNotFound, locale.TR(r, "workspace_not_found")},
			{module.ErrWorkspaceForbidden, http.StatusForbidden, locale.TR(r, "workspace_forbidden")},
			{module.ErrInviteNotFound, http.StatusNotFound, locale.TR(r, "invite_not_found")},
			{module.ErrUserAlreadyInWorkspace, http.StatusConflict, locale.TR(r, "user_already_workspace_member")},
			{module.ErrPendingInviteExists, http.StatusConflict, locale.TR(r, "pending_invite_already_exists")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("inviteId", inviteId.String()).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to delete invite")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_delete_invite"),
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
