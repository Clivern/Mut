// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package k8s

import (
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// toObjectMeta extracts the shared identity subset.
func toObjectMeta(meta metav1.ObjectMeta) ObjectMeta {
	return ObjectMeta{
		Name:              meta.Name,
		Namespace:         meta.Namespace,
		Labels:            meta.Labels,
		CreationTimestamp: meta.CreationTimestamp.Time,
	}
}

// toQuantityMap renders a resource list as plain strings.
func toQuantityMap(list corev1.ResourceList) map[string]string {
	if len(list) == 0 {
		return nil
	}

	out := make(map[string]string, len(list))
	for name, quantity := range list {
		out[string(name)] = quantity.String()
	}

	return out
}

// mergeQuantities sums resource lists in place.
func mergeQuantities(total map[string]string, list corev1.ResourceList) map[string]string {
	if total == nil {
		total = map[string]string{}
	}

	for name, quantity := range list {
		current := resource.MustParse("0")
		if raw, ok := total[string(name)]; ok {
			current = resource.MustParse(raw)
		}
		current.Add(quantity)
		total[string(name)] = current.String()
	}

	return total
}

// toContainerSpecs maps pod template containers.
func toContainerSpecs(containers []corev1.Container) []ContainerSpec {
	specs := make([]ContainerSpec, 0, len(containers))
	for _, container := range containers {
		specs = append(specs, ContainerSpec{
			Name:     container.Name,
			Image:    container.Image,
			Requests: toQuantityMap(container.Resources.Requests),
			Limits:   toQuantityMap(container.Resources.Limits),
		})
	}

	return specs
}

// toPodConditions maps pod conditions.
func toPodConditions(conditions []corev1.PodCondition) []ConditionState {
	out := make([]ConditionState, 0, len(conditions))
	for _, condition := range conditions {
		out = append(out, ConditionState{
			Type:    string(condition.Type),
			Status:  string(condition.Status),
			Reason:  condition.Reason,
			Message: condition.Message,
		})
	}

	return out
}

// toContainerStatuses maps running or waiting containers and sums restarts.
func toContainerStatuses(statuses []corev1.ContainerStatus) ([]ContainerStatusState, int32) {
	out := make([]ContainerStatusState, 0, len(statuses))
	var restarts int32

	for _, status := range statuses {
		restarts += status.RestartCount
		entry := ContainerStatusState{
			Name:     status.Name,
			Image:    status.Image,
			Ready:    status.Ready,
			Restarts: status.RestartCount,
		}

		switch {
		case status.State.Running != nil:
			entry.State = "Running"
		case status.State.Waiting != nil:
			entry.State = "Waiting"
			entry.StateReason = status.State.Waiting.Reason
		case status.State.Terminated != nil:
			entry.State = "Terminated"
			entry.StateReason = status.State.Terminated.Reason
		}

		out = append(out, entry)
	}

	return out, restarts
}

// replicasOrDefault returns 1 when replicas is nil.
func replicasOrDefault(replicas *int32) int32 {
	if replicas == nil {
		return 1
	}

	return *replicas
}

// newWorkloadState builds the fields every controller shares.
func newWorkloadState(meta metav1.ObjectMeta, containers []corev1.Container, desired, ready, available, updated int32, conditions []ConditionState) WorkloadState {
	return WorkloadState{
		ObjectMeta:        toObjectMeta(meta),
		ReplicasDesired:   desired,
		ReplicasReady:     ready,
		ReplicasAvailable: available,
		ReplicasUpdated:   updated,
		Containers:        toContainerSpecs(containers),
		Conditions:        conditions,
	}
}

// toDeploymentConditions maps deployment conditions.
func toDeploymentConditions(conditions []appsv1.DeploymentCondition) []ConditionState {
	out := make([]ConditionState, 0, len(conditions))
	for _, condition := range conditions {
		out = append(out, ConditionState{
			Type:    string(condition.Type),
			Status:  string(condition.Status),
			Reason:  condition.Reason,
			Message: condition.Message,
		})
	}

	return out
}

// toJobConditions maps job conditions.
func toJobConditions(conditions []batchv1.JobCondition) []ConditionState {
	out := make([]ConditionState, 0, len(conditions))
	for _, condition := range conditions {
		out = append(out, ConditionState{
			Type:    string(condition.Type),
			Status:  string(condition.Status),
			Reason:  condition.Reason,
			Message: condition.Message,
		})
	}

	return out
}

// toRuleStates maps RBAC policy rules.
func toRuleStates(rules []rbacv1.PolicyRule) []RuleState {
	out := make([]RuleState, 0, len(rules))
	for _, rule := range rules {
		out = append(out, RuleState{
			Groups:    rule.APIGroups,
			Resources: rule.Resources,
			Names:     rule.ResourceNames,
			Verbs:     rule.Verbs,
		})
	}

	return out
}

// toSubjectStates maps RBAC subjects.
func toSubjectStates(subjects []rbacv1.Subject) []SubjectState {
	out := make([]SubjectState, 0, len(subjects))
	for _, subject := range subjects {
		out = append(out, SubjectState{
			Kind:      subject.Kind,
			Name:      subject.Name,
			Namespace: subject.Namespace,
		})
	}

	return out
}
