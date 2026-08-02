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

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
	"gianlucam76/k8s-cleaner/internal/controller/executor"
)

var _ = Describe("CleanerClient", func() {
	const namePrefix = "rollback-"

	nextMinute := func() int {
		minute := time.Now().Minute() + 1
		if minute == 60 {
			minute = 0
		}
		return minute
	}

	It("Rollback recreates a resource removed by a Delete action", Label("FV"), func() {
		ns := namePrefix + randomString()
		namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
		By(fmt.Sprintf("creating namespace %s", ns))
		Expect(k8sClient.Create(context.TODO(), namespace)).To(Succeed())

		key := randomString()
		value := randomString()

		serviceAccount := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      namePrefix + randomString(),
				Namespace: ns,
				Labels:    map[string]string{key: value},
			},
		}
		By(fmt.Sprintf("creating serviceAccount %s", serviceAccount.Name))
		Expect(k8sClient.Create(context.TODO(), serviceAccount)).To(Succeed())

		cleaner := &appsv1alpha1.Cleaner{
			ObjectMeta: metav1.ObjectMeta{Name: randomString()},
			Spec: appsv1alpha1.CleanerSpec{
				ResourcePolicySet: appsv1alpha1.ResourcePolicySet{
					ResourceSelectors: []appsv1alpha1.ResourceSelector{
						{
							Kind:      kindServiceAccount,
							Group:     "",
							Version:   apiVersionV1,
							Namespace: ns,
							Evaluate:  fmt.Sprintf(evaluateServiceAccounts, key, value),
						},
					},
				},
				Action:   appsv1alpha1.ActionDelete,
				Schedule: fmt.Sprintf("%d * * * *", nextMinute()),
				Rollback: &appsv1alpha1.RollbackOptions{Storage: appsv1alpha1.RollbackStorageReport},
				Notifications: []appsv1alpha1.Notification{
					{Name: randomString(), Type: appsv1alpha1.NotificationTypeCleanerReport},
				},
			},
		}
		By(fmt.Sprintf("creating cleaner %s", cleaner.Name))
		Expect(k8sClient.Create(context.TODO(), cleaner)).To(Succeed())

		By("waiting for the serviceAccount to be deleted")
		Eventually(func() bool {
			currentServiceAccount := &corev1.ServiceAccount{}
			err := k8sClient.Get(context.TODO(),
				types.NamespacedName{Namespace: ns, Name: serviceAccount.Name}, currentServiceAccount)
			return apierrors.IsNotFound(err)
		}, timeout, pollingInterval).Should(BeTrue())

		By("waiting for the Report to capture rollback data")
		Eventually(func() bool {
			report := &appsv1alpha1.Report{}
			err := k8sClient.Get(context.TODO(), types.NamespacedName{Name: cleaner.Name}, report)
			if err != nil {
				return false
			}
			for i := range report.Spec.ResourceInfo {
				if report.Spec.ResourceInfo[i].Resource.Name == serviceAccount.Name && len(report.Spec.ResourceInfo[i].FullResource) > 0 {
					return true
				}
			}
			return false
		}, timeout, pollingInterval).Should(BeTrue())

		By("rolling back the last execution")
		results, err := executor.Rollback(context.TODO(), k8sClient, cleaner.Name, logr.Discard())
		Expect(err).To(BeNil())
		Expect(results).To(HaveLen(1))
		Expect(results[0].Success).To(BeTrue())

		By("verifying the serviceAccount was recreated")
		Eventually(func() error {
			currentServiceAccount := &corev1.ServiceAccount{}
			return k8sClient.Get(context.TODO(),
				types.NamespacedName{Namespace: ns, Name: serviceAccount.Name}, currentServiceAccount)
		}, timeout, pollingInterval).Should(Succeed())

		deleteCleaner(cleaner.Name)
		deleteNamespace(ns)
	})

	It("Rollback restores the previous state of a resource updated by a Transform action", Label("FV"), func() {
		ns := namePrefix + randomString()
		namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
		By(fmt.Sprintf("creating namespace %s", ns))
		Expect(k8sClient.Create(context.TODO(), namespace)).To(Succeed())

		key := randomString()
		value := randomString()
		newValue := randomString()

		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      namePrefix + randomString(),
				Namespace: ns,
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{key: value},
				Ports: []corev1.ServicePort{
					{Port: 80, TargetPort: intstr.IntOrString{IntVal: 80}, Name: randomString()},
				},
			},
		}
		By(fmt.Sprintf("creating service %s", service.Name))
		Expect(k8sClient.Create(context.TODO(), service)).To(Succeed())

		cleaner := &appsv1alpha1.Cleaner{
			ObjectMeta: metav1.ObjectMeta{Name: randomString()},
			Spec: appsv1alpha1.CleanerSpec{
				ResourcePolicySet: appsv1alpha1.ResourcePolicySet{
					ResourceSelectors: []appsv1alpha1.ResourceSelector{
						{
							Kind:      kindService,
							Group:     "",
							Version:   apiVersionV1,
							Namespace: ns,
							Evaluate:  fmt.Sprintf(evaluateService, key, value),
						},
					},
				},
				Transform: fmt.Sprintf(tranformService, key, newValue),
				Action:    appsv1alpha1.ActionTransform,
				Schedule:  fmt.Sprintf("%d * * * *", nextMinute()),
				Rollback:  &appsv1alpha1.RollbackOptions{Storage: appsv1alpha1.RollbackStorageReport},
				Notifications: []appsv1alpha1.Notification{
					{Name: randomString(), Type: appsv1alpha1.NotificationTypeCleanerReport},
				},
			},
		}
		By(fmt.Sprintf("creating cleaner %s", cleaner.Name))
		Expect(k8sClient.Create(context.TODO(), cleaner)).To(Succeed())

		By("waiting for the service to be transformed")
		Eventually(func() bool {
			currentService := &corev1.Service{}
			err := k8sClient.Get(context.TODO(),
				types.NamespacedName{Namespace: ns, Name: service.Name}, currentService)
			if err != nil || currentService.Spec.Selector == nil {
				return false
			}
			return currentService.Spec.Selector[key] == newValue
		}, timeout, pollingInterval).Should(BeTrue())

		By("waiting for the Report to capture rollback data")
		Eventually(func() bool {
			report := &appsv1alpha1.Report{}
			err := k8sClient.Get(context.TODO(), types.NamespacedName{Name: cleaner.Name}, report)
			if err != nil {
				return false
			}
			for i := range report.Spec.ResourceInfo {
				if report.Spec.ResourceInfo[i].Resource.Name == service.Name && len(report.Spec.ResourceInfo[i].FullResource) > 0 {
					return true
				}
			}
			return false
		}, timeout, pollingInterval).Should(BeTrue())

		By("rolling back the last execution")
		results, err := executor.Rollback(context.TODO(), k8sClient, cleaner.Name, logr.Discard())
		Expect(err).To(BeNil())
		Expect(results).To(HaveLen(1))
		Expect(results[0].Success).To(BeTrue())

		By("verifying the service selector was restored")
		Eventually(func() bool {
			currentService := &corev1.Service{}
			err := k8sClient.Get(context.TODO(),
				types.NamespacedName{Namespace: ns, Name: service.Name}, currentService)
			if err != nil || currentService.Spec.Selector == nil {
				return false
			}
			return currentService.Spec.Selector[key] == value
		}, timeout, pollingInterval).Should(BeTrue())

		deleteCleaner(cleaner.Name)
		deleteNamespace(ns)
	})

	It("Rollback misconfiguration leaves resources untouched and surfaces an error", Label("FV"), func() {
		ns := namePrefix + randomString()
		namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
		By(fmt.Sprintf("creating namespace %s", ns))
		Expect(k8sClient.Create(context.TODO(), namespace)).To(Succeed())

		key := randomString()
		value := randomString()

		serviceAccount := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      namePrefix + randomString(),
				Namespace: ns,
				Labels:    map[string]string{key: value},
			},
		}
		By(fmt.Sprintf("creating serviceAccount %s", serviceAccount.Name))
		Expect(k8sClient.Create(context.TODO(), serviceAccount)).To(Succeed())

		// Rollback is enabled but no CleanerReport notification is configured:
		// Cleaner must refuse to run rather than delete a resource it can never
		// offer rollback for.
		cleaner := &appsv1alpha1.Cleaner{
			ObjectMeta: metav1.ObjectMeta{Name: randomString()},
			Spec: appsv1alpha1.CleanerSpec{
				ResourcePolicySet: appsv1alpha1.ResourcePolicySet{
					ResourceSelectors: []appsv1alpha1.ResourceSelector{
						{
							Kind:      kindServiceAccount,
							Group:     "",
							Version:   apiVersionV1,
							Namespace: ns,
							Evaluate:  fmt.Sprintf(evaluateServiceAccounts, key, value),
						},
					},
				},
				Action:   appsv1alpha1.ActionDelete,
				Schedule: everyMinuteSchedule,
				Rollback: &appsv1alpha1.RollbackOptions{Storage: appsv1alpha1.RollbackStorageReport},
			},
		}
		By(fmt.Sprintf("creating cleaner %s", cleaner.Name))
		Expect(k8sClient.Create(context.TODO(), cleaner)).To(Succeed())

		By("waiting for Cleaner to report the misconfiguration")
		Eventually(func() bool {
			currentCleaner := &appsv1alpha1.Cleaner{}
			err := k8sClient.Get(context.TODO(), types.NamespacedName{Name: cleaner.Name}, currentCleaner)
			if err != nil {
				return false
			}
			return currentCleaner.Status.FailureMessage != nil
		}, timeout, pollingInterval).Should(BeTrue())

		By("verifying the serviceAccount was never touched")
		Consistently(func() error {
			currentServiceAccount := &corev1.ServiceAccount{}
			return k8sClient.Get(context.TODO(),
				types.NamespacedName{Namespace: ns, Name: serviceAccount.Name}, currentServiceAccount)
		}, 30*time.Second, pollingInterval).Should(Succeed())

		deleteCleaner(cleaner.Name)
		deleteNamespace(ns)
	})
})
