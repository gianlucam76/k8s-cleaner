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
	"strings"

	"github.com/atc0005/go-teams-notify/v2/adaptivecard"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
	"gianlucam76/k8s-cleaner/internal/controller/executor"
)

const namespaceTest = "test"

func newResourceInfo(kind, name, message string) appsv1alpha1.ResourceInfo {
	return appsv1alpha1.ResourceInfo{
		Resource: corev1.ObjectReference{
			Kind:       kind,
			APIVersion: apiVersionV1,
			Namespace:  namespaceTest,
			Name:       name,
		},
		Message: message,
	}
}

var _ = Describe("Notification formatting", func() {
	It("resourceRef renders namespace/name, or just name for cluster-scoped resources", func() {
		namespaced := corev1.ObjectReference{Namespace: namespaceTest, Name: "my-cm"}
		Expect(executor.ResourceRef(&namespaced)).To(Equal("test/my-cm"))

		clusterScoped := corev1.ObjectReference{Name: "my-clusterrole"}
		Expect(executor.ResourceRef(&clusterScoped)).To(Equal("my-clusterrole"))
	})

	It("truncateResourceInfo passes short lists through untouched", func() {
		resources := []appsv1alpha1.ResourceInfo{
			newResourceInfo(kindConfigMap, "a", ""),
			newResourceInfo(kindConfigMap, "b", ""),
		}
		shown, omitted := executor.TruncateResourceInfo(resources)
		Expect(shown).To(HaveLen(2))
		Expect(omitted).To(Equal(0))
	})

	It("truncateResourceInfo caps long lists and reports how many were omitted", func() {
		resources := make([]appsv1alpha1.ResourceInfo, executor.MaxNotificationResourceLines+5)
		for i := range resources {
			resources[i] = newResourceInfo(kindConfigMap, randomString(), "")
		}
		shown, omitted := executor.TruncateResourceInfo(resources)
		Expect(shown).To(HaveLen(executor.MaxNotificationResourceLines))
		Expect(omitted).To(Equal(5))
	})

	It("slackColorForAction and discordColorForAction differ by action", func() {
		Expect(executor.SlackColorForAction(appsv1alpha1.ActionDelete)).
			ToNot(Equal(executor.SlackColorForAction(appsv1alpha1.ActionTransform)))
		Expect(executor.DiscordColorForAction(appsv1alpha1.ActionDelete)).
			ToNot(Equal(executor.DiscordColorForAction(appsv1alpha1.ActionScan)))
	})

	It("buildSlackAttachment lists resources instead of dumping raw JSON", func() {
		reportSpec := &appsv1alpha1.ReportSpec{
			Action: appsv1alpha1.ActionDelete,
			ResourceInfo: []appsv1alpha1.ResourceInfo{
				newResourceInfo(kindConfigMap, "old-cm", "orphaned"),
			},
		}

		attachment := executor.BuildSlackAttachment(reportSpec, "unused-configmaps")
		Expect(attachment.Color).To(Equal(executor.SlackColorForAction(appsv1alpha1.ActionDelete)))
		Expect(attachment.Text).To(ContainSubstring("old-cm"))
		Expect(attachment.Text).To(ContainSubstring("orphaned"))
		Expect(attachment.Text).ToNot(ContainSubstring("{\"")) // not raw JSON
		Expect(attachment.AuthorName).To(ContainSubstring("unused-configmaps"))
	})

	It("buildSlackAttachment gives the Cleaner name an action-colored marker", func() {
		reportSpec := &appsv1alpha1.ReportSpec{Action: appsv1alpha1.ActionDelete}
		deleteAttachment := executor.BuildSlackAttachment(reportSpec, "my-cleaner")

		reportSpec.Action = appsv1alpha1.ActionTransform
		transformAttachment := executor.BuildSlackAttachment(reportSpec, "my-cleaner")

		Expect(deleteAttachment.AuthorName).To(ContainSubstring("my-cleaner"))
		Expect(transformAttachment.AuthorName).To(ContainSubstring("my-cleaner"))
		Expect(deleteAttachment.AuthorName).ToNot(Equal(transformAttachment.AuthorName))
	})

	It("buildSlackAttachment truncates long resource lists", func() {
		resources := make([]appsv1alpha1.ResourceInfo, executor.MaxNotificationResourceLines+3)
		for i := range resources {
			resources[i] = newResourceInfo(kindConfigMap, randomString(), "")
		}
		reportSpec := &appsv1alpha1.ReportSpec{Action: appsv1alpha1.ActionDelete, ResourceInfo: resources}

		attachment := executor.BuildSlackAttachment(reportSpec, "unused-configmaps")
		Expect(attachment.Text).To(ContainSubstring("...and 3 more"))
		Expect(strings.Count(attachment.Text, "\n")).To(Equal(executor.MaxNotificationResourceLines))
	})

	It("buildSlackAttachment handles an empty resource list", func() {
		reportSpec := &appsv1alpha1.ReportSpec{Action: appsv1alpha1.ActionScan}
		attachment := executor.BuildSlackAttachment(reportSpec, "unused-configmaps")
		Expect(attachment.Text).To(ContainSubstring("No resources matched"))
	})

	It("buildTeamsCard lists resources instead of dumping raw JSON", func() {
		reportSpec := &appsv1alpha1.ReportSpec{
			Action: appsv1alpha1.ActionTransform,
			ResourceInfo: []appsv1alpha1.ResourceInfo{
				newResourceInfo("Deployment", "my-deploy", "scaled down"),
			},
		}

		card, err := executor.BuildTeamsCard(reportSpec, "This report has been generated by k8s-cleaner")
		Expect(err).To(BeNil())

		rendered := cardText(&card)
		Expect(rendered).To(ContainSubstring("my-deploy"))
		Expect(rendered).To(ContainSubstring("scaled down"))
		Expect(rendered).ToNot(ContainSubstring("{\""))
	})

	It("buildDiscordEmbed lists resources instead of dumping raw JSON", func() {
		reportSpec := &appsv1alpha1.ReportSpec{
			Action: appsv1alpha1.ActionDelete,
			ResourceInfo: []appsv1alpha1.ResourceInfo{
				newResourceInfo(kindSecret, "old-secret", "unused"),
			},
		}

		embed := executor.BuildDiscordEmbed(reportSpec)
		Expect(embed.Color).To(Equal(executor.DiscordColorForAction(appsv1alpha1.ActionDelete)))
		Expect(embed.Description).To(ContainSubstring("old-secret"))
		Expect(embed.Description).To(ContainSubstring("unused"))
		Expect(embed.Description).ToNot(ContainSubstring("{\""))
	})
})

// cardText concatenates every TextBlock's text in a Teams Adaptive Card, for
// assertions against its rendered content.
func cardText(card *adaptivecard.Card) string {
	var sb strings.Builder

	var walk func(elements []adaptivecard.Element)
	walk = func(elements []adaptivecard.Element) {
		for i := range elements {
			sb.WriteString(elements[i].Text)
			sb.WriteString("\n")
			if len(elements[i].Items) > 0 {
				walk(elements[i].Items)
			}
		}
	}
	walk(card.Body)

	return sb.String()
}
