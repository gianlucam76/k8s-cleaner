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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	evaluateService = `      function evaluate()
        hs = {}
        hs.matching = false
        if obj.spec.selector ~= nil then
          if obj.spec.selector["%s"] == "%s" then
            hs.matching = true
          end
        end
        return hs
        end`

	tranformService = `      function transform()
        hs = {}
        obj.spec.selector["%s"] = "%s"
        hs.resource = obj
        return hs
        end`
)

var _ = Describe("CleanerClient", func() {
	const namePrefix = "transform-"
	It("Transform Action updates matching resources", Label("FV"), func() {
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
		newValue := randomString()

		service1 := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      namePrefix + randomString(),
				Namespace: ns,
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					key: value,
				},
				Ports: []corev1.ServicePort{
					{
						Port:       80,
						TargetPort: intstr.IntOrString{IntVal: 80},
						Name:       randomString(),
					},
				},
			},
		}
		By(fmt.Sprintf("creating service %s", service1.Name))
		Expect(k8sClient.Create(context.TODO(), service1)).To(Succeed())

		service2 := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      namePrefix + randomString(),
				Namespace: ns,
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					randomString(): randomString(),
				},
				Ports: []corev1.ServicePort{
					{
						Port:       443,
						TargetPort: intstr.IntOrString{IntVal: 443},
						Name:       randomString(),
					},
				},
			},
		}
		By(fmt.Sprintf("creating service %s", service2.Name))
		Expect(k8sClient.Create(context.TODO(), service2)).To(Succeed())

		minute := time.Now().Minute() + 1
		if minute == 60 {
			minute = 0
		}

		// This Cleaner matches Service1 but does not match Service2
		cleaner := &appsv1alpha1.Cleaner{
			ObjectMeta: metav1.ObjectMeta{
				Name: randomString(),
			},
			Spec: appsv1alpha1.CleanerSpec{
				ResourcePolicySet: appsv1alpha1.ResourcePolicySet{
					ResourceSelectors: []appsv1alpha1.ResourceSelector{
						{
							Kind:      "Service",
							Group:     "",
							Version:   "v1",
							Namespace: ns,
							Evaluate:  fmt.Sprintf(evaluateService, key, value),
						},
					},
					Transform: fmt.Sprintf(tranformService, key, newValue),
					Action:    appsv1alpha1.ActionTransform,
				},
				Schedule: fmt.Sprintf("%d * * * *", minute),
			},
		}

		By(fmt.Sprintf("creating cleaner %s", cleaner.Name))
		Expect(k8sClient.Create(context.TODO(), cleaner)).To(Succeed())

		// Cleaner matches Service1. This is then updated
		Eventually(func() bool {
			currentService := &corev1.Service{}
			err := k8sClient.Get(context.TODO(),
				types.NamespacedName{Namespace: ns, Name: service1.Name}, currentService)
			if err != nil {
				return false
			}
			if currentService.Spec.Selector == nil {
				return false
			}
			return currentService.Spec.Selector[key] == newValue
		}, timeout, pollingInterval).Should(BeTrue())

		// Cleaner does not match ServiceAccount2. So this is *not* updated
		currentService := &corev1.Service{}
		Expect(k8sClient.Get(context.TODO(),
			types.NamespacedName{Namespace: ns, Name: service2.Name}, currentService)).To(Succeed())
		Expect(currentService.Spec.Selector).ToNot(BeNil())
		_, ok := currentService.Spec.Selector[key]
		Expect(ok).To(BeFalse())

		deleteCleaner(cleaner.Name)

		deleteNamespace(ns)
	})
})
