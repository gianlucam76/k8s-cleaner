/*
Copyright 2026. projectsveltos.io. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package executor

import (
	"context"
	"fmt"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch

// fetchEvents returns the Events involving resource, most recent first.
// Cluster-scoped resources (empty namespace) yield no events, since core
// Events are always namespace-scoped and cannot involve a cluster-scoped object.
// Events raised through the newer events.k8s.io/v1 API (e.g. by any component
// using k8s.io/client-go/tools/events, as this project itself does) are
// visible here too: the apiserver converts between the two representations,
// so listing core/v1 events does not miss them.
func fetchEvents(ctx context.Context, resource *unstructured.Unstructured) ([]corev1.Event, error) {
	namespace := resource.GetNamespace()
	if namespace == "" {
		return nil, nil
	}

	eventList := &corev1.EventList{}
	if err := k8sClient.List(ctx, eventList, client.InNamespace(namespace)); err != nil {
		return nil, fmt.Errorf("failed to list events in namespace %s: %w", namespace, err)
	}

	kind := resource.GetKind()
	name := resource.GetName()
	uid := resource.GetUID()

	matched := make([]corev1.Event, 0)
	for i := range eventList.Items {
		involved := eventList.Items[i].InvolvedObject
		if involved.Kind == kind && involved.Name == name && (uid == "" || involved.UID == uid) {
			matched = append(matched, eventList.Items[i])
		}
	}

	sort.Slice(matched, func(i, j int) bool {
		return eventTimestamp(&matched[i]).After(eventTimestamp(&matched[j]))
	})

	return matched, nil
}

// eventTimestamp returns the most relevant timestamp for ev, preferring
// EventTime (the events.k8s.io v1 field, sub-second precision) and falling
// back to the older LastTimestamp field.
func eventTimestamp(ev *corev1.Event) time.Time {
	if !ev.EventTime.IsZero() {
		return ev.EventTime.Time
	}
	return ev.LastTimestamp.Time
}
