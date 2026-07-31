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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
)

var (
	evaluateSecrets = `      function evaluate()
        hs = {}
        hs.matching = false
        if obj.metadata.labels ~= nil then
          if obj.metadata.labels["%s"] == "%s" then
            hs.matching = true
          end
        end
        return hs
        end`
)

var _ = Describe("CleanerClient", func() {
	const namePrefix = "blast-radius-"
	It("BlastRadiusLimit aborts Delete action when too many resources match", Label("FV"), func() {
		const numSecrets = 20
		const maxCount = 2

		ns := namePrefix + randomString()

		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: ns,
			},
		}
		By(fmt.Sprintf("creating namespace %s", ns))
		Expect(k8sClient.Create(context.TODO(), namespace)).To(Succeed())

		key := randomString()
		value := randomString()

		secretNames := make([]string, 0, numSecrets)
		By(fmt.Sprintf("creating %d secrets in namespace %s", numSecrets, ns))
		for i := 0; i < numSecrets; i++ {
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      namePrefix + randomString(),
					Namespace: ns,
					Labels: map[string]string{
						key: value,
					},
				},
			}
			Expect(k8sClient.Create(context.TODO(), secret)).To(Succeed())
			secretNames = append(secretNames, secret.Name)
		}

		// Runs every minute rather than once at a fixed minute (as other fv tests do)
		// because status.FailureMessage is only refreshed on the *next* Reconcile
		// (the controller uses a GenerationChangedPredicate, so nothing but the
		// cron-driven requeue triggers a reconcile once the Cleaner is created).
		// A recurring schedule keeps that follow-up reconcile within this test's
		// timeout instead of up to an hour away.
		//
		// This Cleaner matches all numSecrets Secrets, but BlastRadiusLimit.MaxCount
		// only allows maxCount resources to be affected. Every run must abort without
		// deleting any Secret.
		cleaner := &appsv1alpha1.Cleaner{
			ObjectMeta: metav1.ObjectMeta{
				Name: randomString(),
			},
			Spec: appsv1alpha1.CleanerSpec{
				ResourcePolicySet: appsv1alpha1.ResourcePolicySet{
					ResourceSelectors: []appsv1alpha1.ResourceSelector{
						{
							Kind:      kindSecret,
							Group:     "",
							Version:   apiVersionV1,
							Namespace: ns,
							Evaluate:  fmt.Sprintf(evaluateSecrets, key, value),
						},
					},
				},
				Action: appsv1alpha1.ActionDelete,
				BlastRadiusLimit: &appsv1alpha1.BlastRadiusLimit{
					MaxCount: ptrInt(maxCount),
				},
				Schedule: "* * * * *",
			},
		}

		By(fmt.Sprintf("creating cleaner %s", cleaner.Name))
		Expect(k8sClient.Create(context.TODO(), cleaner)).To(Succeed())

		By("verifying the Cleaner reports the blast radius limit was exceeded")
		Eventually(func() bool {
			currentCleaner := &appsv1alpha1.Cleaner{}
			err := k8sClient.Get(context.TODO(), types.NamespacedName{Name: cleaner.Name}, currentCleaner)
			if err != nil {
				return false
			}
			return currentCleaner.Status.FailureMessage != nil &&
				*currentCleaner.Status.FailureMessage != ""
		}, timeout, pollingInterval).Should(BeTrue())

		currentCleaner := &appsv1alpha1.Cleaner{}
		Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Name: cleaner.Name}, currentCleaner)).To(Succeed())
		Expect(*currentCleaner.Status.FailureMessage).To(ContainSubstring("blast radius limit exceeded"))
		Expect(*currentCleaner.Status.FailureMessage).To(ContainSubstring(fmt.Sprintf("%d resources matched", numSecrets)))
		Expect(*currentCleaner.Status.FailureMessage).To(ContainSubstring(fmt.Sprintf("max count %d", maxCount)))

		By("verifying none of the matching Secrets were deleted")
		for i := range secretNames {
			currentSecret := &corev1.Secret{}
			Expect(k8sClient.Get(context.TODO(),
				types.NamespacedName{Namespace: ns, Name: secretNames[i]}, currentSecret)).To(Succeed())
		}

		deleteCleaner(cleaner.Name)

		deleteNamespace(ns)
	})
})

func ptrInt(v int) *int {
	return &v
}
