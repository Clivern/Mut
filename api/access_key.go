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

// CreateWorkspaceAccessKeyAction creates a workspace access key; the raw key is returned only once.
func (a *API) CreateWorkspaceAccessKeyAction(w http.ResponseWriter, r *http.Request) {
	workspaceId := db.Id(chi.URLParam(r, "workspaceId"))

	var req module.CreateAccessKeyRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	user, _ := a.GetUser(r)

	log.Info().
		Str("userId", user.Id.String()).
		Str("workspaceId", workspaceId.String()).
		Msg("New access key request")

	accessKey, err := a.AccessKey.CreateAccessKey(workspaceId, user, &req)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrWorkspaceNotFound, http.StatusNotFound, locale.TR(r, "workspace_not_found")},
			{module.ErrWorkspaceForbidden, http.StatusForbidden, locale.TR(r, "workspace_forbidden")},
			{module.ErrAccessKeyNotFound, http.StatusNotFound, locale.TR(r, "access_key_not_found")},
			{module.ErrInvalidExpiresAt, http.StatusBadRequest, locale.TR(r, "invalid_expires_at_format")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to create access key")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_create_access_key"),
		})
		return
	}

	a.WriteJSON(w, http.StatusCreated, accessKey)
}

// ListWorkspaceAccessKeysAction lists a workspace's access keys (masked, never the secret).
func (a *API) ListWorkspaceAccessKeysAction(w http.ResponseWriter, r *http.Request) {
	workspaceId := db.Id(chi.URLParam(r, "workspaceId"))

	user, _ := a.GetUser(r)

	limit, offset := a.ParsePagination(r)

	result, err := a.AccessKey.ListAccessKeys(workspaceId, user, limit, offset)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrWorkspaceNotFound, http.StatusNotFound, locale.TR(r, "workspace_not_found")},
			{module.ErrWorkspaceForbidden, http.StatusForbidden, locale.TR(r, "workspace_forbidden")},
			{module.ErrAccessKeyNotFound, http.StatusNotFound, locale.TR(r, "access_key_not_found")},
			{module.ErrInvalidExpiresAt, http.StatusBadRequest, locale.TR(r, "invalid_expires_at_format")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Msg("Failed to list access keys")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_list_access_keys"),
		})
		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"accessKeys": result.AccessKeys,
		"_meta": map[string]any{
			"limit":  limit,
			"offset": offset,
			"total":  result.Total,
		},
	})
}

// GetWorkspaceAccessKeyAction returns one workspace access key's metadata (never the secret).
func (a *API) GetWorkspaceAccessKeyAction(w http.ResponseWriter, r *http.Request) {
	workspaceId := db.Id(chi.URLParam(r, "workspaceId"))
	accessKeyId := db.Id(chi.URLParam(r, "keyId"))

	user, _ := a.GetUser(r)

	accessKey, err := a.AccessKey.GetAccessKey(workspaceId, user, accessKeyId)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrWorkspaceNotFound, http.StatusNotFound, locale.TR(r, "workspace_not_found")},
			{module.ErrWorkspaceForbidden, http.StatusForbidden, locale.TR(r, "workspace_forbidden")},
			{module.ErrAccessKeyNotFound, http.StatusNotFound, locale.TR(r, "access_key_not_found")},
			{module.ErrInvalidExpiresAt, http.StatusBadRequest, locale.TR(r, "invalid_expires_at_format")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("accessKeyId", accessKeyId.String()).
			Msg("Failed to get access key")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_get_access_key"),
		})
		return
	}

	a.WriteJSON(w, http.StatusOK, accessKey)
}

// DeleteWorkspaceAccessKeyAction revokes a workspace access key.
func (a *API) DeleteWorkspaceAccessKeyAction(w http.ResponseWriter, r *http.Request) {
	workspaceId := db.Id(chi.URLParam(r, "workspaceId"))
	accessKeyId := db.Id(chi.URLParam(r, "keyId"))

	user, _ := a.GetUser(r)

	err := a.AccessKey.DeleteAccessKey(workspaceId, user, accessKeyId)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrWorkspaceNotFound, http.StatusNotFound, locale.TR(r, "workspace_not_found")},
			{module.ErrWorkspaceForbidden, http.StatusForbidden, locale.TR(r, "workspace_forbidden")},
			{module.ErrAccessKeyNotFound, http.StatusNotFound, locale.TR(r, "access_key_not_found")},
			{module.ErrInvalidExpiresAt, http.StatusBadRequest, locale.TR(r, "invalid_expires_at_format")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("accessKeyId", accessKeyId.String()).
			Msg("Failed to delete access key")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_delete_access_key"),
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
