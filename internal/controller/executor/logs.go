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
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
)

// +kubebuilder:rbac:groups="",resources=pods/log,verbs=get

const (
	defaultLogTailLines = int64(50)
)

// containerLogTails holds container log tails in the two shapes exposed to
// the Evaluate Lua script: Combined (every selected container's tail
// concatenated, headered when more than one container contributed) and
// ByContainer (the same content split per container name).
type containerLogTails struct {
	Combined    string
	ByContainer map[string]string
}

// fetchPodLogs fetches container log tails for pod according to logSource.
// current always reflects each selected container's running/most-recent
// instance; previous is non-nil only when logSource.Previous is set, and
// only carries entries for containers that have actually restarted
// (RestartCount > 0) -- mirroring techsupport's collector/logs.go behavior.
// A fetch failure for an individual container is logged and skipped rather
// than failing the call: a pod with a mix of accessible and inaccessible
// containers should still surface whatever is available.
func fetchPodLogs(ctx context.Context, pod *corev1.Pod, logSource *appsv1alpha1.LogSource,
	logger logr.Logger) (current, previous *containerLogTails, err error) {

	if clientset == nil {
		return nil, nil, fmt.Errorf("kubernetes clientset is not available")
	}

	tailLines := defaultLogTailLines
	if logSource.TailLines != nil {
		tailLines = *logSource.TailLines
	}

	containerNames := selectedContainerNames(pod, logSource.Containers)

	current = &containerLogTails{ByContainer: make(map[string]string, len(containerNames))}
	if logSource.Previous {
		previous = &containerLogTails{ByContainer: make(map[string]string)}
	}

	for _, containerName := range containerNames {
		tail, fetchErr := getContainerLog(ctx, pod, containerName, tailLines, false)
		if fetchErr != nil {
			logger.V(1).Info(fmt.Sprintf("failed to fetch logs for %s/%s container %s: %v",
				pod.Namespace, pod.Name, containerName, fetchErr))
			continue
		}
		current.ByContainer[containerName] = tail

		if previous != nil && hasRestarted(pod, containerName) {
			prevTail, prevErr := getContainerLog(ctx, pod, containerName, tailLines, true)
			if prevErr != nil {
				logger.V(1).Info(fmt.Sprintf("failed to fetch previous logs for %s/%s container %s: %v",
					pod.Namespace, pod.Name, containerName, prevErr))
				continue
			}
			previous.ByContainer[containerName] = prevTail
		}
	}

	current.Combined = combineContainerLogs(current.ByContainer)
	if previous != nil {
		previous.Combined = combineContainerLogs(previous.ByContainer)
	}

	return current, previous, nil
}

// selectedContainerNames returns the container names to fetch logs for:
// every container in pod when want is empty, otherwise the requested names
// that are actually present on pod (unknown names are skipped).
func selectedContainerNames(pod *corev1.Pod, want []string) []string {
	if len(want) == 0 {
		names := make([]string, 0, len(pod.Spec.Containers))
		for i := range pod.Spec.Containers {
			names = append(names, pod.Spec.Containers[i].Name)
		}
		return names
	}

	present := make(map[string]bool, len(pod.Spec.Containers))
	for i := range pod.Spec.Containers {
		present[pod.Spec.Containers[i].Name] = true
	}

	names := make([]string, 0, len(want))
	for _, name := range want {
		if present[name] {
			names = append(names, name)
		}
	}
	return names
}

// hasRestarted reports whether containerName's current instance in pod is
// not its first (RestartCount > 0), meaning a previous instance's log exists.
func hasRestarted(pod *corev1.Pod, containerName string) bool {
	for i := range pod.Status.ContainerStatuses {
		status := &pod.Status.ContainerStatuses[i]
		if status.Name == containerName {
			return status.RestartCount > 0
		}
	}
	return false
}

// getContainerLog fetches the last tailLines lines of containerName's log
// (or its previous instance's log when previous is true).
func getContainerLog(ctx context.Context, pod *corev1.Pod, containerName string,
	tailLines int64, previous bool) (string, error) {

	opts := &corev1.PodLogOptions{
		Container: containerName,
		Previous:  previous,
		TailLines: &tailLines,
	}

	req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, opts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, stream); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// combineContainerLogs concatenates every container's tail into one string,
// headering each block with "==> <container> <==" when there is more than
// one container, so a script can grep the combined text without knowing
// container names, while still being able to tell blocks apart when needed.
func combineContainerLogs(byContainer map[string]string) string {
	if len(byContainer) == 0 {
		return ""
	}
	if len(byContainer) == 1 {
		for _, tail := range byContainer {
			return tail
		}
	}

	// Deterministic order for readable, testable output.
	names := make([]string, 0, len(byContainer))
	for name := range byContainer {
		names = append(names, name)
	}
	sort.Strings(names)

	var buf bytes.Buffer
	for _, name := range names {
		fmt.Fprintf(&buf, "==> %s <==\n%s\n", name, byContainer[name])
	}
	return buf.String()
}
