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

package controller_test

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/zapr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
	"gianlucam76/k8s-cleaner/internal/controller"
)

var _ = Describe("CleanerClient", func() {
	AfterEach(func() {
		cleaners := &appsv1alpha1.CleanerList{}
		Expect(k8sClient.List(context.TODO(), cleaners)).To(Succeed())

		for i := range cleaners.Items {
			cleaner := cleaners.Items[i]
			Expect(k8sClient.Delete(context.TODO(), &cleaner)).To(Succeed())
		}
	})

	It("shouldSchedule return true when current time is past the nextScheduleTime", func() {
		now := time.Now()
		before := now.Add(-time.Second * 30)

		cleaner := &appsv1alpha1.Cleaner{
			ObjectMeta: metav1.ObjectMeta{
				Name: randomString(),
			},
			Status: appsv1alpha1.CleanerStatus{
				NextScheduleTime: &metav1.Time{Time: before},
			},
		}

		// Create a zap logger
		logger, err := zap.NewDevelopment()
		Expect(err).To(BeNil())

		Expect(controller.ShouldSchedule(cleaner, zapr.NewLogger(logger))).To(BeTrue())

		after := now.Add(time.Second * 30)
		cleaner.Status.NextScheduleTime = &metav1.Time{Time: after}

		Expect(controller.ShouldSchedule(cleaner, zapr.NewLogger(logger))).To(BeFalse())
	})

	It("getNextScheduleTime returns the next time cleaner should be scheduled", func() {
		now := time.Now()
		minute := now.Minute() + 1
		if minute == 60 {
			minute = 0
		}

		cleaner := &appsv1alpha1.Cleaner{
			ObjectMeta: metav1.ObjectMeta{
				Name:              randomString(),
				CreationTimestamp: metav1.Time{Time: now},
			},
			Spec: appsv1alpha1.CleanerSpec{
				Schedule: fmt.Sprintf("%d * * * *", minute),
			},
		}

		nextSchedule, err := controller.GetNextScheduleTime(cleaner, now)
		Expect(err).To(BeNil())
		Expect(nextSchedule.Minute()).To(Equal(minute))
	})

	It("addFinalizer adds finalizer", func() {
		cleaner := &appsv1alpha1.Cleaner{
			ObjectMeta: metav1.ObjectMeta{
				Name: randomString(),
			},
			Spec: appsv1alpha1.CleanerSpec{
				MatchingResources: []appsv1alpha1.Resources{
					{
						Kind:      randomString(),
						Group:     randomString(),
						Version:   randomString(),
						Namespace: randomString(),
					},
				},
			},
		}

		Expect(k8sClient.Create(context.TODO(), cleaner)).To(Succeed())

		reconciler := &controller.CleanerReconciler{
			Client: k8sClient,
			Scheme: testEnv.Scheme,
		}

		Expect(controller.AddFinalizer(reconciler, context.TODO(), cleaner, appsv1alpha1.CleanerFinalizer)).To(Succeed())

		currentCleaner := &appsv1alpha1.Cleaner{}
		Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Name: cleaner.Name}, currentCleaner)).To(Succeed())

		Expect(controllerutil.ContainsFinalizer(currentCleaner, appsv1alpha1.CleanerFinalizer)).To(BeTrue())
	})
})
