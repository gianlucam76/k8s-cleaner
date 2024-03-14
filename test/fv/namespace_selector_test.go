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
	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("NamespaceSelector:", func() {
	const namePrefix = "ns-selector-"

	It("Consider resources in namespaces matching NamespaceSelector", Label("FV"), func() {
		key1 := randomString()
		value1 := randomString()
		key2 := randomString()
		value2 := randomString()

		namespace1 := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namePrefix + randomString(),
				Labels: map[string]string{
					key1: value1,
					key2: value2,
				},
			},
		}

		Expect(k8sClient.Create(context.TODO(), namespace1)).To(Succeed())

		By("Creating a secret in a namespace matching NamespaceSelector")
		secret1 := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace1.Name,
				Name:      randomString(),
			},
		}
		Expect(k8sClient.Create(context.TODO(), secret1)).To(Succeed())

		namespace2 := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namePrefix + randomString(),
				Labels: map[string]string{
					key1: value1,
					key2: value2,
				},
			},
		}

		Expect(k8sClient.Create(context.TODO(), namespace2)).To(Succeed())

		By("Creating a secret in a namespace matching NamespaceSelector")
		secret2 := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace2.Name,
				Name:      randomString(),
			},
		}
		Expect(k8sClient.Create(context.TODO(), secret2)).To(Succeed())

		namespace3 := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namePrefix + randomString(),
				Labels: map[string]string{
					randomString(): randomString(),
					key1:           value1,
				},
			},
		}

		Expect(k8sClient.Create(context.TODO(), namespace3)).To(Succeed())

		By("Creating a secret in a namespace non matching NamespaceSelector")
		secret3 := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace3.Name,
				Name:      randomString(),
			},
		}
		Expect(k8sClient.Create(context.TODO(), secret3)).To(Succeed())

		minute := time.Now().Minute() + 1
		if minute == 60 {
			minute = 0
		}
		// This Cleaner matches Secrets and uses NamespaceSelector
		cleaner := &appsv1alpha1.Cleaner{
			ObjectMeta: metav1.ObjectMeta{
				Name: randomString(),
			},
			Spec: appsv1alpha1.CleanerSpec{
				ResourcePolicySet: appsv1alpha1.ResourcePolicySet{
					ResourceSelectors: []appsv1alpha1.ResourceSelector{
						{
							Kind:              "Secret",
							Group:             "",
							Version:           "v1",
							NamespaceSelector: fmt.Sprintf("%s=%s, %s=%s", key1, value1, key2, value2),
						},
					},
				},
				Action:   appsv1alpha1.ActionDelete,
				Schedule: fmt.Sprintf("%d * * * *", minute),
			},
		}

		By(fmt.Sprintf("creating cleaner %s", cleaner.Name))
		Expect(k8sClient.Create(context.TODO(), cleaner)).To(Succeed())

		// Cleaner matches Secret1. This is then deleted
		Eventually(func() bool {
			currentSecret := &corev1.Secret{}
			err := k8sClient.Get(context.TODO(),
				types.NamespacedName{Namespace: secret1.Namespace, Name: secret1.Name}, currentSecret)
			if err == nil {
				return false
			}
			return apierrors.IsNotFound(err)
		}, timeout, pollingInterval).Should(BeTrue())

		// Cleaner matches Secret2. This is then deleted
		Eventually(func() bool {
			currentSecret := &corev1.Secret{}
			err := k8sClient.Get(context.TODO(),
				types.NamespacedName{Namespace: secret2.Namespace, Name: secret2.Name}, currentSecret)
			if err == nil {
				return false
			}
			return apierrors.IsNotFound(err)
		}, timeout, pollingInterval).Should(BeTrue())

		// Cleaner does matches Secret3. This is then not deleted
		currentSecret := &corev1.Secret{}
		Expect(k8sClient.Get(context.TODO(),
			types.NamespacedName{Namespace: secret3.Namespace, Name: secret3.Name}, currentSecret)).To(Succeed())
	})
})
