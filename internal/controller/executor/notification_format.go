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

package executor

import (
	"fmt"
	"strings"

	"github.com/atc0005/go-teams-notify/v2/adaptivecard"
	"github.com/bwmarrin/discordgo"
	"github.com/slack-go/slack"
	corev1 "k8s.io/api/core/v1"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
)

const (
	// maxNotificationResourceLines caps how many resources are listed in a
	// formatted Slack/Teams/Discord notification, so a run matching hundreds
	// of resources doesn't turn the message into an unreadable wall of text.
	maxNotificationResourceLines = 20

	// notificationFooter is shown in the footer of formatted notifications.
	notificationFooter = "k8s-cleaner"

	// Discord embed colors, matching the Slack attachment colors below.
	discordColorDelete    = 0xE01E5A
	discordColorTransform = 0xECB22E
	discordColorScan      = 0x2EB67D
	discordColorDefault   = 0x36C5F0
)

// truncateResourceInfo returns the resources to render in a formatted
// notification, and how many were left out.
func truncateResourceInfo(resources []appsv1alpha1.ResourceInfo) (shown []appsv1alpha1.ResourceInfo, omitted int) {
	if len(resources) <= maxNotificationResourceLines {
		return resources, 0
	}
	return resources[:maxNotificationResourceLines], len(resources) - maxNotificationResourceLines
}

// resourceRef renders a resource reference as "namespace/name", or just
// "name" for cluster-scoped resources.
func resourceRef(ref *corev1.ObjectReference) string {
	if ref.Namespace != "" {
		return fmt.Sprintf("%s/%s", ref.Namespace, ref.Name)
	}
	return ref.Name
}

// reportSummary is the one-line "what happened" summary shown at the top of
// a formatted notification.
func reportSummary(reportSpec *appsv1alpha1.ReportSpec) string {
	return fmt.Sprintf("%s — %d resource(s)", reportSpec.Action, len(reportSpec.ResourceInfo))
}

// --- Slack ---

func slackColorForAction(action appsv1alpha1.Action) string {
	switch action {
	case appsv1alpha1.ActionDelete:
		return "#e01e5a"
	case appsv1alpha1.ActionTransform:
		return "#ecb22e"
	case appsv1alpha1.ActionScan:
		return "#2eb67d"
	default:
		return "#36c5f0"
	}
}

// slackDotForAction is a colored-dot emoji matching slackColorForAction, used
// to give the Cleaner name in AuthorName an actual color marker: Slack has no
// support for coloring arbitrary message text.
func slackDotForAction(action appsv1alpha1.Action) string {
	switch action {
	case appsv1alpha1.ActionDelete:
		return "🔴"
	case appsv1alpha1.ActionTransform:
		return "🟡"
	case appsv1alpha1.ActionScan:
		return "🟢"
	default:
		return "🔵"
	}
}

// buildSlackAttachment renders a ReportSpec as a color-coded Slack
// attachment, listing each resource on its own line, instead of dumping the
// raw ReportSpec JSON. cleanerName is shown as the attachment's author line,
// set apart from the resource list below it.
func buildSlackAttachment(reportSpec *appsv1alpha1.ReportSpec, cleanerName string) slack.Attachment {
	shown, omitted := truncateResourceInfo(reportSpec.ResourceInfo)

	lines := make([]string, 0, len(shown)+1)
	for i := range shown {
		line := fmt.Sprintf("• *%s* `%s`", shown[i].Resource.Kind, resourceRef(&shown[i].Resource))
		if shown[i].Message != "" {
			line += fmt.Sprintf(" — %s", shown[i].Message)
		}
		lines = append(lines, line)
	}
	if omitted > 0 {
		lines = append(lines, fmt.Sprintf("_...and %d more_", omitted))
	}
	if len(lines) == 0 {
		lines = append(lines, "_No resources matched._")
	}

	return slack.Attachment{
		Color:      slackColorForAction(reportSpec.Action),
		AuthorName: fmt.Sprintf("%s %s", slackDotForAction(reportSpec.Action), cleanerName),
		Title:      reportSummary(reportSpec),
		Text:       strings.Join(lines, "\n"),
		MarkdownIn: []string{"text"},
		Footer:     notificationFooter,
	}
}

// --- Teams ---

func teamsContainerStyleForAction(action appsv1alpha1.Action) string {
	switch action {
	case appsv1alpha1.ActionDelete:
		return adaptivecard.ContainerStyleAttention
	case appsv1alpha1.ActionTransform:
		return adaptivecard.ContainerStyleWarning
	case appsv1alpha1.ActionScan:
		return adaptivecard.ContainerStyleGood
	default:
		return adaptivecard.ContainerStyleDefault
	}
}

// buildTeamsCard renders a ReportSpec as a color-coded Adaptive Card,
// listing each resource on its own line, instead of dumping the raw
// ReportSpec JSON.
func buildTeamsCard(reportSpec *appsv1alpha1.ReportSpec, message string) (adaptivecard.Card, error) {
	card := adaptivecard.NewCard()

	if err := card.AddElement(false, adaptivecard.NewTitleTextBlock(message, true)); err != nil {
		return card, err
	}
	if err := card.AddElement(false, adaptivecard.NewTextBlock(reportSummary(reportSpec), true)); err != nil {
		return card, err
	}

	container := adaptivecard.NewContainer()
	container.Style = teamsContainerStyleForAction(reportSpec.Action)

	shown, omitted := truncateResourceInfo(reportSpec.ResourceInfo)
	for i := range shown {
		text := fmt.Sprintf("**%s** %s", shown[i].Resource.Kind, resourceRef(&shown[i].Resource))
		if shown[i].Message != "" {
			text += fmt.Sprintf(" — %s", shown[i].Message)
		}
		container.Items = append(container.Items, adaptivecard.NewTextBlock(text, true))
	}
	if omitted > 0 {
		container.Items = append(container.Items,
			adaptivecard.NewTextBlock(fmt.Sprintf("...and %d more", omitted), true))
	}
	if len(shown) == 0 {
		container.Items = append(container.Items, adaptivecard.NewTextBlock("No resources matched.", true))
	}

	if err := card.AddContainer(false, container); err != nil {
		return card, err
	}

	return card, nil
}

// --- Discord ---

func discordColorForAction(action appsv1alpha1.Action) int {
	switch action {
	case appsv1alpha1.ActionDelete:
		return discordColorDelete
	case appsv1alpha1.ActionTransform:
		return discordColorTransform
	case appsv1alpha1.ActionScan:
		return discordColorScan
	default:
		return discordColorDefault
	}
}

// buildDiscordEmbed renders a ReportSpec as a color-coded Discord embed,
// listing each resource on its own line, instead of uploading the raw
// ReportSpec JSON as a file attachment.
func buildDiscordEmbed(reportSpec *appsv1alpha1.ReportSpec) *discordgo.MessageEmbed {
	shown, omitted := truncateResourceInfo(reportSpec.ResourceInfo)

	lines := make([]string, 0, len(shown)+1)
	for i := range shown {
		line := fmt.Sprintf("**%s** %s", shown[i].Resource.Kind, resourceRef(&shown[i].Resource))
		if shown[i].Message != "" {
			line += fmt.Sprintf(" — %s", shown[i].Message)
		}
		lines = append(lines, line)
	}
	if omitted > 0 {
		lines = append(lines, fmt.Sprintf("...and %d more", omitted))
	}
	if len(lines) == 0 {
		lines = append(lines, "No resources matched.")
	}

	return &discordgo.MessageEmbed{
		Title:       reportSummary(reportSpec),
		Description: strings.Join(lines, "\n"),
		Color:       discordColorForAction(reportSpec.Action),
		Footer:      &discordgo.MessageEmbedFooter{Text: notificationFooter},
	}
}
