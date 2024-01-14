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

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"

	"github.com/projectsveltos/libsveltos/lib/utils"
)

var (
	podTemplate = `apiVersion: v1
kind: Pod
metadata:
  name: %s
  namespace: %s
spec:
  containers:
  - name: my-container
    image: nginx:latest
    volumeMounts:
    - name: secret-volume
      mountPath: /etc/secret-volume
      readOnly: true
  volumes:
  - name: secret-volume
    secret:
      secretName: %s`

	aggregatedSelection = `function evaluate()
        local hs = {}
        hs.valid = true
        hs.message = ""

        local pods = {}
        local secrets = {}
		local unmountedSecrets = {}

        -- Separate pods and secrets from the resources
        for _, resource in ipairs(resources) do
            local kind = resource.kind
            if kind == "Pod" then
                table.insert(pods, resource)
            elseif kind == "Secret" then
                table.insert(secrets, resource)
            end
        end

		-- Identify secrets not mounted by any pod
		for _, secret in ipairs(secrets) do
			local mountedByPod = false
			for _, pod in ipairs(pods) do
			    if pod.spec.volumes ~= nil then
				  for _, volume in ipairs(pod.spec.volumes) do
				    if volume.secret ~= nil then
                      if volume.secret.secretName == secret.metadata.name then
                        mountedByPod = true
                        break
                      end
					end
                  end	
                end
                if mountedByPod then
                  break
                end
            end
            if not mountedByPod then
              table.insert(unmountedSecrets, {resource = secret})
            end
        end
	
        -- Set the result
        hs.resources = unmountedSecrets
        return hs
    end`
)

var _ = Describe("Aggregated Filtering", func() {
	const namePrefix = "aggregated-"
	It("Delete all Secrets which are not mounted by any Pod", Label("FV"), func() {
		ns := namePrefix + randomString()

		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: ns,
			},
		}
		By(fmt.Sprintf("creating namespace %s", ns))
		Expect(k8sClient.Create(context.TODO(), namespace)).To(Succeed())

		secret1 := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      randomString(),
				Namespace: ns,
			},
			Data: map[string][]byte{},
		}

		By(fmt.Sprintf("creating secret %s/%s", secret1.Namespace, secret1.Name))
		Expect(k8sClient.Create(context.TODO(), secret1)).To(Succeed())

		podName := randomString()
		podYAML := fmt.Sprintf(podTemplate, podName, ns, secret1.Name)
		pod, err := utils.GetUnstructured([]byte(podYAML))
		Expect(err).To(BeNil())

		By(fmt.Sprintf("creating pod %s/%s that mounts secret %s/%s",
			ns, podName, secret1.Namespace, secret1.Name))
		Expect(k8sClient.Create(context.TODO(), pod)).To(Succeed())

		secret2 := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      randomString(),
				Namespace: ns,
			},
			Data: map[string][]byte{},
		}

		By(fmt.Sprintf("creating secret %s/%s which is not mounted by any Pod", secret2.Namespace, secret2.Name))
		Expect(k8sClient.Create(context.TODO(), secret2)).To(Succeed())

		minute := time.Now().Minute() + 1
		if minute == 60 {
			minute = 0
		}

		// This Cleaner matches ServiceAccount1 but does not match ServiceAccount2
		cleaner := &appsv1alpha1.Cleaner{
			ObjectMeta: metav1.ObjectMeta{
				Name: randomString(),
			},
			Spec: appsv1alpha1.CleanerSpec{
				ResourcePolicySet: appsv1alpha1.ResourcePolicySet{
					ResourceSelectors: []appsv1alpha1.ResourceSelector{
						{
							Kind:      "Secret",
							Group:     "",
							Version:   "v1",
							Namespace: ns,
						},
						{
							Kind:      "Pod",
							Group:     "",
							Version:   "v1",
							Namespace: ns,
						},
					},
					AggregatedSelection: aggregatedSelection,
				},
				Action:   appsv1alpha1.ActionDelete,
				Schedule: fmt.Sprintf("%d * * * *", minute),
			},
		}

		By(fmt.Sprintf("creating cleaner %s", cleaner.Name))
		Expect(k8sClient.Create(context.TODO(), cleaner)).To(Succeed())

		// Cleaner matches Secret2. This is then deleted
		Eventually(func() bool {
			currentSecret := &corev1.Secret{}
			err := k8sClient.Get(context.TODO(),
				types.NamespacedName{Namespace: ns, Name: secret2.Name}, currentSecret)
			if err == nil {
				return false
			}
			return apierrors.IsNotFound(err)
		}, timeout, pollingInterval).Should(BeTrue())

		// Cleaner does not match Secret1. So this is *not* deleted
		currentSecret := &corev1.Secret{}
		Expect(k8sClient.Get(context.TODO(),
			types.NamespacedName{Namespace: ns, Name: secret1.Name},
			currentSecret)).To(Succeed())

		deleteCleaner(cleaner.Name)

		deleteNamespace(ns)
	})
})
