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

package executor_test

import (
	"context"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
	"gianlucam76/k8s-cleaner/internal/controller/executor"
)

const (
	containerApp     = "app"
	containerSidecar = "sidecar"
	kindPod          = "Pod"

	logOOM       = "OutOfMemoryError\n"
	logOK        = "ok\n"
	logRunning   = "running\n"
	logOOMKilled = "OOMKilled\n"
)

func podWithContainers(names ...string) *corev1.Pod {
	containers := make([]corev1.Container, len(names))
	for i, name := range names {
		containers[i] = corev1.Container{Name: name}
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "pod-" + randomString()},
		Spec:       corev1.PodSpec{Containers: containers},
	}
}

func podResource(namespace, name string) *unstructured.Unstructured {
	resource := &unstructured.Unstructured{}
	resource.SetAPIVersion(apiVersionV1)
	resource.SetKind(kindPod)
	resource.SetNamespace(namespace)
	resource.SetName(name)
	return resource
}

var _ = Describe("selectedContainerNames", func() {
	It("returns every container name when want is empty", func() {
		pod := podWithContainers(containerApp, containerSidecar)
		Expect(executor.SelectedContainerNames(pod, nil)).To(ConsistOf(containerApp, containerSidecar))
	})

	It("returns only the requested names that are present on the pod", func() {
		pod := podWithContainers(containerApp, containerSidecar)
		names := executor.SelectedContainerNames(pod, []string{containerSidecar, "does-not-exist"})
		Expect(names).To(ConsistOf(containerSidecar))
	})

	It("returns an empty slice when none of the requested names are present", func() {
		pod := podWithContainers(containerApp)
		Expect(executor.SelectedContainerNames(pod, []string{"missing"})).To(BeEmpty())
	})
})

var _ = Describe("hasRestarted", func() {
	It("returns true when the named container's RestartCount is greater than zero", func() {
		pod := podWithContainers(containerApp)
		pod.Status.ContainerStatuses = []corev1.ContainerStatus{
			{Name: containerApp, RestartCount: 3},
		}
		Expect(executor.HasRestarted(pod, containerApp)).To(BeTrue())
	})

	It("returns false when the named container's RestartCount is zero", func() {
		pod := podWithContainers(containerApp)
		pod.Status.ContainerStatuses = []corev1.ContainerStatus{
			{Name: containerApp, RestartCount: 0},
		}
		Expect(executor.HasRestarted(pod, containerApp)).To(BeFalse())
	})

	It("returns false when the container has no status yet", func() {
		pod := podWithContainers(containerApp)
		Expect(executor.HasRestarted(pod, containerApp)).To(BeFalse())
	})
})

var _ = Describe("combineContainerLogs", func() {
	It("returns an empty string for an empty map", func() {
		Expect(executor.CombineContainerLogs(map[string]string{})).To(BeEmpty())
	})

	It("returns the tail unmodified, with no header, for a single container", func() {
		combined := executor.CombineContainerLogs(map[string]string{containerApp: "line1\nline2\n"})
		Expect(combined).To(Equal("line1\nline2\n"))
	})

	It("headers each block with its container name, in sorted order, for multiple containers", func() {
		combined := executor.CombineContainerLogs(map[string]string{
			containerSidecar: "sidecar log\n",
			containerApp:     "app log\n",
		})
		Expect(combined).To(Equal("==> app <==\napp log\n\n==> sidecar <==\nsidecar log\n\n"))
	})
})

var _ = Describe("fetchPodLogs", func() {
	It("skips a container whose log cannot be fetched rather than failing the whole call", func() {
		// pod was never created in the cluster (envtest has no kubelet to serve
		// container logs anyway), so the underlying GetLogs call fails for "app".
		// fetchPodLogs must swallow that per-container and return an empty tail
		// rather than propagating the error, so one unreachable container doesn't
		// stop the Evaluate script from seeing logs for the others.
		pod := podWithContainers(containerApp)
		current, previous, err := executor.FetchPodLogs(context.TODO(), pod, &appsv1alpha1.LogSource{}, logr.Logger{})
		Expect(err).ToNot(HaveOccurred())
		Expect(previous).To(BeNil())
		Expect(current).ToNot(BeNil())
		Expect(current.ByContainer).To(BeEmpty())
		Expect(current.Combined).To(BeEmpty())
	})
})

var _ = Describe("logs/logsByContainer (via IsMatch)", func() {
	resource := podResource("ns", "n")

	It("exposes the combined tail on logs without needing a container name", func() {
		current := &executor.ContainerLogTails{
			Combined:    "==> app <==\n" + logOOM + "\n==> sidecar <==\n" + logOK + "\n",
			ByContainer: map[string]string{containerApp: logOOM, containerSidecar: logOK},
		}

		script := `
function evaluate(obj)
  hs = {}
  hs.matching = string.find(logs, "OutOfMemoryError") ~= nil
  return hs
end`

		matching, _, err := executor.IsMatch(resource, script, nil, nil, current, nil, logr.Logger{})
		Expect(err).ToNot(HaveOccurred())
		Expect(matching).To(BeTrue())
	})

	It("exposes each container's tail individually on logsByContainer", func() {
		current := &executor.ContainerLogTails{
			ByContainer: map[string]string{containerApp: logOOM, containerSidecar: logOK},
		}

		script := `
function evaluate(obj)
  hs = {}
  hs.matching = logsByContainer["sidecar"] == "ok\n" and
                string.find(logsByContainer["app"], "OutOfMemoryError") ~= nil
  return hs
end`

		matching, _, err := executor.IsMatch(resource, script, nil, nil, current, nil, logr.Logger{})
		Expect(err).ToNot(HaveOccurred())
		Expect(matching).To(BeTrue())
	})

	It("leaves logs/logsByContainer empty, not nil, when LogSource was not set", func() {
		script := `
function evaluate(obj)
  hs = {}
  hs.matching = (logs == "") and (logsByContainer["app"] == nil)
  return hs
end`

		matching, _, err := executor.IsMatch(resource, script, nil, nil, nil, nil, logr.Logger{})
		Expect(err).ToNot(HaveOccurred())
		Expect(matching).To(BeTrue())
	})

	It("populates logsPrevious/logsPreviousByContainer separately from logs/logsByContainer", func() {
		current := &executor.ContainerLogTails{Combined: logRunning, ByContainer: map[string]string{containerApp: logRunning}}
		previous := &executor.ContainerLogTails{Combined: logOOMKilled, ByContainer: map[string]string{containerApp: logOOMKilled}}

		script := `
function evaluate(obj)
  hs = {}
  hs.matching = string.find(logs, "running") ~= nil and
                string.find(logsPrevious, "OOMKilled") ~= nil
  return hs
end`

		matching, _, err := executor.IsMatch(resource, script, nil, nil, current, previous, logr.Logger{})
		Expect(err).ToNot(HaveOccurred())
		Expect(matching).To(BeTrue())
	})
})
