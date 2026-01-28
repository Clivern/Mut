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

// ListClustersAction returns a page of a workspace's registered clusters.
func (a *API) ListClustersAction(w http.ResponseWriter, r *http.Request) {
	workspaceId := db.Id(chi.URLParam(r, "workspaceId"))

	user, _ := a.GetUser(r)

	limit, offset := a.ParsePagination(r)

	result, err := a.Cluster.ListClusters(workspaceId, user, limit, offset)
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
			Msg("Failed to list clusters")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_cluster_request"),
		})
		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"clusters": result.Clusters,
		"_meta": map[string]any{
			"limit":  limit,
			"offset": offset,
			"total":  result.Total,
		},
	})
}

// CreateClusterAction registers a cluster in a workspace.
func (a *API) CreateClusterAction(w http.ResponseWriter, r *http.Request) {
	workspaceId := db.Id(chi.URLParam(r, "workspaceId"))

	var req module.CreateClusterRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	user, _ := a.GetUser(r)

	cluster, err := a.Cluster.CreateCluster(workspaceId, user, &req)
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
			Msg("Failed to create cluster")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_cluster_request"),
		})
		return
	}

	log.Info().
		Str("userId", user.Id.String()).
		Str("workspaceId", workspaceId.String()).
		Str("clusterId", cluster.Id.String()).
		Msg("Cluster registered")

	a.WriteJSON(w, http.StatusCreated, cluster)
}

// GetClusterDetailAction returns one of a workspace's registered clusters.
func (a *API) GetClusterDetailAction(w http.ResponseWriter, r *http.Request) {
	workspaceId := db.Id(chi.URLParam(r, "workspaceId"))
	clusterId := db.Id(chi.URLParam(r, "clusterId"))

	user, _ := a.GetUser(r)

	cluster, err := a.Cluster.GetCluster(workspaceId, user, clusterId)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrWorkspaceNotFound, http.StatusNotFound, locale.TR(r, "workspace_not_found")},
			{module.ErrWorkspaceForbidden, http.StatusForbidden, locale.TR(r, "workspace_forbidden")},
			{module.ErrClusterNotFound, http.StatusNotFound, locale.TR(r, "cluster_not_found")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("clusterId", clusterId.String()).
			Msg("Failed to get cluster")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_cluster_request"),
		})
		return
	}

	a.WriteJSON(w, http.StatusOK, cluster)
}

// UpdateClusterAction renames one of a workspace's registered clusters or changes its description.
func (a *API) UpdateClusterAction(w http.ResponseWriter, r *http.Request) {
	workspaceId := db.Id(chi.URLParam(r, "workspaceId"))
	clusterId := db.Id(chi.URLParam(r, "clusterId"))

	var req module.UpdateClusterRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	user, _ := a.GetUser(r)

	cluster, err := a.Cluster.UpdateCluster(workspaceId, user, clusterId, &req)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrWorkspaceNotFound, http.StatusNotFound, locale.TR(r, "workspace_not_found")},
			{module.ErrWorkspaceForbidden, http.StatusForbidden, locale.TR(r, "workspace_forbidden")},
			{module.ErrClusterNotFound, http.StatusNotFound, locale.TR(r, "cluster_not_found")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("clusterId", clusterId.String()).
			Msg("Failed to update cluster")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_cluster_request"),
		})
		return
	}

	a.WriteJSON(w, http.StatusOK, cluster)
}

// DeleteClusterAction removes one of a workspace's registered clusters.
func (a *API) DeleteClusterAction(w http.ResponseWriter, r *http.Request) {
	workspaceId := db.Id(chi.URLParam(r, "workspaceId"))
	clusterId := db.Id(chi.URLParam(r, "clusterId"))

	user, _ := a.GetUser(r)

	err := a.Cluster.DeleteCluster(workspaceId, user, clusterId)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrWorkspaceNotFound, http.StatusNotFound, locale.TR(r, "workspace_not_found")},
			{module.ErrWorkspaceForbidden, http.StatusForbidden, locale.TR(r, "workspace_forbidden")},
			{module.ErrClusterNotFound, http.StatusNotFound, locale.TR(r, "cluster_not_found")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("clusterId", clusterId.String()).
			Msg("Failed to delete cluster")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_cluster_request"),
		})
		return
	}

	log.Info().
		Str("userId", user.Id.String()).
		Str("workspaceId", workspaceId.String()).
		Str("clusterId", clusterId.String()).
		Msg("Cluster removed")

	w.WriteHeader(http.StatusNoContent)
}
