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

	appsv1alpha1 "gianlucam76/k8s-pruner/api/v1alpha1"
	"gianlucam76/k8s-pruner/internal/controller"
)

var _ = Describe("PrunerClient", func() {
	AfterEach(func() {
		pruners := &appsv1alpha1.PrunerList{}
		Expect(k8sClient.List(context.TODO(), pruners)).To(Succeed())

		for i := range pruners.Items {
			pruner := pruners.Items[i]
			Expect(k8sClient.Delete(context.TODO(), &pruner)).To(Succeed())
		}
	})

	It("shouldSchedule return true when current time is past the nextScheduleTime", func() {
		now := time.Now()
		before := now.Add(-time.Second * 30)

		pruner := &appsv1alpha1.Pruner{
			ObjectMeta: metav1.ObjectMeta{
				Name: randomString(),
			},
			Status: appsv1alpha1.PrunerStatus{
				NextScheduleTime: &metav1.Time{Time: before},
			},
		}

		// Create a zap logger
		logger, err := zap.NewDevelopment()
		Expect(err).To(BeNil())

		Expect(controller.ShouldSchedule(pruner, zapr.NewLogger(logger))).To(BeTrue())

		after := now.Add(time.Second * 30)
		pruner.Status.NextScheduleTime = &metav1.Time{Time: after}

		Expect(controller.ShouldSchedule(pruner, zapr.NewLogger(logger))).To(BeFalse())
	})

	It("getNextScheduleTime returns the next time pruner should be scheduled", func() {
		now := time.Now()

		pruner := &appsv1alpha1.Pruner{
			ObjectMeta: metav1.ObjectMeta{
				Name:              randomString(),
				CreationTimestamp: metav1.Time{Time: now},
			},
			Spec: appsv1alpha1.PrunerSpec{
				Schedule: fmt.Sprintf("%d * * * *", now.Minute()+1),
			},
		}

		nextSchedule, err := controller.GetNextScheduleTime(pruner, now)
		Expect(err).To(BeNil())
		Expect(nextSchedule.Minute()).To(Equal(now.Minute() + 1))
	})

	It("addFinalizer adds finalizer", func() {
		pruner := &appsv1alpha1.Pruner{
			ObjectMeta: metav1.ObjectMeta{
				Name: randomString(),
			},
			Spec: appsv1alpha1.PrunerSpec{
				StaleResources: []appsv1alpha1.Resources{
					{
						Kind:      randomString(),
						Group:     randomString(),
						Version:   randomString(),
						Namespace: randomString(),
					},
				},
			},
		}

		Expect(k8sClient.Create(context.TODO(), pruner)).To(Succeed())

		reconciler := &controller.PrunerReconciler{
			Client: k8sClient,
			Scheme: testEnv.Scheme,
		}

		Expect(controller.AddFinalizer(reconciler, context.TODO(), pruner, appsv1alpha1.PrunerFinalizer)).To(Succeed())

		currentPruner := &appsv1alpha1.Pruner{}
		Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Name: pruner.Name}, currentPruner)).To(Succeed())

		Expect(controllerutil.ContainsFinalizer(currentPruner, appsv1alpha1.PrunerFinalizer)).To(BeTrue())
	})
})
