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

// ListWorkspaceMembersAction returns a page of a workspace's members.
func (a *API) ListWorkspaceMembersAction(w http.ResponseWriter, r *http.Request) {
	workspaceId := db.Id(chi.URLParam(r, "workspaceId"))

	limit, offset := a.ParsePagination(r)

	user, _ := a.GetUser(r)

	result, err := a.WorkspaceMembers.ListMembers(workspaceId, user, limit, offset)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrWorkspaceNotFound, http.StatusNotFound, locale.TR(r, "workspace_not_found")},
			{module.ErrWorkspaceForbidden, http.StatusForbidden, locale.TR(r, "workspace_forbidden")},
			{module.ErrWorkspaceMemberNotFound, http.StatusNotFound, locale.TR(r, "workspace_member_not_found")},
			{module.ErrWorkspaceOwnerProtected, http.StatusForbidden, locale.TR(r, "workspace_owner_protected")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to list workspace members")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_list_workspace_members"),
		})
		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"members": result.Members,
		"_meta": map[string]any{
			"limit":  limit,
			"offset": offset,
			"total":  result.Total,
		},
	})
}

// UpdateWorkspaceMemberAction changes a member's role. Owners and admins may do this.
func (a *API) UpdateWorkspaceMemberAction(w http.ResponseWriter, r *http.Request) {
	workspaceId := db.Id(chi.URLParam(r, "workspaceId"))
	memberUserId := db.Id(chi.URLParam(r, "memberUserId"))

	var req module.UpdateWorkspaceMemberRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	user, _ := a.GetUser(r)

	member, err := a.WorkspaceMembers.UpdateMemberRole(workspaceId, user, memberUserId, &req)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrWorkspaceNotFound, http.StatusNotFound, locale.TR(r, "workspace_not_found")},
			{module.ErrWorkspaceForbidden, http.StatusForbidden, locale.TR(r, "workspace_forbidden")},
			{module.ErrWorkspaceMemberNotFound, http.StatusNotFound, locale.TR(r, "workspace_member_not_found")},
			{module.ErrWorkspaceOwnerProtected, http.StatusForbidden, locale.TR(r, "workspace_owner_protected")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to update workspace member role")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_update_workspace_member"),
		})
		return
	}

	a.WriteJSON(w, http.StatusOK, member)
}

// RemoveWorkspaceMemberAction removes a member from a workspace. Owners and admins may do this.
func (a *API) RemoveWorkspaceMemberAction(w http.ResponseWriter, r *http.Request) {
	workspaceId := db.Id(chi.URLParam(r, "workspaceId"))
	memberUserId := db.Id(chi.URLParam(r, "memberUserId"))

	user, _ := a.GetUser(r)

	err := a.WorkspaceMembers.RemoveMember(workspaceId, user, memberUserId)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrWorkspaceNotFound, http.StatusNotFound, locale.TR(r, "workspace_not_found")},
			{module.ErrWorkspaceForbidden, http.StatusForbidden, locale.TR(r, "workspace_forbidden")},
			{module.ErrWorkspaceMemberNotFound, http.StatusNotFound, locale.TR(r, "workspace_member_not_found")},
			{module.ErrWorkspaceOwnerProtected, http.StatusForbidden, locale.TR(r, "workspace_owner_protected")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to remove workspace member")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_remove_workspace_member"),
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
