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
	"encoding/json"
	"strings"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
	"gianlucam76/k8s-cleaner/internal/controller/executor"
)

func newRollbackCleaner(name string, action appsv1alpha1.Action,
	rollback *appsv1alpha1.RollbackOptions, notifications ...appsv1alpha1.Notification) *appsv1alpha1.Cleaner {

	return &appsv1alpha1.Cleaner{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appsv1alpha1.CleanerSpec{
			Action:        action,
			Rollback:      rollback,
			Notifications: notifications,
		},
	}
}

func newConfigMapResourceResult(namespace, name string, data map[string]string) executor.ResourceResult {
	u := &unstructured.Unstructured{}
	u.SetAPIVersion(apiVersionV1)
	u.SetKind(kindConfigMap)
	u.SetNamespace(namespace)
	u.SetName(name)
	if data != nil {
		Expect(unstructured.SetNestedStringMap(u.Object, data, "data")).To(Succeed())
	}
	return executor.ResourceResult{Resource: u}
}

const kindConfigMap = "ConfigMap"

var _ = Describe("Rollback capture", func() {
	It("addRollbackResourceData populates FullResource only when Rollback is enabled", func() {
		resources := []executor.ResourceResult{newConfigMapResourceResult(randomString(), randomString(), nil)}

		withoutRollback := newRollbackCleaner(randomString(), appsv1alpha1.ActionDelete, nil)
		reportSpec := executor.GenerateReportSpec(resources, withoutRollback)
		reportSpec = executor.AddRollbackResourceData(reportSpec, resources, withoutRollback, logr.Discard())
		Expect(reportSpec.ResourceInfo[0].FullResource).To(BeEmpty())

		withRollback := newRollbackCleaner(randomString(), appsv1alpha1.ActionDelete,
			&appsv1alpha1.RollbackOptions{Storage: appsv1alpha1.RollbackStorageReport})
		reportSpec = executor.GenerateReportSpec(resources, withRollback)
		reportSpec = executor.AddRollbackResourceData(reportSpec, resources, withRollback, logr.Discard())
		Expect(reportSpec.ResourceInfo[0].FullResource).ToNot(BeEmpty())

		var captured unstructured.Unstructured
		Expect(json.Unmarshal(reportSpec.ResourceInfo[0].FullResource, &captured.Object)).To(Succeed())
		Expect(captured.GetName()).To(Equal(resources[0].Resource.GetName()))
	})

	It("addRollbackResourceData never populates FullResource for a Scan action", func() {
		resources := []executor.ResourceResult{newConfigMapResourceResult(randomString(), randomString(), nil)}
		cleaner := newRollbackCleaner(randomString(), appsv1alpha1.ActionScan,
			&appsv1alpha1.RollbackOptions{Storage: appsv1alpha1.RollbackStorageReport})

		reportSpec := executor.GenerateReportSpec(resources, cleaner)
		reportSpec = executor.AddRollbackResourceData(reportSpec, resources, cleaner, logr.Discard())
		Expect(reportSpec.ResourceInfo[0].FullResource).To(BeEmpty())
	})

	It("addRollbackResourceData skips resources larger than the size cap", func() {
		hugeData := map[string]string{"blob": strings.Repeat("a", executor.MaxRollbackResourceSize)}
		resources := []executor.ResourceResult{newConfigMapResourceResult(randomString(), randomString(), hugeData)}
		cleaner := newRollbackCleaner(randomString(), appsv1alpha1.ActionDelete,
			&appsv1alpha1.RollbackOptions{Storage: appsv1alpha1.RollbackStorageReport})

		reportSpec := executor.GenerateReportSpec(resources, cleaner)
		reportSpec = executor.AddRollbackResourceData(reportSpec, resources, cleaner, logr.Discard())
		Expect(reportSpec.ResourceInfo[0].FullResource).To(BeEmpty())
		Expect(reportSpec.ResourceInfo[0].Message).To(ContainSubstring("rollback data not stored"))
	})

	It("validateRollbackConfig requires a CleanerReport notification when Rollback is enabled", func() {
		withRollbackNoNotification := newRollbackCleaner(randomString(), appsv1alpha1.ActionDelete,
			&appsv1alpha1.RollbackOptions{Storage: appsv1alpha1.RollbackStorageReport})
		Expect(executor.ValidateRollbackConfig(withRollbackNoNotification)).ToNot(Succeed())

		withRollbackAndNotification := newRollbackCleaner(randomString(), appsv1alpha1.ActionDelete,
			&appsv1alpha1.RollbackOptions{Storage: appsv1alpha1.RollbackStorageReport},
			appsv1alpha1.Notification{Name: randomString(), Type: appsv1alpha1.NotificationTypeCleanerReport})
		Expect(executor.ValidateRollbackConfig(withRollbackAndNotification)).To(Succeed())

		withoutRollback := newRollbackCleaner(randomString(), appsv1alpha1.ActionDelete, nil)
		Expect(executor.ValidateRollbackConfig(withoutRollback)).To(Succeed())
	})

	It("persistRollbackSnapshot creates a Report carrying FullResource", func() {
		cleanerName := randomString()
		cleaner := newRollbackCleaner(cleanerName, appsv1alpha1.ActionDelete,
			&appsv1alpha1.RollbackOptions{Storage: appsv1alpha1.RollbackStorageReport},
			appsv1alpha1.Notification{Name: randomString(), Type: appsv1alpha1.NotificationTypeCleanerReport})

		resources := []executor.ResourceResult{
			newConfigMapResourceResult(randomString(), randomString(), map[string]string{"k": "v"}),
		}

		Expect(executor.PersistRollbackSnapshot(context.TODO(), cleaner, resources, logr.Discard())).To(Succeed())

		report := &appsv1alpha1.Report{}
		Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Name: cleanerName}, report)).To(Succeed())
		Expect(report.Spec.ResourceInfo).To(HaveLen(1))
		Expect(report.Spec.ResourceInfo[0].FullResource).ToNot(BeEmpty())
	})
})

var _ = Describe("Rollback", func() {
	var ns *corev1.Namespace

	BeforeEach(func() {
		ns = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: randomString()},
		}
		Expect(k8sClient.Create(context.TODO(), ns)).To(Succeed())
		Expect(waitForObject(context.TODO(), k8sClient, ns)).To(Succeed())
	})

	AfterEach(func() {
		Expect(k8sClient.Delete(context.TODO(), ns)).To(Succeed())
	})

	newReport := func(name string, action appsv1alpha1.Action, resourceInfo []appsv1alpha1.ResourceInfo) *appsv1alpha1.Report {
		return &appsv1alpha1.Report{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Spec: appsv1alpha1.ReportSpec{
				Action:       action,
				ResourceInfo: resourceInfo,
			},
		}
	}

	It("recreates a resource deleted by a Delete action", func() {
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Namespace: ns.Name, Name: randomString()},
			Data:       map[string]string{"k": "v"},
		}
		Expect(k8sClient.Create(context.TODO(), cm)).To(Succeed())
		Expect(waitForObject(context.TODO(), k8sClient, cm)).To(Succeed())

		fullResource, err := json.Marshal(newConfigMapResourceResult(cm.Namespace, cm.Name, cm.Data).Resource)
		Expect(err).To(BeNil())

		cleanerName := randomString()
		report := newReport(cleanerName, appsv1alpha1.ActionDelete, []appsv1alpha1.ResourceInfo{
			{
				Resource: corev1.ObjectReference{
					Kind: kindConfigMap, APIVersion: apiVersionV1, Namespace: cm.Namespace, Name: cm.Name,
				},
				FullResource: fullResource,
			},
		})
		Expect(k8sClient.Create(context.TODO(), report)).To(Succeed())

		Expect(k8sClient.Delete(context.TODO(), cm)).To(Succeed())

		results, err := executor.Rollback(context.TODO(), k8sClient, cleanerName, logr.Discard())
		Expect(err).To(BeNil())
		Expect(results).To(HaveLen(1))
		Expect(results[0].Success).To(BeTrue())

		recreated := &corev1.ConfigMap{}
		Expect(k8sClient.Get(context.TODO(),
			types.NamespacedName{Namespace: cm.Namespace, Name: cm.Name}, recreated)).To(Succeed())
		Expect(recreated.Data).To(Equal(cm.Data))

		Expect(k8sClient.Delete(context.TODO(), report)).To(Succeed())
	})

	It("restores the previous state of a resource modified by a Transform action", func() {
		originalData := map[string]string{"k": "original"}
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Namespace: ns.Name, Name: randomString()},
			Data:       originalData,
		}
		Expect(k8sClient.Create(context.TODO(), cm)).To(Succeed())
		Expect(waitForObject(context.TODO(), k8sClient, cm)).To(Succeed())

		fullResource, err := json.Marshal(newConfigMapResourceResult(cm.Namespace, cm.Name, originalData).Resource)
		Expect(err).To(BeNil())

		cleanerName := randomString()
		report := newReport(cleanerName, appsv1alpha1.ActionTransform, []appsv1alpha1.ResourceInfo{
			{
				Resource: corev1.ObjectReference{
					Kind: kindConfigMap, APIVersion: apiVersionV1, Namespace: cm.Namespace, Name: cm.Name,
				},
				FullResource: fullResource,
			},
		})
		Expect(k8sClient.Create(context.TODO(), report)).To(Succeed())

		currentCm := &corev1.ConfigMap{}
		Expect(k8sClient.Get(context.TODO(),
			types.NamespacedName{Namespace: cm.Namespace, Name: cm.Name}, currentCm)).To(Succeed())
		currentCm.Data = map[string]string{"k": "transformed"}
		Expect(k8sClient.Update(context.TODO(), currentCm)).To(Succeed())

		results, err := executor.Rollback(context.TODO(), k8sClient, cleanerName, logr.Discard())
		Expect(err).To(BeNil())
		Expect(results).To(HaveLen(1))
		Expect(results[0].Success).To(BeTrue())

		restored := &corev1.ConfigMap{}
		Expect(k8sClient.Get(context.TODO(),
			types.NamespacedName{Namespace: cm.Namespace, Name: cm.Name}, restored)).To(Succeed())
		Expect(restored.Data).To(Equal(originalData))

		Expect(k8sClient.Delete(context.TODO(), report)).To(Succeed())
	})

	It("reports a per-resource failure when no rollback data was captured", func() {
		cleanerName := randomString()
		report := newReport(cleanerName, appsv1alpha1.ActionDelete, []appsv1alpha1.ResourceInfo{
			{Resource: corev1.ObjectReference{Kind: kindConfigMap, APIVersion: apiVersionV1, Name: randomString()}},
		})
		Expect(k8sClient.Create(context.TODO(), report)).To(Succeed())

		results, err := executor.Rollback(context.TODO(), k8sClient, cleanerName, logr.Discard())
		Expect(err).To(BeNil())
		Expect(results).To(HaveLen(1))
		Expect(results[0].Success).To(BeFalse())
		Expect(results[0].Message).To(ContainSubstring("no rollback data"))

		Expect(k8sClient.Delete(context.TODO(), report)).To(Succeed())
	})

	It("errors when the report's last action was Scan", func() {
		cleanerName := randomString()
		report := newReport(cleanerName, appsv1alpha1.ActionScan, []appsv1alpha1.ResourceInfo{})
		Expect(k8sClient.Create(context.TODO(), report)).To(Succeed())

		_, err := executor.Rollback(context.TODO(), k8sClient, cleanerName, logr.Discard())
		Expect(err).ToNot(BeNil())

		Expect(k8sClient.Delete(context.TODO(), report)).To(Succeed())
	})

	It("errors when no report exists for the cleaner", func() {
		_, err := executor.Rollback(context.TODO(), k8sClient, randomString(), logr.Discard())
		Expect(err).ToNot(BeNil())
		Expect(apierrors.IsNotFound(err)).To(BeTrue())
	})
})
