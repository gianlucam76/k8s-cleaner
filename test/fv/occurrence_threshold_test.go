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

package fv_test

import (
	"context"
	"fmt"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/projectsveltos/libsveltos/lib/k8s_utils"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Occurrence Threshold:", func() {
	const namePrefix = "threshold-"

	var podTemplate = `apiVersion: v1
kind: Pod
metadata:
  name: stuck-in-limbo
  namespace: %s
spec:
  containers:
  - name: nginx
    image: nginx
  tolerations:
  - key: "dedicated"
    operator: "Equal"
    value: "high-gpu-nodes"
    effect: "NoSchedule"
  nodeSelector:
    disktype: "ssd"`

	It("Matching resources are deleted after occurrenceThreshold is reached", Label("FV"), func() {
		namespace := randomString()

		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}
		Expect(k8sClient.Create(context.TODO(), ns)).To(Succeed())

		pod, err := k8s_utils.GetUnstructured([]byte(fmt.Sprintf(podTemplate, namespace)))
		Expect(err).To(BeNil())
		By(fmt.Sprintf("Creating pod %s/%s. This pod will stay in Pending state", pod.GetNamespace(), pod.GetName()))
		Expect(k8sClient.Create(context.TODO(), pod)).To(Succeed())

		currentPod := &corev1.Pod{}
		Expect(k8sClient.Get(context.TODO(),
			types.NamespacedName{Namespace: pod.GetNamespace(), Name: pod.GetName()},
			currentPod)).To(Succeed())

		var cleanerYAML = `apiVersion: apps.projectsveltos.io/v1alpha1
kind: Cleaner
metadata:
  name: %s-%s
spec:
  schedule: "* * * * *"
  # OccurrenceThreshold is useful here to avoid deleting
  # pods that are only pending for a few seconds.
  occurrenceThreshold: 3
  resourcePolicySet:
    resourceSelectors:
    - kind: Pod
      group: ""
      version: v1
      evaluate: |
        function evaluate()
          hs = {}
          hs.matching = false

          -- Match pods in Pending phase
          if obj.status.phase == "Pending" then
            hs.matching = true
            hs.message = "Pod is stuck in Pending state"
          end

          return hs
        end
  action: Delete`

		cleaner, err := k8s_utils.GetUnstructured([]byte(fmt.Sprintf(cleanerYAML, namePrefix, randomString())))
		Expect(err).To(BeNil())
		By(fmt.Sprintf("Creating cleaner %s", cleaner.GetName()))
		Expect(k8sClient.Create(context.TODO(), cleaner)).To(Succeed())

		// Cleaner matches Pod and creates a ConfigMap
		By("Verifying ConfigMap with matching resource is created and pod is a match resource")
		Eventually(func() bool {
			podKey := fmt.Sprintf("Pod__%s", string(currentPod.GetUID()))
			configMapName := fmt.Sprintf("cleaner-%s", cleaner.GetName())
			currentConfigMap := &corev1.ConfigMap{}
			err := k8sClient.Get(context.TODO(),
				types.NamespacedName{Namespace: "projectsveltos", Name: configMapName}, currentConfigMap)
			if err != nil {
				return false
			}
			if currentConfigMap.Data == nil {
				return false
			}
			_, ok := currentConfigMap.Data[podKey]
			return ok
		}, timeout, pollingInterval).Should(BeTrue())

		By(fmt.Sprintf("Verifying pod %s/%s is still not deleted", currentPod.GetNamespace(), currentPod.GetName()))
		Expect(k8sClient.Get(context.TODO(),
			types.NamespacedName{Namespace: pod.GetNamespace(), Name: pod.GetName()},
			currentPod)).To(Succeed())
		Expect(currentPod.DeletionTimestamp.IsZero()).To(BeTrue())

		// Cleaner matches Pod and creates a ConfigMap
		By("Verifying ConfigMap with matching resource is updated")
		Eventually(func() bool {
			podKey := fmt.Sprintf("Pod__%s", string(currentPod.GetUID()))
			configMapName := fmt.Sprintf("cleaner-%s", cleaner.GetName())
			currentConfigMap := &corev1.ConfigMap{}
			err := k8sClient.Get(context.TODO(),
				types.NamespacedName{Namespace: "projectsveltos", Name: configMapName}, currentConfigMap)
			if err != nil {
				return false
			}
			if currentConfigMap.Data == nil {
				return false
			}
			v, ok := currentConfigMap.Data[podKey]
			if !ok {
				return false
			}
			return v == strconv.Itoa(2)
		}, timeout, pollingInterval).Should(BeTrue())

		By(fmt.Sprintf("Verifying pod %s/%s is deleted", pod.GetNamespace(), pod.GetName()))
		Eventually(func() bool {
			err = k8sClient.Get(context.TODO(),
				types.NamespacedName{Namespace: pod.GetNamespace(), Name: pod.GetName()},
				currentPod)
			if err == nil {
				return false
			}
			return apierrors.IsNotFound(err)
		}, timeout, pollingInterval).Should(BeTrue())

		deleteCleaner(cleaner.GetName())

		By("Verifying ConfigMap is deleted")
		Eventually(func() bool {
			configMapName := fmt.Sprintf("cleaner-%s", cleaner.GetName())
			currentConfigMap := &corev1.ConfigMap{}
			err = k8sClient.Get(context.TODO(),
				types.NamespacedName{Namespace: "projectsveltos", Name: configMapName}, currentConfigMap)
			if err == nil {
				return false
			}
			return apierrors.IsNotFound(err)
		}, timeout, pollingInterval).Should(BeTrue())

	})
})
