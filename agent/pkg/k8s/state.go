// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package k8s

import (
	"time"
)

// ClusterState is the typed view of everything Mut fetches from one Kubernetes cluster.
type ClusterState struct {
	Cluster         ClusterMeta         `json:"cluster"`
	Nodes           []NodeState         `json:"nodes"`
	Namespaces      []NamespaceState    `json:"namespaces"`
	Deployments     []DeploymentState   `json:"deployments"`
	StatefulSets    []StatefulSetState  `json:"statefulSets"`
	DaemonSets      []DaemonSetState    `json:"daemonSets"`
	ReplicaSets     []ReplicaSetState   `json:"replicaSets"`
	Jobs            []JobState          `json:"jobs"`
	CronJobs        []CronJobState      `json:"cronJobs"`
	Pods            []PodState          `json:"pods"`
	Services        []ServiceState      `json:"services"`
	Ingresses       []IngressState      `json:"ingresses"`
	ConfigMaps      []ConfigMapState    `json:"configMaps"`
	Secrets         []SecretState       `json:"secrets"`
	Volumes         []VolumeState       `json:"volumes"`
	Claims          []ClaimState        `json:"claims"`
	StorageClasses  []StorageClassState `json:"storageClasses"`
	Autoscalers     []AutoscalerState   `json:"autoscalers"`
	Disruptions     []DisruptionState   `json:"disruptions"`
	Quotas          []QuotaState        `json:"quotas"`
	Policies        []PolicyState       `json:"policies"`
	Endpoints       []EndpointState     `json:"endpoints"`
	Accounts        []AccountState      `json:"accounts"`
	Roles           []RoleState         `json:"roles"`
	Bindings        []BindingState      `json:"bindings"`
	ClusterRoles    []RoleState         `json:"clusterRoles"`
	ClusterBindings []BindingState      `json:"clusterBindings"`
	Limits          []LimitState        `json:"limits"`
	Priorities      []PriorityState     `json:"priorities"`
	IngressClasses  []IngressClassState `json:"ingressClasses"`
	RuntimeClasses  []RuntimeClassState `json:"runtimeClasses"`
	Events          []EventState        `json:"events"`
	Errors          map[string]string   `json:"errors,omitempty"`
}

// ClusterMeta identifies the snapshot source.
type ClusterMeta struct {
	Name          string    `json:"name"`
	ServerVersion string    `json:"serverVersion"`
	CollectedAt   time.Time `json:"collectedAt"`
}

// ObjectMeta is the shared identity subset. Embedded without a tag so its fields flatten into the parent object.
type ObjectMeta struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	CreationTimestamp time.Time         `json:"creationTimestamp"`
}

// ResourceUsage holds observed usage when metrics are available.
type ResourceUsage struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

// ConditionState is one status condition.
type ConditionState struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

// ContainerSpec describes one container in a pod template.
type ContainerSpec struct {
	Name     string            `json:"name"`
	Image    string            `json:"image"`
	Requests map[string]string `json:"requests,omitempty"`
	Limits   map[string]string `json:"limits,omitempty"`
}

// NodeState describes capacity, pressure and version of one node.
type NodeState struct {
	ObjectMeta
	Capacity         map[string]string `json:"capacity,omitempty"`
	Allocatable      map[string]string `json:"allocatable,omitempty"`
	Conditions       []ConditionState  `json:"conditions,omitempty"`
	Taints           []TaintState      `json:"taints,omitempty"`
	Kubelet          string            `json:"kubelet,omitempty"`
	ContainerRuntime string            `json:"containerRuntime,omitempty"`
	Usage            *ResourceUsage    `json:"usage,omitempty"`
}

// TaintState describes one node taint.
type TaintState struct {
	Key    string `json:"key"`
	Value  string `json:"value,omitempty"`
	Effect string `json:"effect"`
}

// NamespaceState describes one namespace.
type NamespaceState struct {
	ObjectMeta
	Phase string `json:"phase,omitempty"`
}

// WorkloadState holds the fields every controller shares.
type WorkloadState struct {
	ObjectMeta
	ReplicasDesired   int32            `json:"replicasDesired"`
	ReplicasReady     int32            `json:"replicasReady"`
	ReplicasAvailable int32            `json:"replicasAvailable,omitempty"`
	ReplicasUpdated   int32            `json:"replicasUpdated,omitempty"`
	Containers        []ContainerSpec  `json:"containers,omitempty"`
	Conditions        []ConditionState `json:"conditions,omitempty"`
}

// DeploymentState describes one deployment.
type DeploymentState struct {
	WorkloadState
	Strategy string `json:"strategy,omitempty"`
}

// StatefulSetState describes one stateful set.
type StatefulSetState struct {
	WorkloadState
	ServiceName string `json:"serviceName,omitempty"`
}

// DaemonSetState describes one daemon set.
type DaemonSetState struct {
	WorkloadState
	DesiredScheduled     int32 `json:"desiredScheduled"`
	ReadyScheduled       int32 `json:"readyScheduled,omitempty"`
	UpdatedScheduled     int32 `json:"updatedScheduled,omitempty"`
	UnavailableScheduled int32 `json:"unavailableScheduled,omitempty"`
}

// ReplicaSetState describes one replica set.
type ReplicaSetState struct {
	WorkloadState
}

// JobState describes one job.
type JobState struct {
	ObjectMeta
	Active     int32            `json:"active"`
	Succeeded  int32            `json:"succeeded"`
	Failed     int32            `json:"failed"`
	Containers []ContainerSpec  `json:"containers,omitempty"`
	Conditions []ConditionState `json:"conditions,omitempty"`
}

// CronJobState describes one cron job.
type CronJobState struct {
	ObjectMeta
	Schedule     string          `json:"schedule,omitempty"`
	Suspend      bool            `json:"suspend"`
	Active       int32           `json:"active"`
	LastSchedule *time.Time      `json:"lastSchedule,omitempty"`
	Containers   []ContainerSpec `json:"containers,omitempty"`
}

// ContainerStatusState describes one running or waiting container.
type ContainerStatusState struct {
	Name         string     `json:"name"`
	Image        string     `json:"image"`
	Ready        bool       `json:"ready"`
	Restarts     int32      `json:"restarts"`
	State        string     `json:"state,omitempty"`
	StateReason  string     `json:"stateReason,omitempty"`
	WaitingSince *time.Time `json:"waitingSince,omitempty"`
}

// PodState describes one pod with summed requests and limits for rightsizing.
type PodState struct {
	ObjectMeta
	Node         string                 `json:"node,omitempty"`
	Phase        string                 `json:"phase,omitempty"`
	PodIP        string                 `json:"podIP,omitempty"`
	OwnerKind    string                 `json:"ownerKind,omitempty"`
	OwnerName    string                 `json:"ownerName,omitempty"`
	Conditions   []ConditionState       `json:"conditions,omitempty"`
	Restarts     int32                  `json:"restarts"`
	InitRestarts int32                  `json:"initRestarts,omitempty"`
	Containers   []ContainerStatusState `json:"containers,omitempty"`
	Requests     map[string]string      `json:"requests,omitempty"`
	Limits       map[string]string      `json:"limits,omitempty"`
	Usage        *ResourceUsage         `json:"usage,omitempty"`
}

// ServicePortState describes one service port.
type ServicePortState struct {
	Name       string `json:"name,omitempty"`
	Protocol   string `json:"protocol,omitempty"`
	Port       int32  `json:"port"`
	TargetPort string `json:"targetPort,omitempty"`
	NodePort   int32  `json:"nodePort,omitempty"`
}

// ServiceState describes one service.
type ServiceState struct {
	ObjectMeta
	Type      string             `json:"type,omitempty"`
	ClusterIP string             `json:"clusterIP,omitempty"`
	Ports     []ServicePortState `json:"ports,omitempty"`
}

// IngressRuleState describes one ingress host rule.
type IngressRuleState struct {
	Host  string   `json:"host,omitempty"`
	Paths []string `json:"paths,omitempty"`
}

// IngressState describes one ingress.
type IngressState struct {
	ObjectMeta
	Class string             `json:"class,omitempty"`
	Rules []IngressRuleState `json:"rules,omitempty"`
	TLS   []string           `json:"tls,omitempty"`
}

// ConfigMapState describes one config map by key names only.
type ConfigMapState struct {
	ObjectMeta
	Keys []string `json:"keys,omitempty"`
}

// SecretState describes one secret by key names only, never values.
type SecretState struct {
	ObjectMeta
	Type string   `json:"type,omitempty"`
	Keys []string `json:"keys,omitempty"`
}

// VolumeState describes one persistent volume.
type VolumeState struct {
	ObjectMeta
	Capacity     string `json:"capacity,omitempty"`
	Phase        string `json:"phase,omitempty"`
	StorageClass string `json:"storageClass,omitempty"`
	Claim        string `json:"claim,omitempty"`
}

// ClaimState describes one persistent volume claim.
type ClaimState struct {
	ObjectMeta
	Phase        string `json:"phase,omitempty"`
	StorageClass string `json:"storageClass,omitempty"`
	Volume       string `json:"volume,omitempty"`
	Requested    string `json:"requested,omitempty"`
}

// StorageClassState describes one storage class.
type StorageClassState struct {
	ObjectMeta
	Provisioner string `json:"provisioner,omitempty"`
	Reclaim     string `json:"reclaim,omitempty"`
	Expansion   bool   `json:"expansion"`
}

// AutoscalerState describes one horizontal pod autoscaler.
type AutoscalerState struct {
	ObjectMeta
	TargetKind      string           `json:"targetKind,omitempty"`
	TargetName      string           `json:"targetName,omitempty"`
	MinReplicas     int32            `json:"minReplicas"`
	MaxReplicas     int32            `json:"maxReplicas"`
	CurrentReplicas int32            `json:"currentReplicas"`
	DesiredReplicas int32            `json:"desiredReplicas"`
	Conditions      []ConditionState `json:"conditions,omitempty"`
}

// DisruptionState describes one pod disruption budget.
type DisruptionState struct {
	ObjectMeta
	Allowed        int32 `json:"allowed"`
	CurrentHealthy int32 `json:"currentHealthy"`
	DesiredHealthy int32 `json:"desiredHealthy"`
	ExpectedPods   int32 `json:"expectedPods"`
}

// QuotaState describes one resource quota with hard limits and usage.
type QuotaState struct {
	ObjectMeta
	Hard map[string]string `json:"hard,omitempty"`
	Used map[string]string `json:"used,omitempty"`
}

// PolicyState describes one network policy.
type PolicyState struct {
	ObjectMeta
	Ingress bool `json:"ingress"`
	Egress  bool `json:"egress"`
}

// EndpointState describes one endpoint slice port-address group.
type EndpointState struct {
	ObjectMeta
	Service     string   `json:"service,omitempty"`
	Addresses   []string `json:"addresses,omitempty"`
	Ready       int32    `json:"ready"`
	Serving     int32    `json:"serving,omitempty"`
	Terminating int32    `json:"terminating,omitempty"`
}

// AccountState describes one service account.
type AccountState struct {
	ObjectMeta
	Secrets []string `json:"secrets,omitempty"`
}

// RuleState describes one RBAC policy rule.
type RuleState struct {
	Groups    []string `json:"groups,omitempty"`
	Resources []string `json:"resources,omitempty"`
	Names     []string `json:"names,omitempty"`
	Verbs     []string `json:"verbs,omitempty"`
}

// RoleState describes one role or cluster role.
type RoleState struct {
	ObjectMeta
	Rules []RuleState `json:"rules,omitempty"`
}

// SubjectState describes one RBAC subject.
type SubjectState struct {
	Kind      string `json:"kind,omitempty"`
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// BindingState describes one role or cluster role binding.
type BindingState struct {
	ObjectMeta
	RoleKind string         `json:"roleKind,omitempty"`
	RoleName string         `json:"roleName,omitempty"`
	Subjects []SubjectState `json:"subjects,omitempty"`
}

// LimitState describes one limit range with default constraints.
type LimitState struct {
	ObjectMeta
	Defaults map[string]string `json:"defaults,omitempty"`
}

// PriorityState describes one priority class.
type PriorityState struct {
	ObjectMeta
	Value         int32 `json:"value"`
	GlobalDefault bool  `json:"globalDefault"`
}

// IngressClassState describes one ingress class.
type IngressClassState struct {
	ObjectMeta
	Controller string `json:"controller,omitempty"`
	Default    bool   `json:"default"`
}

// RuntimeClassState describes one runtime class.
type RuntimeClassState struct {
	ObjectMeta
	Handler string `json:"handler,omitempty"`
}

// EventState describes one cluster event, newest first.
type EventState struct {
	Namespace string    `json:"namespace,omitempty"`
	Kind      string    `json:"kind,omitempty"`
	Name      string    `json:"name,omitempty"`
	Type      string    `json:"type,omitempty"`
	Reason    string    `json:"reason,omitempty"`
	Message   string    `json:"message,omitempty"`
	Count     int32     `json:"count"`
	FirstSeen time.Time `json:"firstSeen"`
	LastSeen  time.Time `json:"lastSeen"`
}

// NewClusterState returns an empty state for the named cluster.
func NewClusterState(name, serverVersion string) *ClusterState {
	return &ClusterState{
		Cluster: ClusterMeta{
			Name:          name,
			ServerVersion: serverVersion,
			CollectedAt:   time.Now().UTC(),
		},
		Nodes:           []NodeState{},
		Namespaces:      []NamespaceState{},
		Deployments:     []DeploymentState{},
		StatefulSets:    []StatefulSetState{},
		DaemonSets:      []DaemonSetState{},
		ReplicaSets:     []ReplicaSetState{},
		Jobs:            []JobState{},
		CronJobs:        []CronJobState{},
		Pods:            []PodState{},
		Services:        []ServiceState{},
		Ingresses:       []IngressState{},
		ConfigMaps:      []ConfigMapState{},
		Secrets:         []SecretState{},
		Volumes:         []VolumeState{},
		Claims:          []ClaimState{},
		StorageClasses:  []StorageClassState{},
		Autoscalers:     []AutoscalerState{},
		Disruptions:     []DisruptionState{},
		Quotas:          []QuotaState{},
		Policies:        []PolicyState{},
		Endpoints:       []EndpointState{},
		Accounts:        []AccountState{},
		Roles:           []RoleState{},
		Bindings:        []BindingState{},
		ClusterRoles:    []RoleState{},
		ClusterBindings: []BindingState{},
		Limits:          []LimitState{},
		Priorities:      []PriorityState{},
		IngressClasses:  []IngressClassState{},
		RuntimeClasses:  []RuntimeClassState{},
		Events:          []EventState{},
		Errors:          map[string]string{},
	}
}
