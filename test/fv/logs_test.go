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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
)

// logMarkerPod returns a Pod whose single container writes marker to stdout
// once, then sleeps -- long enough to stay Running for a ResourceSelector
// scan to fetch its logs.
func logMarkerPod(namePrefix, ns, marker string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: namePrefix + randomString(), Namespace: ns},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "app",
					Image:   "busybox:stable",
					Command: []string{"sh", "-c", fmt.Sprintf("echo %s; sleep 3600", marker)},
				},
			},
		},
	}
}

var _ = Describe("ResourceSelector LogSource evaluate", func() {
	const namePrefix = "logs-"

	It("Delete Action uses the logs global in the Lua evaluate script", Label("FV"), func() {
		ns := namePrefix + randomString()
		By(fmt.Sprintf("creating namespace %s", ns))
		Expect(k8sClient.Create(context.TODO(), &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: ns},
		})).To(Succeed())

		const marker = "OOM_MARKER_STRING"

		// podMatch logs the marker: the Cleaner should delete it.
		podMatch := logMarkerPod(namePrefix, ns, marker)
		By(fmt.Sprintf("creating Pod %s (expects deletion)", podMatch.Name))
		Expect(k8sClient.Create(context.TODO(), podMatch)).To(Succeed())

		// podNoMatch logs something else: the Cleaner should leave it alone.
		podNoMatch := logMarkerPod(namePrefix, ns, "unrelated log line")
		By(fmt.Sprintf("creating Pod %s (expects no deletion)", podNoMatch.Name))
		Expect(k8sClient.Create(context.TODO(), podNoMatch)).To(Succeed())

		for _, pod := range []*corev1.Pod{podMatch, podNoMatch} {
			By(fmt.Sprintf("waiting for Pod %s to be Running", pod.Name))
			Eventually(func() bool {
				current := &corev1.Pod{}
				if err := k8sClient.Get(context.TODO(),
					types.NamespacedName{Namespace: ns, Name: pod.Name}, current); err != nil {
					return false
				}
				return current.Status.Phase == corev1.PodRunning
			}, timeout, pollingInterval).Should(BeTrue())
		}

		minute := time.Now().Minute() + 1
		if minute == 60 {
			minute = 0
		}

		evaluateOnMarker := fmt.Sprintf(`function evaluate(obj)
  hs = {}
  hs.matching = string.find(logs, %q) ~= nil
  return hs
end`, marker)

		cleaner := &appsv1alpha1.Cleaner{
			ObjectMeta: metav1.ObjectMeta{Name: namePrefix + randomString()},
			Spec: appsv1alpha1.CleanerSpec{
				ResourcePolicySet: appsv1alpha1.ResourcePolicySet{
					ResourceSelectors: []appsv1alpha1.ResourceSelector{
						{
							Kind:      kindPod,
							Group:     "",
							Version:   apiVersionV1,
							Namespace: ns,
							LogSource: &appsv1alpha1.LogSource{},
							Evaluate:  evaluateOnMarker,
						},
					},
				},
				Action:   appsv1alpha1.ActionDelete,
				Schedule: fmt.Sprintf("%d * * * *", minute),
			},
		}
		By(fmt.Sprintf("creating Cleaner %s", cleaner.Name))
		Expect(k8sClient.Create(context.TODO(), cleaner)).To(Succeed())

		By(fmt.Sprintf("verifying Pod %s is deleted (log contains marker)", podMatch.Name))
		Eventually(func() bool {
			err := k8sClient.Get(context.TODO(),
				types.NamespacedName{Namespace: ns, Name: podMatch.Name}, &corev1.Pod{})
			return apierrors.IsNotFound(err)
		}, timeout, pollingInterval).Should(BeTrue())

		By(fmt.Sprintf("verifying Pod %s is NOT deleted (log does not contain marker)", podNoMatch.Name))
		Expect(k8sClient.Get(context.TODO(),
			types.NamespacedName{Namespace: ns, Name: podNoMatch.Name},
			&corev1.Pod{})).To(Succeed())

		deleteCleaner(cleaner.Name)
		deleteNamespace(ns)
	})
})
