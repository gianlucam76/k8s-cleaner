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

package executor_test

import (
	"context"

	"github.com/go-logr/zapr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
	"gianlucam76/k8s-cleaner/internal/controller/executor"

	libsveltosv1alpha1 "github.com/projectsveltos/libsveltos/api/v1alpha1"
)

var _ = Describe("Worker", func() {
	var ns *corev1.Namespace

	BeforeEach(func() {
		ns = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: randomString(),
			},
		}

		Expect(k8sClient.Create(context.TODO(), ns)).To(Succeed())
	})

	AfterEach(func() {
		if ns != nil {
			currentNamespace := &corev1.Namespace{}
			Expect(k8sClient.Get(context.TODO(),
				types.NamespacedName{Name: ns.Name}, currentNamespace)).To(Succeed())
			Expect(k8sClient.Delete(context.TODO(), currentNamespace)).To(Succeed())
		}
	})

	It("fetchResources gets all resources", func() {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      randomString(),
			},
		}

		Expect(k8sClient.Create(context.TODO(), secret)).To(Succeed())
		matchingResources := &appsv1alpha1.ResourceSelector{
			Kind:    "Secret",
			Group:   "",
			Version: "v1",
		}

		list, err := executor.FetchResources(context.TODO(), matchingResources)
		Expect(err).To(BeNil())
		Expect(len(list.Items)).To(Equal(1))
	})

	It("fetchResources gets all resources with proper labels", func() {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      randomString(),
				Labels: map[string]string{
					randomString(): randomString(),
				},
			},
		}
		Expect(k8sClient.Create(context.TODO(), secret)).To(Succeed())

		key := randomString()
		value := randomString()
		secret2 := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      randomString(),
				Labels: map[string]string{
					key: value,
				},
			},
		}

		Expect(k8sClient.Create(context.TODO(), secret2)).To(Succeed())
		matchingResources := &appsv1alpha1.ResourceSelector{
			Kind:    "Secret",
			Group:   "",
			Version: "v1",
			LabelFilters: []libsveltosv1alpha1.LabelFilter{
				{
					Key:       key,
					Operation: libsveltosv1alpha1.OperationEqual,
					Value:     value,
				},
			},
		}

		list, err := executor.FetchResources(context.TODO(), matchingResources)
		Expect(err).To(BeNil())
		Expect(len(list.Items)).To(Equal(1))
		Expect(list.Items[0].GetName()).To(Equal(secret2.Name))
	})

	It("getMatchingResources gets stale resources", func() {
		value := randomString()
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      randomString(),
				Labels: map[string]string{
					"foo": value,
				},
			},
		}

		Expect(k8sClient.Create(context.TODO(), secret)).To(Succeed())

		secret2 := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      randomString(),
				Labels: map[string]string{
					randomString(): randomString(),
				},
			},
		}

		Expect(k8sClient.Create(context.TODO(), secret2)).To(Succeed())

		secretWithLabel := `function evaluate()
   local hs = {}
   hs.matching = false
   if obj.metadata.labels ~= nil then
     for key, value in pairs(obj.metadata.labels) do
	   if key == "foo" then
	     hs.matching = true
	     break
	   end
	 end
   end
   return hs
   end
   `

		matchingResources := &appsv1alpha1.ResourceSelector{
			Kind:     "Secret",
			Group:    "",
			Version:  "v1",
			Evaluate: secretWithLabel,
		}
		logger, err := zap.NewDevelopment()
		Expect(err).To(BeNil())
		resources, err := executor.GetMatchingResources(context.TODO(), matchingResources,
			zapr.NewLogger(logger))
		Expect(err).To(BeNil())
		Expect(resources).ToNot(BeNil())
		Expect(len(resources)).To(Equal(1))
		Expect(resources[0].Resource.GetName()).To(Equal(secret.Name))
	})
})
