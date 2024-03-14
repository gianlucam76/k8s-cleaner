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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
	"gianlucam76/k8s-cleaner/internal/controller/executor"

	libsveltosv1alpha1 "github.com/projectsveltos/libsveltos/api/v1alpha1"
)

var _ = Describe("Notification", func() {
	It("getWebexInfo get webex information from Secret", func() {
		webexRoomID := randomString()
		webexToken := randomString()

		secretNamespace := randomString()
		secretNs := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: secretNamespace,
			},
		}
		Expect(k8sClient.Create(context.TODO(), secretNs)).To(Succeed())
		Expect(waitForObject(context.TODO(), k8sClient, secretNs)).To(Succeed())

		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      randomString(),
				Namespace: secretNamespace,
			},
			Data: map[string][]byte{
				libsveltosv1alpha1.WebexRoomID: []byte(webexRoomID),
				libsveltosv1alpha1.WebexToken:  []byte(webexToken),
			},
		}

		notification := &appsv1alpha1.Notification{
			Name: randomString(),
			Type: appsv1alpha1.NotificationTypeWebex,
			NotificationRef: &corev1.ObjectReference{
				Kind:       "Secret",
				APIVersion: "v1",
				Namespace:  secret.Namespace,
				Name:       secret.Name,
			},
		}

		Expect(k8sClient.Create(context.TODO(), secret)).To(Succeed())
		Expect(waitForObject(context.TODO(), k8sClient, secret)).To(Succeed())

		webexInfo, err := executor.GetWebexInfo(context.TODO(), notification)
		Expect(err).To(BeNil())
		Expect(webexInfo).ToNot(BeNil())
		Expect(executor.GetWebexRoom(webexInfo)).To(Equal(webexRoomID))
		Expect(executor.GetWebexToken(webexInfo)).To(Equal(webexToken))
	})

	It("getSlackInfo get slack information from Secret", func() {
		slackChannelID := randomString()
		slackToken := randomString()
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      randomString(),
				Namespace: randomString(),
			},
			Data: map[string][]byte{
				libsveltosv1alpha1.SlackChannelID: []byte(slackChannelID),
				libsveltosv1alpha1.SlackToken:     []byte(slackToken),
			},
		}

		notification := &appsv1alpha1.Notification{
			Name: randomString(),
			Type: appsv1alpha1.NotificationTypeSlack,
			NotificationRef: &corev1.ObjectReference{
				Kind:       "Secret",
				APIVersion: "v1",
				Namespace:  secret.Namespace,
				Name:       secret.Name,
			},
		}

		secretNs := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: secret.Namespace,
			},
		}
		Expect(k8sClient.Create(context.TODO(), secretNs)).To(Succeed())
		Expect(waitForObject(context.TODO(), k8sClient, secretNs)).To(Succeed())

		Expect(k8sClient.Create(context.TODO(), secret)).To(Succeed())
		Expect(waitForObject(context.TODO(), k8sClient, secret)).To(Succeed())

		slackInfo, err := executor.GetSlackInfo(context.TODO(), notification)
		Expect(err).To(BeNil())
		Expect(slackInfo).ToNot(BeNil())
		Expect(executor.GetSlackChannelID(slackInfo)).To(Equal(slackChannelID))
		Expect(executor.GetSlackToken(slackInfo)).To(Equal(slackToken))
	})
})
