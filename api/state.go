// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/clivern/mut/db"
	"github.com/clivern/mut/locale"
	"github.com/clivern/mut/module"
	"github.com/clivern/mut/pkg/k8s"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// GetClusterStateAction returns a read-only snapshot of a workspace cluster.
func (a *API) GetClusterStateAction(w http.ResponseWriter, r *http.Request) {
	user, ok := a.GetUser(r)
	if !ok || user == nil {
		a.WriteJSON(w, http.StatusUnauthorized, map[string]any{
			"errorMessage": locale.TR(r, "not_authenticated"),
		})
		return
	}

	workspaceId := db.Id(chi.URLParam(r, "workspaceId"))
	clusterId := db.Id(chi.URLParam(r, "clusterId"))

	log.Info().
		Str("userId", user.Id.String()).
		Str("workspaceId", workspaceId.String()).
		Str("clusterId", clusterId.String()).
		Msg("Cluster state request")

	state, err := a.Cluster.GetState(r.Context(), workspaceId, clusterId)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrClusterNotFound, http.StatusNotFound, locale.TR(r, "cluster_not_found")},
			{module.ErrClusterStateNotFound, http.StatusNotFound, locale.TR(r, "cluster_state_not_found")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("userId", user.Id.String()).
			Str("workspaceId", workspaceId.String()).
			Str("clusterId", clusterId.String()).
			Msg("Failed to get cluster state")

		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_get_cluster_state"),
		})
		return
	}

	a.WriteJSON(w, http.StatusOK, state)
}

// SaveClusterStateAction stores a cluster state pushed by an in-cluster agent.
func (a *API) SaveClusterStateAction(w http.ResponseWriter, r *http.Request) {
	workspaceId := db.Id(chi.URLParam(r, "workspaceId"))
	clusterId := db.Id(chi.URLParam(r, "clusterId"))

	var state k8s.ClusterState
	err := a.DecodeAndValidate(r, &state)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	err = a.Cluster.SaveState(workspaceId, clusterId, &state)
	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrClusterNotFound, http.StatusNotFound, locale.TR(r, "cluster_not_found")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", workspaceId.String()).
			Str("clusterId", clusterId.String()).
			Msg("Failed to save cluster state")

		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_save_cluster_state"),
		})
		return
	}

	// Compact states
	err = module.EnqueueTask(
		db.AsyncTaskTypeCompactStates,
		map[string]string{"clusterId": clusterId.String()},
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("clusterId", clusterId.String()).
			Msg("Failed to enqueue cluster state compaction")
	}

	// TODO: Analyze states
	err = module.EnqueueTask(
		db.AsyncTaskTypeAnalyzeStates,
		map[string]string{"clusterId": clusterId.String()},
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("clusterId", clusterId.String()).
			Msg("Failed to enqueue cluster state analysis")
	}

	w.WriteHeader(http.StatusNoContent)
}
