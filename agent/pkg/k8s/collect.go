// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package k8s

import (
	"context"
	"sort"

	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// eventCap bounds how many events a state carries, newest first.
const eventCap = 500

// listLimit pages large collections.
const listLimit = 1000

// Collect walks the API with typed clients and fills a ClusterState.
func (c *Cluster) Collect(ctx context.Context) *ClusterState {
	state := NewClusterState(c.config.Name, "")

	if version, err := c.clientset.Discovery().ServerVersion(); err != nil {
		state.Errors["version"] = err.Error()
	} else {
		state.Cluster.ServerVersion = version.GitVersion
	}

	collectors := map[string]func(context.Context, *ClusterState) error{
		"namespaces":      c.CollectNamespaces,
		"nodes":           c.CollectNodes,
		"pods":            c.CollectPods,
		"deployments":     c.CollectDeployments,
		"statefulSets":    c.CollectStatefulSets,
		"daemonSets":      c.CollectDaemonSets,
		"replicaSets":     c.CollectReplicaSets,
		"jobs":            c.CollectJobs,
		"cronJobs":        c.CollectCronJobs,
		"services":        c.CollectServices,
		"ingresses":       c.CollectIngresses,
		"configMaps":      c.CollectConfigMaps,
		"secrets":         c.CollectSecrets,
		"volumes":         c.CollectVolumes,
		"claims":          c.CollectClaims,
		"storageClasses":  c.CollectStorageClasses,
		"autoscalers":     c.CollectAutoscalers,
		"disruptions":     c.CollectDisruptions,
		"quotas":          c.CollectQuotas,
		"policies":        c.CollectPolicies,
		"endpoints":       c.CollectEndpoints,
		"accounts":        c.CollectAccounts,
		"roles":           c.CollectRoles,
		"bindings":        c.CollectBindings,
		"clusterRoles":    c.CollectClusterRoles,
		"clusterBindings": c.CollectClusterBindings,
		"limits":          c.CollectLimits,
		"priorities":      c.CollectPriorities,
		"ingressClasses":  c.CollectIngressClasses,
		"runtimeClasses":  c.CollectRuntimeClasses,
		"events":          c.CollectEvents,
	}

	for key, collect := range collectors {
		if err := collect(ctx, state); err != nil {
			log.Warn().
				Err(err).
				Str("cluster", c.config.Name).
				Str("resource", key).
				Msg("Failed to collect resource for cluster state")
			state.Errors[key] = err.Error()
		}
	}

	return state
}

// objectMeta extracts the shared identity subset. CollectNamespaces lists all namespaces.
func (c *Cluster) CollectNamespaces(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		state.Namespaces = append(state.Namespaces, NamespaceState{
			ObjectMeta: toObjectMeta(item.ObjectMeta),
			Phase:      string(item.Status.Phase),
		})
	}

	return nil
}

// CollectNodes lists all nodes with capacity, taints and pressure.
func (c *Cluster) CollectNodes(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		node := NodeState{
			ObjectMeta:       toObjectMeta(item.ObjectMeta),
			Capacity:         toQuantityMap(item.Status.Capacity),
			Allocatable:      toQuantityMap(item.Status.Allocatable),
			Kubelet:          item.Status.NodeInfo.KubeletVersion,
			ContainerRuntime: item.Status.NodeInfo.ContainerRuntimeVersion,
		}

		for _, condition := range item.Status.Conditions {
			node.Conditions = append(node.Conditions, ConditionState{
				Type:    string(condition.Type),
				Status:  string(condition.Status),
				Reason:  condition.Reason,
				Message: condition.Message,
			})
		}

		for _, taint := range item.Spec.Taints {
			node.Taints = append(node.Taints, TaintState{
				Key:    taint.Key,
				Value:  taint.Value,
				Effect: string(taint.Effect),
			})
		}

		state.Nodes = append(state.Nodes, node)
	}

	return nil
}

// CollectPods lists pods across all namespaces, paging through large clusters.
func (c *Cluster) CollectPods(ctx context.Context, state *ClusterState) error {
	continueToken := ""

	for {
		list, err := c.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{
			Limit:    listLimit,
			Continue: continueToken,
		})
		if err != nil {
			return err
		}

		for _, item := range list.Items {
			pod := PodState{
				ObjectMeta: toObjectMeta(item.ObjectMeta),
				Node:       item.Spec.NodeName,
				Phase:      string(item.Status.Phase),
				PodIP:      item.Status.PodIP,
				Conditions: toPodConditions(item.Status.Conditions),
			}

			if len(item.OwnerReferences) > 0 {
				pod.OwnerKind = item.OwnerReferences[0].Kind
				pod.OwnerName = item.OwnerReferences[0].Name
			}

			containers, restarts := toContainerStatuses(item.Status.ContainerStatuses)
			initStatuses, initRestarts := toContainerStatuses(item.Status.InitContainerStatuses)
			pod.Containers = append(containers, initStatuses...)
			pod.Restarts = restarts
			pod.InitRestarts = initRestarts

			for _, container := range item.Spec.Containers {
				pod.Requests = mergeQuantities(pod.Requests, container.Resources.Requests)
				pod.Limits = mergeQuantities(pod.Limits, container.Resources.Limits)
			}
			for _, container := range item.Spec.InitContainers {
				pod.Requests = mergeQuantities(pod.Requests, container.Resources.Requests)
				pod.Limits = mergeQuantities(pod.Limits, container.Resources.Limits)
			}

			state.Pods = append(state.Pods, pod)
		}

		continueToken = list.GetContinue()
		if continueToken == "" {
			break
		}
	}

	return nil
}

// CollectDeployments lists deployments across all namespaces.
func (c *Cluster) CollectDeployments(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		base := newWorkloadState(
			item.ObjectMeta,
			item.Spec.Template.Spec.Containers,
			replicasOrDefault(item.Spec.Replicas),
			item.Status.ReadyReplicas,
			item.Status.AvailableReplicas,
			item.Status.UpdatedReplicas,
			toDeploymentConditions(item.Status.Conditions),
		)
		state.Deployments = append(state.Deployments, DeploymentState{
			WorkloadState: base,
			Strategy:      string(item.Spec.Strategy.Type),
		})
	}

	return nil
}

// CollectStatefulSets lists stateful sets across all namespaces.
func (c *Cluster) CollectStatefulSets(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.AppsV1().StatefulSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		base := newWorkloadState(
			item.ObjectMeta,
			item.Spec.Template.Spec.Containers,
			replicasOrDefault(item.Spec.Replicas),
			item.Status.ReadyReplicas,
			0,
			item.Status.UpdatedReplicas,
			nil,
		)
		state.StatefulSets = append(state.StatefulSets, StatefulSetState{
			WorkloadState: base,
			ServiceName:   item.Spec.ServiceName,
		})
	}

	return nil
}

// CollectDaemonSets lists daemon sets across all namespaces.
func (c *Cluster) CollectDaemonSets(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.AppsV1().DaemonSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		state.DaemonSets = append(state.DaemonSets, DaemonSetState{
			WorkloadState: newWorkloadState(
				item.ObjectMeta,
				item.Spec.Template.Spec.Containers,
				item.Status.DesiredNumberScheduled,
				item.Status.NumberReady,
				0,
				item.Status.UpdatedNumberScheduled,
				nil,
			),
			DesiredScheduled:     item.Status.DesiredNumberScheduled,
			ReadyScheduled:       item.Status.NumberReady,
			UpdatedScheduled:     item.Status.UpdatedNumberScheduled,
			UnavailableScheduled: item.Status.NumberUnavailable,
		})
	}

	return nil
}

// CollectReplicaSets lists replica sets across all namespaces.
func (c *Cluster) CollectReplicaSets(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.AppsV1().ReplicaSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		base := newWorkloadState(
			item.ObjectMeta,
			item.Spec.Template.Spec.Containers,
			replicasOrDefault(item.Spec.Replicas),
			item.Status.ReadyReplicas,
			item.Status.AvailableReplicas,
			0,
			nil,
		)
		state.ReplicaSets = append(state.ReplicaSets, ReplicaSetState{WorkloadState: base})
	}

	return nil
}

// CollectJobs lists jobs across all namespaces.
func (c *Cluster) CollectJobs(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.BatchV1().Jobs("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		state.Jobs = append(state.Jobs, JobState{
			ObjectMeta: toObjectMeta(item.ObjectMeta),
			Active:     item.Status.Active,
			Succeeded:  item.Status.Succeeded,
			Failed:     item.Status.Failed,
			Containers: toContainerSpecs(item.Spec.Template.Spec.Containers),
			Conditions: toJobConditions(item.Status.Conditions),
		})
	}

	return nil
}

// CollectCronJobs lists cron jobs across all namespaces.
func (c *Cluster) CollectCronJobs(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.BatchV1().CronJobs("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		job := CronJobState{
			ObjectMeta: toObjectMeta(item.ObjectMeta),
			Schedule:   item.Spec.Schedule,
			Active:     int32(len(item.Status.Active)),
			Containers: toContainerSpecs(item.Spec.JobTemplate.Spec.Template.Spec.Containers),
		}

		if item.Spec.Suspend != nil {
			job.Suspend = *item.Spec.Suspend
		}
		if item.Status.LastScheduleTime != nil {
			when := item.Status.LastScheduleTime.Time
			job.LastSchedule = &when
		}

		state.CronJobs = append(state.CronJobs, job)
	}

	return nil
}

// CollectServices lists services across all namespaces.
func (c *Cluster) CollectServices(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		service := ServiceState{
			ObjectMeta: toObjectMeta(item.ObjectMeta),
			Type:       string(item.Spec.Type),
			ClusterIP:  item.Spec.ClusterIP,
		}

		for _, port := range item.Spec.Ports {
			service.Ports = append(service.Ports, ServicePortState{
				Name:       port.Name,
				Protocol:   string(port.Protocol),
				Port:       port.Port,
				TargetPort: port.TargetPort.String(),
				NodePort:   port.NodePort,
			})
		}

		state.Services = append(state.Services, service)
	}

	return nil
}

// CollectIngresses lists ingresses across all namespaces.
func (c *Cluster) CollectIngresses(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		ingress := IngressState{ObjectMeta: toObjectMeta(item.ObjectMeta)}

		if item.Spec.IngressClassName != nil {
			ingress.Class = *item.Spec.IngressClassName
		}

		for _, rule := range item.Spec.Rules {
			entry := IngressRuleState{Host: rule.Host}
			if rule.HTTP != nil {
				for _, path := range rule.HTTP.Paths {
					entry.Paths = append(entry.Paths, path.Path)
				}
			}
			ingress.Rules = append(ingress.Rules, entry)
		}

		for _, tls := range item.Spec.TLS {
			ingress.TLS = append(ingress.TLS, tls.Hosts...)
		}

		state.Ingresses = append(state.Ingresses, ingress)
	}

	return nil
}

// CollectConfigMaps lists config maps by key names across all namespaces.
func (c *Cluster) CollectConfigMaps(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.CoreV1().ConfigMaps("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		keys := make([]string, 0, len(item.Data)+len(item.BinaryData))
		for key := range item.Data {
			keys = append(keys, key)
		}
		for key := range item.BinaryData {
			keys = append(keys, key)
		}

		state.ConfigMaps = append(state.ConfigMaps, ConfigMapState{
			ObjectMeta: toObjectMeta(item.ObjectMeta),
			Keys:       keys,
		})
	}

	return nil
}

// CollectSecrets lists secrets by key names across all namespaces, never values.
func (c *Cluster) CollectSecrets(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.CoreV1().Secrets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		keys := make([]string, 0, len(item.Data)+len(item.StringData))
		for key := range item.Data {
			keys = append(keys, key)
		}
		for key := range item.StringData {
			keys = append(keys, key)
		}

		state.Secrets = append(state.Secrets, SecretState{
			ObjectMeta: toObjectMeta(item.ObjectMeta),
			Type:       string(item.Type),
			Keys:       keys,
		})
	}

	return nil
}

// CollectVolumes lists persistent volumes.
func (c *Cluster) CollectVolumes(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		volume := VolumeState{
			ObjectMeta:   toObjectMeta(item.ObjectMeta),
			Capacity:     item.Spec.Capacity.Storage().String(),
			Phase:        string(item.Status.Phase),
			StorageClass: item.Spec.StorageClassName,
		}

		if item.Spec.ClaimRef != nil {
			volume.Claim = item.Spec.ClaimRef.Namespace + "/" + item.Spec.ClaimRef.Name
		}

		state.Volumes = append(state.Volumes, volume)
	}

	return nil
}

// CollectClaims lists persistent volume claims across all namespaces.
func (c *Cluster) CollectClaims(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		claim := ClaimState{
			ObjectMeta: toObjectMeta(item.ObjectMeta),
			Phase:      string(item.Status.Phase),
			Volume:     item.Spec.VolumeName,
		}

		if item.Spec.StorageClassName != nil {
			claim.StorageClass = *item.Spec.StorageClassName
		}
		if quantity, ok := item.Spec.Resources.Requests[corev1.ResourceStorage]; ok {
			claim.Requested = quantity.String()
		}

		state.Claims = append(state.Claims, claim)
	}

	return nil
}

// CollectStorageClasses lists storage classes.
func (c *Cluster) CollectStorageClasses(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		class := StorageClassState{
			ObjectMeta:  toObjectMeta(item.ObjectMeta),
			Provisioner: item.Provisioner,
		}

		if item.ReclaimPolicy != nil {
			class.Reclaim = string(*item.ReclaimPolicy)
		}
		if item.AllowVolumeExpansion != nil {
			class.Expansion = *item.AllowVolumeExpansion
		}

		state.StorageClasses = append(state.StorageClasses, class)
	}

	return nil
}

// CollectAutoscalers lists horizontal pod autoscalers across all namespaces.
func (c *Cluster) CollectAutoscalers(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.AutoscalingV2().HorizontalPodAutoscalers("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		scaler := AutoscalerState{
			ObjectMeta:      toObjectMeta(item.ObjectMeta),
			TargetKind:      item.Spec.ScaleTargetRef.Kind,
			TargetName:      item.Spec.ScaleTargetRef.Name,
			MaxReplicas:     item.Spec.MaxReplicas,
			CurrentReplicas: item.Status.CurrentReplicas,
			DesiredReplicas: item.Status.DesiredReplicas,
		}

		if item.Spec.MinReplicas != nil {
			scaler.MinReplicas = *item.Spec.MinReplicas
		}

		for _, condition := range item.Status.Conditions {
			scaler.Conditions = append(scaler.Conditions, ConditionState{
				Type:    string(condition.Type),
				Status:  string(condition.Status),
				Reason:  condition.Reason,
				Message: condition.Message,
			})
		}

		state.Autoscalers = append(state.Autoscalers, scaler)
	}

	return nil
}

// CollectDisruptions lists pod disruption budgets across all namespaces.
func (c *Cluster) CollectDisruptions(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.PolicyV1().PodDisruptionBudgets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		state.Disruptions = append(state.Disruptions, DisruptionState{
			ObjectMeta:     toObjectMeta(item.ObjectMeta),
			Allowed:        item.Status.DisruptionsAllowed,
			CurrentHealthy: item.Status.CurrentHealthy,
			DesiredHealthy: item.Status.DesiredHealthy,
			ExpectedPods:   item.Status.ExpectedPods,
		})
	}

	return nil
}

// CollectQuotas lists resource quotas across all namespaces.
func (c *Cluster) CollectQuotas(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.CoreV1().ResourceQuotas("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		state.Quotas = append(state.Quotas, QuotaState{
			ObjectMeta: toObjectMeta(item.ObjectMeta),
			Hard:       toQuantityMap(item.Status.Hard),
			Used:       toQuantityMap(item.Status.Used),
		})
	}

	return nil
}

// CollectPolicies lists network policies across all namespaces.
func (c *Cluster) CollectPolicies(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.NetworkingV1().NetworkPolicies("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		state.Policies = append(state.Policies, PolicyState{
			ObjectMeta: toObjectMeta(item.ObjectMeta),
			Ingress:    len(item.Spec.Ingress) > 0,
			Egress:     len(item.Spec.Egress) > 0,
		})
	}

	return nil
}

// CollectEndpoints lists endpoint slices across all namespaces.
func (c *Cluster) CollectEndpoints(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.DiscoveryV1().EndpointSlices("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		slice := EndpointState{ObjectMeta: toObjectMeta(item.ObjectMeta)}

		if service, ok := item.Labels["kubernetes.io/service-name"]; ok {
			slice.Service = service
		}

		for _, endpoint := range item.Endpoints {
			slice.Addresses = append(slice.Addresses, endpoint.Addresses...)
			if endpoint.Conditions.Ready != nil && *endpoint.Conditions.Ready {
				slice.Ready++
			}
			if endpoint.Conditions.Serving != nil && *endpoint.Conditions.Serving {
				slice.Serving++
			}
			if endpoint.Conditions.Terminating != nil && *endpoint.Conditions.Terminating {
				slice.Terminating++
			}
		}

		state.Endpoints = append(state.Endpoints, slice)
	}

	return nil
}

// CollectAccounts lists service accounts across all namespaces.
func (c *Cluster) CollectAccounts(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.CoreV1().ServiceAccounts("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		account := AccountState{ObjectMeta: toObjectMeta(item.ObjectMeta)}
		for _, secret := range item.Secrets {
			account.Secrets = append(account.Secrets, secret.Name)
		}

		state.Accounts = append(state.Accounts, account)
	}

	return nil
}

// CollectRoles lists roles across all namespaces.
func (c *Cluster) CollectRoles(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.RbacV1().Roles("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		state.Roles = append(state.Roles, RoleState{
			ObjectMeta: toObjectMeta(item.ObjectMeta),
			Rules:      toRuleStates(item.Rules),
		})
	}

	return nil
}

// CollectBindings lists role bindings across all namespaces.
func (c *Cluster) CollectBindings(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.RbacV1().RoleBindings("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		state.Bindings = append(state.Bindings, BindingState{
			ObjectMeta: toObjectMeta(item.ObjectMeta),
			RoleKind:   item.RoleRef.Kind,
			RoleName:   item.RoleRef.Name,
			Subjects:   toSubjectStates(item.Subjects),
		})
	}

	return nil
}

// CollectClusterRoles lists cluster roles.
func (c *Cluster) CollectClusterRoles(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		state.ClusterRoles = append(state.ClusterRoles, RoleState{
			ObjectMeta: toObjectMeta(item.ObjectMeta),
			Rules:      toRuleStates(item.Rules),
		})
	}

	return nil
}

// CollectClusterBindings lists cluster role bindings.
func (c *Cluster) CollectClusterBindings(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		state.ClusterBindings = append(state.ClusterBindings, BindingState{
			ObjectMeta: toObjectMeta(item.ObjectMeta),
			RoleKind:   item.RoleRef.Kind,
			RoleName:   item.RoleRef.Name,
			Subjects:   toSubjectStates(item.Subjects),
		})
	}

	return nil
}

// CollectLimits lists limit ranges across all namespaces.
func (c *Cluster) CollectLimits(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.CoreV1().LimitRanges("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		limit := LimitState{ObjectMeta: toObjectMeta(item.ObjectMeta)}

		for _, itemLimit := range item.Spec.Limits {
			for name, quantity := range itemLimit.Default {
				if limit.Defaults == nil {
					limit.Defaults = map[string]string{}
				}
				limit.Defaults[string(itemLimit.Type)+"/"+string(name)] = quantity.String()
			}
		}

		state.Limits = append(state.Limits, limit)
	}

	return nil
}

// CollectPriorities lists priority classes.
func (c *Cluster) CollectPriorities(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.SchedulingV1().PriorityClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		state.Priorities = append(state.Priorities, PriorityState{
			ObjectMeta:    toObjectMeta(item.ObjectMeta),
			Value:         item.Value,
			GlobalDefault: item.GlobalDefault,
		})
	}

	return nil
}

// CollectIngressClasses lists ingress classes.
func (c *Cluster) CollectIngressClasses(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.NetworkingV1().IngressClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		class := IngressClassState{
			ObjectMeta: toObjectMeta(item.ObjectMeta),
			Controller: item.Spec.Controller,
		}

		if value, ok := item.Annotations["ingressclass.kubernetes.io/is-default-class"]; ok && value == "true" {
			class.Default = true
		}

		state.IngressClasses = append(state.IngressClasses, class)
	}

	return nil
}

// CollectRuntimeClasses lists runtime classes.
func (c *Cluster) CollectRuntimeClasses(ctx context.Context, state *ClusterState) error {
	list, err := c.clientset.NodeV1().RuntimeClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		state.RuntimeClasses = append(state.RuntimeClasses, RuntimeClassState{
			ObjectMeta: toObjectMeta(item.ObjectMeta),
			Handler:    item.Handler,
		})
	}

	return nil
}

// CollectEvents lists events newest first, capped for size.
func (c *Cluster) CollectEvents(ctx context.Context, state *ClusterState) error {
	continueToken := ""
	var events []EventState

	for {
		list, err := c.clientset.CoreV1().Events("").List(ctx, metav1.ListOptions{
			Limit:    listLimit,
			Continue: continueToken,
		})
		if err != nil {
			return err
		}

		for _, item := range list.Items {
			events = append(events, EventState{
				Namespace: item.InvolvedObject.Namespace,
				Kind:      item.InvolvedObject.Kind,
				Name:      item.InvolvedObject.Name,
				Type:      item.Type,
				Reason:    item.Reason,
				Message:   item.Message,
				Count:     item.Count,
				FirstSeen: item.FirstTimestamp.Time,
				LastSeen:  item.LastTimestamp.Time,
			})
		}

		continueToken = list.GetContinue()
		if continueToken == "" {
			break
		}
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].LastSeen.After(events[j].LastSeen)
	})

	if len(events) > eventCap {
		events = events[:eventCap]
	}

	state.Events = events

	return nil
}
