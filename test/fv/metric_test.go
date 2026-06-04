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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
)

// mockPrometheusNginxConf is an nginx configuration that serves a static
// Prometheus-compatible scalar response (value=1) for any incoming request,
// regardless of query parameters. This is sufficient to exercise the
// MetricSource/MetricQueries flow without deploying a real Prometheus server.
const mockPrometheusNginxConf = `
server {
    listen 80;
    location / {
        add_header Content-Type application/json;
        return 200 '{"status":"success","data":{"resultType":"scalar","result":[0,"1"]}}';
    }
}
`

var _ = Describe("ResourceSelector metric-based evaluate", func() {
	const namePrefix = "metric-"

	It("Delete Action uses metric value from MetricSource in Lua evaluate script",
		Label("FV"), func() {

			ns := namePrefix + randomString()
			By(fmt.Sprintf("creating namespace %s", ns))
			Expect(k8sClient.Create(context.TODO(), &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{Name: ns},
			})).To(Succeed())

			// ── Mock Prometheus: nginx pod that returns a fixed scalar value of 1 ──

			cmName := namePrefix + randomString()
			By(fmt.Sprintf("creating nginx ConfigMap %s", cmName))
			Expect(k8sClient.Create(context.TODO(), &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: cmName, Namespace: ns},
				Data:       map[string]string{"default.conf": mockPrometheusNginxConf},
			})).To(Succeed())

			podName := namePrefix + randomString()
			By(fmt.Sprintf("creating nginx Pod %s (mock Prometheus endpoint)", podName))
			Expect(k8sClient.Create(context.TODO(), &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: ns,
					Labels:    map[string]string{"app": podName},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:alpine",
							Ports: []corev1.ContainerPort{{ContainerPort: 80}},
							VolumeMounts: []corev1.VolumeMount{
								{Name: "conf", MountPath: "/etc/nginx/conf.d"},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "conf",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: cmName},
								},
							},
						},
					},
				},
			})).To(Succeed())

			svcName := namePrefix + randomString()
			By(fmt.Sprintf("creating Service %s for mock Prometheus", svcName))
			Expect(k8sClient.Create(context.TODO(), &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{Name: svcName, Namespace: ns},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{"app": podName},
					Ports:    []corev1.ServicePort{{Port: 80, Protocol: corev1.ProtocolTCP}},
				},
			})).To(Succeed())

			By(fmt.Sprintf("waiting for Pod %s to be Running", podName))
			Eventually(func() bool {
				pod := &corev1.Pod{}
				if err := k8sClient.Get(context.TODO(),
					types.NamespacedName{Namespace: ns, Name: podName}, pod); err != nil {
					return false
				}
				return pod.Status.Phase == corev1.PodRunning
			}, timeout, pollingInterval).Should(BeTrue())

			// ── Test resources ────────────────────────────────────────────────────

			key := randomString()
			value := randomString()

			// saMatch has the label: Cleaner should delete it (metric "up" == 1, label matches).
			saMatch := &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      namePrefix + randomString(),
					Namespace: ns,
					Labels:    map[string]string{key: value},
				},
			}
			By(fmt.Sprintf("creating ServiceAccount %s (expects deletion)", saMatch.Name))
			Expect(k8sClient.Create(context.TODO(), saMatch)).To(Succeed())

			// saNoLabel does not have the label: Cleaner should leave it alone.
			saNoLabel := &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      namePrefix + randomString(),
					Namespace: ns,
				},
			}
			By(fmt.Sprintf("creating ServiceAccount %s (expects no deletion)", saNoLabel.Name))
			Expect(k8sClient.Create(context.TODO(), saNoLabel)).To(Succeed())

			// ── Cleaner ───────────────────────────────────────────────────────────

			minute := time.Now().Minute() + 1
			if minute == 60 {
				minute = 0
			}

			prometheusURL := fmt.Sprintf("http://%s.%s.svc", svcName, ns)

			// The Lua evaluate function uses metrics["up"] (fetched from MetricSource) AND
			// the resource label. With the mock endpoint returning scalar 1, only ServiceAccounts
			// carrying the expected label will be matched and deleted.
			evaluateWithMetric := fmt.Sprintf(`
function evaluate()
  hs = {}
  hs.matching = false
  if metrics["up"] == nil or metrics["up"] < 1 then
    return hs
  end
  if obj.metadata.labels ~= nil and obj.metadata.labels["%s"] == "%s" then
    hs.matching = true
  end
  return hs
end`, key, value)

			cleaner := &appsv1alpha1.Cleaner{
				ObjectMeta: metav1.ObjectMeta{
					Name: namePrefix + randomString(),
				},
				Spec: appsv1alpha1.CleanerSpec{
					ResourcePolicySet: appsv1alpha1.ResourcePolicySet{
						ResourceSelectors: []appsv1alpha1.ResourceSelector{
							{
								Kind:      "ServiceAccount",
								Group:     "",
								Version:   "v1",
								Namespace: ns,
								Evaluate:  evaluateWithMetric,
								MetricSource: &appsv1alpha1.MetricSource{
									URL: prometheusURL,
								},
								MetricQueries: []appsv1alpha1.MetricQuery{
									{
										Name:  "up",
										Query: "up",
									},
								},
							},
						},
					},
					Action:   appsv1alpha1.ActionDelete,
					Schedule: fmt.Sprintf("%d * * * *", minute),
				},
			}

			By(fmt.Sprintf("creating Cleaner %s", cleaner.Name))
			Expect(k8sClient.Create(context.TODO(), cleaner)).To(Succeed())

			// ── Assertions ────────────────────────────────────────────────────────

			By(fmt.Sprintf("verifying ServiceAccount %s is deleted (metric > 0, label matches)", saMatch.Name))
			Eventually(func() bool {
				err := k8sClient.Get(context.TODO(),
					types.NamespacedName{Namespace: ns, Name: saMatch.Name}, &corev1.ServiceAccount{})
				return apierrors.IsNotFound(err)
			}, timeout, pollingInterval).Should(BeTrue())

			By(fmt.Sprintf("verifying ServiceAccount %s is NOT deleted (no matching label)", saNoLabel.Name))
			Expect(k8sClient.Get(context.TODO(),
				types.NamespacedName{Namespace: ns, Name: saNoLabel.Name},
				&corev1.ServiceAccount{})).To(Succeed())

			// ── Cleanup ───────────────────────────────────────────────────────────
			deleteCleaner(cleaner.Name)
			deleteNamespace(ns)
		})
})
