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
	"fmt"
	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
	"gianlucam76/k8s-cleaner/internal/controller/executor"
	"os"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Consecutive Failures", func() {
	const namespaceEnv = "NAMESPACE"

	var getConfigMapName = func(cleanerName string) string {
		return fmt.Sprintf("cleaner-%s", cleanerName)
	}

	It("getThrottledResources returns list of resources that have failed in the previous run(s)", func() {
		resource1 := &unstructured.Unstructured{}
		resource1.SetKind(randomString())
		resource1.SetUID(types.UID(randomString()))
		const resource1Failures int = 3

		resource2 := &unstructured.Unstructured{}
		resource2.SetKind(randomString())
		resource2.SetUID(types.UID(randomString()))
		const resource2Failures int = 2

		cleanerName := randomString()
		namespace := randomString()
		os.Setenv(namespaceEnv, namespace)

		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}

		Expect(k8sClient.Create(context.TODO(), ns)).To(Succeed())
		Expect(waitForObject(context.TODO(), k8sClient, ns)).To(Succeed())

		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      getConfigMapName(cleanerName),
			},
			Data: map[string]string{
				executor.GetResourceKey(resource1): strconv.Itoa(resource1Failures),
				executor.GetResourceKey(resource2): strconv.Itoa(resource2Failures),
			},
		}

		Expect(k8sClient.Create(context.TODO(), configMap)).To(Succeed())
		Expect(waitForObject(context.TODO(), k8sClient, configMap)).To(Succeed())

		cleaner := &appsv1alpha1.Cleaner{
			ObjectMeta: metav1.ObjectMeta{
				Name: cleanerName,
			},
			Spec: appsv1alpha1.CleanerSpec{
				OccurrenceThreshold: 5,
			},
		}

		throttled, err := executor.GetThrottledResources(context.TODO(), cleaner)
		Expect(err).To(BeNil())
		Expect(throttled).ToNot(BeNil())
		v, ok := throttled[executor.GetResourceKey(resource1)]
		Expect(ok).To(BeTrue())
		Expect(v).To(Equal(resource1Failures))
		v, ok = throttled[executor.GetResourceKey(resource2)]
		Expect(ok).To(BeTrue())
		Expect(v).To(Equal(resource2Failures))
	})

	It("filterResourcesByThreshold returns only resources that have had more consecutive failures than threshold", func() {
		resource1 := &unstructured.Unstructured{}
		resource1.SetKind(randomString())
		resource1.SetUID(types.UID(randomString()))
		const resource1Failures int = 1

		resource2 := &unstructured.Unstructured{}
		resource2.SetKind(randomString())
		resource2.SetUID(types.UID(randomString()))
		const resource2Failures int = 2

		resource3 := &unstructured.Unstructured{}
		resource3.SetKind(randomString())
		resource3.SetUID(types.UID(randomString()))
		const resource3Failures int = 3

		throttledResources := map[string]int{}
		throttledResources[executor.GetResourceKey(resource1)] = resource1Failures
		throttledResources[executor.GetResourceKey(resource2)] = resource2Failures
		throttledResources[executor.GetResourceKey(resource3)] = resource3Failures

		resource4 := &unstructured.Unstructured{}
		resource4.SetKind(randomString())
		resource4.SetUID(types.UID(randomString()))

		results := make([]executor.ResourceResult, 0)
		results = append(results, executor.ResourceResult{
			Resource: resource1,
		})
		results = append(results, executor.ResourceResult{
			Resource: resource2,
		})
		results = append(results, executor.ResourceResult{
			Resource: resource3,
		})
		results = append(results, executor.ResourceResult{
			Resource: resource4,
		})

		// all resources are a match
		threshold := 1
		filtered := executor.FilterResourcesByThreshold(results, throttledResources, threshold)
		Expect(len(filtered)).To(Equal(4))
		Expect(filtered).To(ContainElement(executor.ResourceResult{Resource: resource1}))
		Expect(filtered).To(ContainElement(executor.ResourceResult{Resource: resource2}))
		Expect(filtered).To(ContainElement(executor.ResourceResult{Resource: resource3}))
		Expect(filtered).To(ContainElement(executor.ResourceResult{Resource: resource4}))

		// only resource2 and resource3 are a match
		threshold = 2
		filtered = executor.FilterResourcesByThreshold(results, throttledResources, threshold)
		Expect(len(filtered)).To(Equal(2))
		Expect(filtered).To(ContainElement(executor.ResourceResult{Resource: resource2}))
		Expect(filtered).To(ContainElement(executor.ResourceResult{Resource: resource3}))

		// only resource3 is a match
		threshold = 3
		filtered = executor.FilterResourcesByThreshold(results, throttledResources, threshold)
		Expect(len(filtered)).To(Equal(1))
		Expect(filtered).To(ContainElement(executor.ResourceResult{Resource: resource3}))

		// no resource is a match
		threshold = 4
		filtered = executor.FilterResourcesByThreshold(results, throttledResources, threshold)
		Expect(len(filtered)).To(Equal(0))
	})

	It("updateRegistry updates ConfigMap with consecutive failures per resource", func() {
		resource1 := &unstructured.Unstructured{}
		resource1.SetKind(randomString())
		resource1.SetUID(types.UID(randomString()))

		resource2 := &unstructured.Unstructured{}
		resource2.SetKind(randomString())
		resource2.SetUID(types.UID(randomString()))

		cleanerName := randomString()
		namespace := randomString()
		os.Setenv(namespaceEnv, namespace)

		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}

		Expect(k8sClient.Create(context.TODO(), ns)).To(Succeed())
		Expect(waitForObject(context.TODO(), k8sClient, ns)).To(Succeed())

		cleaner := &appsv1alpha1.Cleaner{
			ObjectMeta: metav1.ObjectMeta{
				Name: cleanerName,
				UID:  types.UID(randomString()),
			},
			Spec: appsv1alpha1.CleanerSpec{
				OccurrenceThreshold: 5,
			},
		}

		results := make([]executor.ResourceResult, 0)
		results = append(results, executor.ResourceResult{
			Resource: resource1,
		})
		results = append(results, executor.ResourceResult{
			Resource: resource2,
		})

		Expect(executor.UpdateRegistry(context.TODO(), cleaner, results, map[string]int{})).To(Succeed())
		configMap := &corev1.ConfigMap{}
		Expect(k8sClient.Get(context.TODO(),
			types.NamespacedName{Namespace: namespace, Name: getConfigMapName(cleanerName)},
			configMap)).To(Succeed())
		Expect(len(configMap.Data)).To(Equal(2))
		Expect(configMap.Data[executor.GetResourceKey(resource1)]).To(Equal(strconv.Itoa(1)))
		Expect(configMap.Data[executor.GetResourceKey(resource2)]).To(Equal(strconv.Itoa(1)))

		oldRegistry := map[string]int{}
		oldRegistry[executor.GetResourceKey(resource1)] = 1

		resource3 := &unstructured.Unstructured{}
		resource3.SetKind(randomString())
		resource3.SetUID(types.UID(randomString()))

		resource4 := &unstructured.Unstructured{}
		resource4.SetKind(randomString())
		resource4.SetUID(types.UID(randomString()))

		// Remove resource2 and add resource3 and resource4
		results = make([]executor.ResourceResult, 0)
		results = append(results, executor.ResourceResult{
			Resource: resource1,
		})
		results = append(results, executor.ResourceResult{
			Resource: resource3,
		})
		results = append(results, executor.ResourceResult{
			Resource: resource4,
		})

		Expect(executor.UpdateRegistry(context.TODO(), cleaner, results, oldRegistry)).To(Succeed())
		Expect(k8sClient.Get(context.TODO(),
			types.NamespacedName{Namespace: namespace, Name: getConfigMapName(cleanerName)},
			configMap)).To(Succeed())
		Expect(len(configMap.Data)).To(Equal(3))
		Expect(configMap.Data[executor.GetResourceKey(resource1)]).To(Equal(strconv.Itoa(2)))
		Expect(configMap.Data[executor.GetResourceKey(resource2)]).To(Equal("")) // resource2 has been removed
		Expect(configMap.Data[executor.GetResourceKey(resource3)]).To(Equal(strconv.Itoa(1)))
		Expect(configMap.Data[executor.GetResourceKey(resource4)]).To(Equal(strconv.Itoa(1)))
	})
})
