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

var _ = Describe("ResourceSelector IncludeEvents evaluate", func() {
	const namePrefix = "events-"
	const reasonFailedMount = "FailedMount"

	It("Delete Action uses the events global in the Lua evaluate script", Label("FV"), func() {
		ns := namePrefix + randomString()
		By(fmt.Sprintf("creating namespace %s", ns))
		Expect(k8sClient.Create(context.TODO(), &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: ns},
		})).To(Succeed())

		// saMatch has a FailedMount Event recorded against it: the Cleaner should delete it.
		saMatch := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{Name: namePrefix + randomString(), Namespace: ns},
		}
		By(fmt.Sprintf("creating ServiceAccount %s (expects deletion)", saMatch.Name))
		Expect(k8sClient.Create(context.TODO(), saMatch)).To(Succeed())

		By(fmt.Sprintf("recording a %s Event against %s", reasonFailedMount, saMatch.Name))
		Expect(k8sClient.Create(context.TODO(), &corev1.Event{
			ObjectMeta: metav1.ObjectMeta{Name: namePrefix + randomString(), Namespace: ns},
			InvolvedObject: corev1.ObjectReference{
				Kind: kindServiceAccount, APIVersion: apiVersionV1,
				Namespace: saMatch.Namespace, Name: saMatch.Name, UID: saMatch.UID,
			},
			Reason:        reasonFailedMount,
			Type:          corev1.EventTypeWarning,
			LastTimestamp: metav1.NewTime(time.Now()),
		})).To(Succeed())

		// saNoEvent has no Event recorded: the Cleaner should leave it alone.
		saNoEvent := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{Name: namePrefix + randomString(), Namespace: ns},
		}
		By(fmt.Sprintf("creating ServiceAccount %s (expects no deletion)", saNoEvent.Name))
		Expect(k8sClient.Create(context.TODO(), saNoEvent)).To(Succeed())

		minute := time.Now().Minute() + 1
		if minute == 60 {
			minute = 0
		}

		evaluateOnFailedMount := fmt.Sprintf(`function evaluate(obj)
  hs = {}
  hs.matching = false
  for _, e in ipairs(events) do
    if e.reason == %q then
      hs.matching = true
    end
  end
  return hs
end`, reasonFailedMount)

		cleaner := &appsv1alpha1.Cleaner{
			ObjectMeta: metav1.ObjectMeta{Name: namePrefix + randomString()},
			Spec: appsv1alpha1.CleanerSpec{
				ResourcePolicySet: appsv1alpha1.ResourcePolicySet{
					ResourceSelectors: []appsv1alpha1.ResourceSelector{
						{
							Kind:          kindServiceAccount,
							Group:         "",
							Version:       apiVersionV1,
							Namespace:     ns,
							IncludeEvents: true,
							Evaluate:      evaluateOnFailedMount,
						},
					},
				},
				Action:   appsv1alpha1.ActionDelete,
				Schedule: fmt.Sprintf("%d * * * *", minute),
			},
		}
		By(fmt.Sprintf("creating Cleaner %s", cleaner.Name))
		Expect(k8sClient.Create(context.TODO(), cleaner)).To(Succeed())

		By(fmt.Sprintf("verifying ServiceAccount %s is deleted (FailedMount event present)", saMatch.Name))
		Eventually(func() bool {
			err := k8sClient.Get(context.TODO(),
				types.NamespacedName{Namespace: ns, Name: saMatch.Name}, &corev1.ServiceAccount{})
			return apierrors.IsNotFound(err)
		}, timeout, pollingInterval).Should(BeTrue())

		By(fmt.Sprintf("verifying ServiceAccount %s is NOT deleted (no matching event)", saNoEvent.Name))
		Expect(k8sClient.Get(context.TODO(),
			types.NamespacedName{Namespace: ns, Name: saNoEvent.Name},
			&corev1.ServiceAccount{})).To(Succeed())

		deleteCleaner(cleaner.Name)
		deleteNamespace(ns)
	})
})
