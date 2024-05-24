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
	"fmt"

	"github.com/go-logr/logr"
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
			Kind:      "Secret",
			Group:     "",
			Version:   "v1",
			Namespace: secret.Namespace,
		}

		list, err := executor.FetchResources(context.TODO(), matchingResources, logr.Logger{})
		Expect(err).To(BeNil())
		Expect(len(list)).To(Equal(1))
	})

	It("fetchResources gets resources considering all namespaces matching NamespaceSelector", func() {
		key := randomString()
		value := randomString()

		ns1 := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: randomString(),
				Labels: map[string]string{
					key:            value,
					randomString(): randomString(),
				},
			},
		}

		Expect(k8sClient.Create(context.TODO(), ns1)).To(Succeed())
		Expect(waitForObject(context.TODO(), k8sClient, ns1)).To(Succeed())

		secret1 := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns1.Name,
				Name:      randomString(),
			},
		}

		Expect(k8sClient.Create(context.TODO(), secret1)).To(Succeed())
		Expect(waitForObject(context.TODO(), k8sClient, secret1)).To(Succeed())

		ns2 := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: randomString(),
				Labels: map[string]string{
					key:            value,
					randomString(): randomString(),
				},
			},
		}

		Expect(k8sClient.Create(context.TODO(), ns2)).To(Succeed())
		Expect(waitForObject(context.TODO(), k8sClient, ns2)).To(Succeed())

		secret2 := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns2.Name,
				Name:      randomString(),
			},
		}

		Expect(k8sClient.Create(context.TODO(), secret2)).To(Succeed())
		Expect(waitForObject(context.TODO(), k8sClient, secret2)).To(Succeed())

		ns3 := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: randomString(),
				Labels: map[string]string{
					randomString(): randomString(),
				},
			},
		}

		Expect(k8sClient.Create(context.TODO(), ns3)).To(Succeed())
		Expect(waitForObject(context.TODO(), k8sClient, ns3)).To(Succeed())

		secret3 := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns3.Name,
				Name:      randomString(),
			},
		}

		Expect(k8sClient.Create(context.TODO(), secret3)).To(Succeed())
		Expect(waitForObject(context.TODO(), k8sClient, secret3)).To(Succeed())

		matchingResources := &appsv1alpha1.ResourceSelector{
			Kind:              "Secret",
			Group:             "",
			Version:           "v1",
			NamespaceSelector: fmt.Sprintf("%s=%s", key, value),
		}

		list, err := executor.FetchResources(context.TODO(), matchingResources, logr.Logger{})
		Expect(err).To(BeNil())
		Expect(len(list)).To(Equal(2)) // Contains secret1 and secret2 but not secret3
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

		list, err := executor.FetchResources(context.TODO(), matchingResources, logr.Logger{})
		Expect(err).To(BeNil())
		Expect(len(list)).To(Equal(1))
		Expect(list[0].GetName()).To(Equal(secret2.Name))
	})

	It("getNamespaces gets all namespaces matching namespaceSelector", func() {
		key := randomString()
		value := randomString()

		ns1 := corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: randomString(),
				Labels: map[string]string{
					key:            value,
					randomString(): randomString(),
				},
			},
		}

		ns2 := corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: randomString(),
				Labels: map[string]string{
					key:            value,
					randomString(): randomString(),
				},
			},
		}

		ns3 := corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: randomString(),
				Labels: map[string]string{
					randomString(): randomString(),
				},
			},
		}

		ns4 := corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: randomString(),
				Labels: map[string]string{
					key: randomString(),
				},
			},
		}

		Expect(k8sClient.Create(context.TODO(), &ns1)).To(Succeed())
		Expect(waitForObject(context.TODO(), k8sClient, &ns1)).To(Succeed())
		Expect(k8sClient.Create(context.TODO(), &ns2)).To(Succeed())
		Expect(waitForObject(context.TODO(), k8sClient, &ns2)).To(Succeed())
		Expect(k8sClient.Create(context.TODO(), &ns3)).To(Succeed())
		Expect(waitForObject(context.TODO(), k8sClient, &ns3)).To(Succeed())
		Expect(k8sClient.Create(context.TODO(), &ns4)).To(Succeed())
		Expect(waitForObject(context.TODO(), k8sClient, &ns4)).To(Succeed())

		resourceSelector := &appsv1alpha1.ResourceSelector{
			NamespaceSelector: fmt.Sprintf("%s=%s", key, value),
		}

		namespaces, err := executor.GetNamespaces(context.TODO(), resourceSelector, logr.Logger{})
		Expect(err).To(BeNil())
		Expect(len(namespaces)).To(Equal(2))
		Expect(namespaces).To(ContainElement(ns1.Name))
		Expect(namespaces).To(ContainElement(ns2.Name))
	})

	It("getNamespaces adds namespace only once", func() {
		// If resourceSelector.Namespace is defined and the same namespace
		// is also a match for resourceSelector.NamespaceSelector, namespace is present
		// only once
		key := randomString()
		value := randomString()

		ns1 := corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: randomString(),
				Labels: map[string]string{
					key:            value,
					randomString(): randomString(),
				},
			},
		}

		Expect(k8sClient.Create(context.TODO(), &ns1)).To(Succeed())
		Expect(waitForObject(context.TODO(), k8sClient, &ns1)).To(Succeed())

		// ns matches NamespaceSelector and it is also selected by Namespace
		resourceSelector := &appsv1alpha1.ResourceSelector{
			Namespace:         ns1.Name,
			NamespaceSelector: fmt.Sprintf("%s=%s", key, value),
		}

		namespaces, err := executor.GetNamespaces(context.TODO(), resourceSelector, logr.Logger{})
		Expect(err).To(BeNil())
		Expect(len(namespaces)).To(Equal(1))
		Expect(namespaces).To(ContainElement(ns1.Name))
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
