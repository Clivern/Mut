// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"net/http"
	"time"

	"github.com/clivern/mut/db"
	"github.com/clivern/mut/locale"
	"github.com/clivern/mut/module"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// CreateWorkspaceAction creates a workspace owned by the authenticated user.
func (a *API) CreateWorkspaceAction(w http.ResponseWriter, r *http.Request) {
	var req module.WorkspaceRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	user, _ := a.GetUser(r)

	workspace, err := a.Workspace.CreateWorkspace(user.Id, &req)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Msg("Failed to create workspace")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_create_workspace"),
		})
		return
	}

	a.WriteJSON(w, http.StatusCreated, workspaceResponse(workspace))
}

// ListWorkspacesAction returns a page of the workspaces the authenticated user belongs to.
func (a *API) ListWorkspacesAction(w http.ResponseWriter, r *http.Request) {
	limit, offset := a.ParsePagination(r)

	user, _ := a.GetUser(r)

	result, err := a.Workspace.ListWorkspaces(user.Id, limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Msg("Failed to list workspaces")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_list_workspaces"),
		})
		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"workspaces": lo.Map(result.Workspaces, func(workspace *db.Workspace, _ int) map[string]any {
			return workspaceResponse(workspace)
		}),
		"_meta": map[string]any{
			"limit":  limit,
			"offset": offset,
			"total":  result.Total,
		},
	})
}

// GetWorkspaceAction returns a workspace the authenticated user belongs to.
func (a *API) GetWorkspaceAction(w http.ResponseWriter, r *http.Request) {
	workspaceId := db.Id(chi.URLParam(r, "workspaceId"))

	var workspace *db.Workspace
	var err error

	user, isUser := a.GetUser(r)
	if isUser {
		workspace, err = a.Workspace.GetWorkspace(user.Id, workspaceId)
	} else {
		workspace, err = a.Workspace.GetWorkspaceForAccessKey(workspaceId)
	}
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrWorkspaceNotFound, http.StatusNotFound, locale.TR(r, "workspace_not_found")},
			{module.ErrWorkspaceForbidden, http.StatusForbidden, locale.TR(r, "workspace_forbidden")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to get workspace")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_get_workspace"),
		})
		return
	}

	a.WriteJSON(w, http.StatusOK, workspaceResponse(workspace))
}

// UpdateWorkspaceAction renames a workspace. Owners and admins of the workspace may do this.
func (a *API) UpdateWorkspaceAction(w http.ResponseWriter, r *http.Request) {
	workspaceId := db.Id(chi.URLParam(r, "workspaceId"))

	var req module.WorkspaceRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	user, _ := a.GetUser(r)

	workspace, err := a.Workspace.UpdateWorkspace(user.Id, workspaceId, &req)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrWorkspaceNotFound, http.StatusNotFound, locale.TR(r, "workspace_not_found")},
			{module.ErrWorkspaceForbidden, http.StatusForbidden, locale.TR(r, "workspace_forbidden")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to update workspace")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_update_workspace"),
		})
		return
	}

	a.WriteJSON(w, http.StatusOK, workspaceResponse(workspace))
}

// DeleteWorkspaceAction removes a workspace. Owners and admins of the workspace may do this.
func (a *API) DeleteWorkspaceAction(w http.ResponseWriter, r *http.Request) {
	workspaceId := db.Id(chi.URLParam(r, "workspaceId"))

	user, _ := a.GetUser(r)

	err := a.Workspace.DeleteWorkspace(user.Id, workspaceId)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrWorkspaceNotFound, http.StatusNotFound, locale.TR(r, "workspace_not_found")},
			{module.ErrWorkspaceForbidden, http.StatusForbidden, locale.TR(r, "workspace_forbidden")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to delete workspace")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_delete_workspace"),
		})
		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"successMessage": locale.TR(r, "workspace_deleted_successfully"),
	})
}

// workspaceResponse is the JSON shape of a workspace for the calling member.
func workspaceResponse(workspace *db.Workspace) map[string]any {
	return map[string]any{
		"id":           workspace.Id,
		"name":         workspace.Name,
		"role":         workspace.Role,
		"status":       "active",
		"membersCount": workspace.MembersCount,
		"createdAt":    workspace.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":    workspace.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
