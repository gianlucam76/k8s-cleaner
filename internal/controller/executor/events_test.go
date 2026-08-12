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
	"fmt"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"gianlucam76/k8s-cleaner/internal/controller/executor"
)

func serviceAccountResource(sa *corev1.ServiceAccount) *unstructured.Unstructured {
	resource := &unstructured.Unstructured{}
	resource.SetAPIVersion(apiVersionV1)
	resource.SetKind(kindServiceAccount)
	resource.SetNamespace(sa.Namespace)
	resource.SetName(sa.Name)
	resource.SetUID(sa.UID)
	return resource
}

var _ = Describe("fetchEvents", func() {
	var ns *corev1.Namespace
	var sa *corev1.ServiceAccount
	var resource *unstructured.Unstructured

	BeforeEach(func() {
		ns = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: randomString()},
		}
		Expect(k8sClient.Create(context.TODO(), ns)).To(Succeed())

		sa = &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{Namespace: ns.Name, Name: randomString()},
		}
		Expect(k8sClient.Create(context.TODO(), sa)).To(Succeed())

		resource = serviceAccountResource(sa)
	})

	AfterEach(func() {
		Expect(k8sClient.Delete(context.TODO(), ns)).To(Succeed())
	})

	It("returns nil for a cluster-scoped resource (no namespace)", func() {
		clusterScoped := &unstructured.Unstructured{}
		clusterScoped.SetKind("Node")
		clusterScoped.SetName(randomString())

		events, err := executor.FetchEvents(context.TODO(), clusterScoped)
		Expect(err).ToNot(HaveOccurred())
		Expect(events).To(BeNil())
	})

	It("returns only events involving the given resource, most recent first", func() {
		other := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{Namespace: ns.Name, Name: randomString()},
		}
		Expect(k8sClient.Create(context.TODO(), other)).To(Succeed())

		older := &corev1.Event{
			ObjectMeta: metav1.ObjectMeta{Namespace: ns.Name, Name: randomString()},
			InvolvedObject: corev1.ObjectReference{
				Kind: kindServiceAccount, APIVersion: apiVersionV1,
				Namespace: sa.Namespace, Name: sa.Name, UID: sa.UID,
			},
			Reason:        "OlderReason",
			Type:          corev1.EventTypeNormal,
			LastTimestamp: metav1.NewTime(time.Now().Add(-time.Hour)),
		}
		Expect(k8sClient.Create(context.TODO(), older)).To(Succeed())

		newer := &corev1.Event{
			ObjectMeta: metav1.ObjectMeta{Namespace: ns.Name, Name: randomString()},
			InvolvedObject: corev1.ObjectReference{
				Kind: kindServiceAccount, APIVersion: apiVersionV1,
				Namespace: sa.Namespace, Name: sa.Name, UID: sa.UID,
			},
			Reason:        "NewerReason",
			Type:          corev1.EventTypeWarning,
			LastTimestamp: metav1.NewTime(time.Now()),
		}
		Expect(k8sClient.Create(context.TODO(), newer)).To(Succeed())

		unrelated := &corev1.Event{
			ObjectMeta: metav1.ObjectMeta{Namespace: ns.Name, Name: randomString()},
			InvolvedObject: corev1.ObjectReference{
				Kind: kindServiceAccount, APIVersion: apiVersionV1,
				Namespace: other.Namespace, Name: other.Name, UID: other.UID,
			},
			Reason:        "UnrelatedReason",
			Type:          corev1.EventTypeNormal,
			LastTimestamp: metav1.NewTime(time.Now()),
		}
		Expect(k8sClient.Create(context.TODO(), unrelated)).To(Succeed())

		events, err := executor.FetchEvents(context.TODO(), resource)
		Expect(err).ToNot(HaveOccurred())
		Expect(events).To(HaveLen(2))
		Expect(events[0].Reason).To(Equal("NewerReason"))
		Expect(events[1].Reason).To(Equal("OlderReason"))
	})

	It("sees events raised through the events.k8s.io/v1 API", func() {
		// k8s-cleaner's own EventRecorder (client.go) creates events.k8s.io/v1
		// events, as does any component using k8s.io/client-go/tools/events. The
		// apiserver converts between the two representations, so listing
		// core/v1 events must not miss them.
		regarding := corev1.ObjectReference{
			Kind: kindServiceAccount, APIVersion: apiVersionV1,
			Namespace: sa.Namespace, Name: sa.Name, UID: sa.UID,
		}
		modernEvent := &eventsv1.Event{
			ObjectMeta:          metav1.ObjectMeta{Namespace: ns.Name, Name: randomString()},
			EventTime:           metav1.NewMicroTime(time.Now()),
			Regarding:           regarding,
			Reason:              reasonModernAPI,
			Note:                "raised via events.k8s.io/v1",
			Type:                corev1.EventTypeWarning,
			Action:              "Testing",
			ReportingController: "k8s-cleaner-test",
			ReportingInstance:   "test-instance",
		}
		Expect(k8sClient.Create(context.TODO(), modernEvent)).To(Succeed())

		Eventually(func() bool {
			events, err := executor.FetchEvents(context.TODO(), resource)
			if err != nil {
				return false
			}
			for i := range events {
				if events[i].Reason == reasonModernAPI {
					return true
				}
			}
			return false
		}, timeout, pollingInterval).Should(BeTrue())
	})
})

var _ = Describe("fetchEvents (via IsMatch)", func() {
	It("exposes reason/message/type/count/lastTimestamp on the events global", func() {
		resource := serviceAccountResource(&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "n"},
		})

		events := []corev1.Event{
			{
				Reason:        reasonFailedMount,
				Message:       "unable to mount volume",
				Type:          corev1.EventTypeWarning,
				Count:         7,
				LastTimestamp: metav1.NewTime(time.Now()),
			},
		}

		script := fmt.Sprintf(`
function evaluate(obj)
  hs = {}
  hs.matching = false
  if events[1] ~= nil and events[1].reason == %q and events[1].count == 7 then
    hs.matching = true
    hs.message = events[1].message
  end
  return hs
end`, reasonFailedMount)

		matching, message, err := executor.IsMatch(resource, script, nil, events, nil, nil, logr.Logger{})
		Expect(err).ToNot(HaveOccurred())
		Expect(matching).To(BeTrue())
		Expect(message).To(Equal("unable to mount volume"))
	})
})
