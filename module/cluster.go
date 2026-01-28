// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"time"

	"github.com/clivern/mut/db"
)

// CreateClusterRequest is the body for registering a cluster.
type CreateClusterRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100" label:"Name"`
	Description string `json:"description" validate:"required,max=500" label:"Description"`
}

// UpdateClusterRequest is the body for renaming a cluster or changing its description.
type UpdateClusterRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100" label:"Name"`
	Description string `json:"description" validate:"required,max=500" label:"Description"`
}

// ClusterResponse is one registered cluster.
type ClusterResponse struct {
	Id          db.Id  `json:"id"`
	WorkspaceId db.Id  `json:"workspaceId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// StateCollectedAt is when the agent last pushed state, or nil before its first push.
	StateCollectedAt *time.Time `json:"stateCollectedAt"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

// ListClustersResult is one page of registered clusters with the total count.
type ListClustersResult struct {
	Clusters []*ClusterResponse
	Total    int64
}

// CreateCluster stores a cluster registered in a workspace.
func (c *Cluster) CreateCluster(workspaceId db.Id, user *db.User, req *CreateClusterRequest) (*ClusterResponse, error) {
	cluster := &db.Cluster{
		WorkspaceId: workspaceId,
		Name:        req.Name,
		Description: req.Description,
	}

	err := c.ClusterRepository.Create(cluster)
	if err != nil {
		return nil, err
	}

	return toClusterResponse(cluster, nil), nil
}

// ListClusters returns a page of a workspace's clusters.
func (c *Cluster) ListClusters(workspaceId db.Id, user *db.User, limit, offset int) (*ListClustersResult, error) {
	clusters, err := c.ClusterRepository.ListByWorkspaceId(workspaceId, limit, offset)
	if err != nil {
		return nil, err
	}

	total, err := c.ClusterRepository.CountByWorkspaceId(workspaceId)
	if err != nil {
		return nil, err
	}

	responses := make([]*ClusterResponse, 0, len(clusters))
	for _, cluster := range clusters {
		collectedAt, err := c.latestStateCollectedAt(cluster.Id)
		if err != nil {
			return nil, err
		}

		responses = append(responses, toClusterResponse(cluster, collectedAt))
	}

	return &ListClustersResult{Clusters: responses, Total: total}, nil
}

// GetCluster returns one of a workspace's clusters.
func (c *Cluster) GetCluster(workspaceId db.Id, user *db.User, clusterId db.Id) (*ClusterResponse, error) {
	cluster, err := c.ClusterRepository.GetById(clusterId)
	if err != nil {
		return nil, err
	}

	if cluster == nil || cluster.WorkspaceId != workspaceId {
		return nil, ErrClusterNotFound
	}

	collectedAt, err := c.latestStateCollectedAt(cluster.Id)
	if err != nil {
		return nil, err
	}

	return toClusterResponse(cluster, collectedAt), nil
}

// UpdateCluster renames one of a workspace's clusters or changes its description.
func (c *Cluster) UpdateCluster(workspaceId db.Id, user *db.User, clusterId db.Id, req *UpdateClusterRequest) (*ClusterResponse, error) {
	cluster, err := c.ClusterRepository.GetById(clusterId)
	if err != nil {
		return nil, err
	}

	if cluster == nil || cluster.WorkspaceId != workspaceId {
		return nil, ErrClusterNotFound
	}

	cluster.Name = req.Name
	cluster.Description = req.Description

	err = c.ClusterRepository.Update(cluster)
	if err != nil {
		return nil, err
	}

	collectedAt, err := c.latestStateCollectedAt(cluster.Id)
	if err != nil {
		return nil, err
	}

	return toClusterResponse(cluster, collectedAt), nil
}

// DeleteCluster removes one of a workspace's clusters and its stored states.
func (c *Cluster) DeleteCluster(workspaceId db.Id, user *db.User, clusterId db.Id) error {
	cluster, err := c.ClusterRepository.GetById(clusterId)
	if err != nil {
		return err
	}

	if cluster == nil || cluster.WorkspaceId != workspaceId {
		return ErrClusterNotFound
	}

	return c.ClusterRepository.Delete(clusterId)
}

// latestStateCollectedAt returns when a cluster's latest state was collected, or nil when no agent has pushed yet.
func (c *Cluster) latestStateCollectedAt(clusterId db.Id) (*time.Time, error) {
	states, err := c.ClusterStateRepository.ListByClusterId(clusterId, 1)
	if err != nil {
		return nil, err
	}

	if len(states) == 0 {
		return nil, nil
	}

	return &states[0].CollectedAt, nil
}

// toClusterResponse shapes a stored cluster for API responses.
func toClusterResponse(cluster *db.Cluster, stateCollectedAt *time.Time) *ClusterResponse {
	return &ClusterResponse{
		Id:               cluster.Id,
		WorkspaceId:      cluster.WorkspaceId,
		Name:             cluster.Name,
		Description:      cluster.Description,
		StateCollectedAt: stateCollectedAt,
		CreatedAt:        cluster.CreatedAt,
		UpdatedAt:        cluster.UpdatedAt,
	}
}
