/*
Copyright 2023. projectsveltos.io. All rights reserved.

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

	appsv1alpha1 "gianlucam76/k8s-pruner/api/v1alpha1"
)

var (
	evaluateServiceAccounts = `      function evaluate()
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

var _ = Describe("PrunerClient", func() {
	const namePrefix = "delete-"
	It("Delete Action removes matching resources", Label("FV"), func() {
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

		serviceAccount1 := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      namePrefix + randomString(),
				Namespace: ns,
				Labels: map[string]string{
					key: value,
				},
			},
		}
		By(fmt.Sprintf("creating serviceAccount %s", serviceAccount1.Name))
		Expect(k8sClient.Create(context.TODO(), serviceAccount1)).To(Succeed())

		serviceAccount2 := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      namePrefix + randomString(),
				Namespace: ns,
				Labels: map[string]string{
					key: randomString(),
				},
			},
		}
		By(fmt.Sprintf("creating serviceAccount %s", serviceAccount2.Name))
		Expect(k8sClient.Create(context.TODO(), serviceAccount2)).To(Succeed())

		// This Pruner matches ServiceAccount1 but does not match ServiceAccount2
		pruner := &appsv1alpha1.Pruner{
			ObjectMeta: metav1.ObjectMeta{
				Name: randomString(),
			},
			Spec: appsv1alpha1.PrunerSpec{
				StaleResources: []appsv1alpha1.Resources{
					{
						Kind:      "ServiceAccount",
						Group:     "",
						Version:   "v1",
						Namespace: ns,
						Evaluate:  fmt.Sprintf(evaluateServiceAccounts, key, value),
						Action:    appsv1alpha1.ActionDelete,
					},
				},
				Schedule: fmt.Sprintf("%d * * * *", time.Now().Minute()+1),
			},
		}

		By(fmt.Sprintf("creating pruner %s", pruner.Name))
		Expect(k8sClient.Create(context.TODO(), pruner)).To(Succeed())

		// Pruner matches ServiceAccount1. This is then deleted
		Eventually(func() bool {
			currentServiceAccount := &corev1.ServiceAccount{}
			err := k8sClient.Get(context.TODO(),
				types.NamespacedName{Namespace: ns, Name: serviceAccount1.Name}, currentServiceAccount)
			if err == nil {
				return false
			}
			return apierrors.IsNotFound(err)
		}, timeout, pollingInterval).Should(BeTrue())

		// Pruner does not match ServiceAccount2. So this is *not* deleted
		currentServiceAccount := &corev1.ServiceAccount{}
		Expect(k8sClient.Get(context.TODO(),
			types.NamespacedName{Namespace: ns, Name: serviceAccount2.Name},
			currentServiceAccount)).To(Succeed())

		deletePruner(pruner.Name)
	})
})
