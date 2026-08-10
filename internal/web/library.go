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

package web

import (
	"embed"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	"sigs.k8s.io/yaml"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
	unhealthyresources "gianlucam76/k8s-cleaner/examples-unhealthy-resources"
	unusedresources "gianlucam76/k8s-cleaner/examples-unused-resources"
)

// Library categories. These match the source directories the recipes are
// curated from (examples-unused-resources, examples-unhealthy-resources).
const (
	LibraryCategoryUnused    = "unused-resources"
	LibraryCategoryUnhealthy = "unhealthy-resources"
)

// Groups cluster related recipes together within a category so they render
// as adjacent, labeled sections in the dashboard (e.g. RBAC keeps ClusterRoles
// and Roles next to each other) instead of one flat, unordered list.
const (
	LibraryGroupRBAC            = "RBAC"
	LibraryGroupConfigSecrets   = "Config & Secrets" //nolint:gosec // resource kind name, not a credential
	LibraryGroupWorkloads       = "Workloads"
	LibraryGroupNetworking      = "Networking"
	LibraryGroupStorage         = "Storage"
	LibraryGroupCluster         = "Cluster"
	LibraryGroupTimeBased       = "Time-Based"
	LibraryGroupPods            = "Pods"
	LibraryGroupBrokenReference = "Broken References"
)

// libraryManifestEntry describes one curated Cleaner recipe available in the
// dashboard's library. ID doubles as the recipe's suggested Cleaner name.
//
// This manifest is a deliberately curated subset of examples-unused-resources
// and examples-unhealthy-resources: every entry here is self-contained (no
// namespace/label/threshold hardcoded to one cluster) and has a unique
// default name. Recipes that need hand-editing before they're useful, or
// that duplicate another recipe, are intentionally left out.
//
// Entries are listed in display order: within a Category, all entries
// sharing a Group are kept contiguous (see the "contiguous" test in
// library_test.go) so the dashboard can render one labeled section per group
// without re-sorting.
type libraryManifestEntry struct {
	ID          string
	Category    string
	Group       string
	Title       string
	Description string
	sourcePath  string // path within the category's embedded FS
}

const (
	// unusedClusterRolesRecipeID is the canonical ID of the "Unused ClusterRoles"
	// recipe, named here so tests can reference it directly instead of
	// duplicating the literal.
	unusedClusterRolesRecipeID = "unused-clusterroles"
)

var (
	libraryManifest = []libraryManifestEntry{
		// --- unused-resources / RBAC ---
		{
			ID: unusedClusterRolesRecipeID, Category: LibraryCategoryUnused, Group: LibraryGroupRBAC,
			sourcePath:  "clusterroles/unused_clusterroles.yaml",
			Title:       "Unused ClusterRoles",
			Description: "Finds ClusterRole instances not referenced by any ClusterRoleBinding or RoleBinding.",
		},
		{
			ID: "unused-roles", Category: LibraryCategoryUnused, Group: LibraryGroupRBAC,
			sourcePath:  "roles/unused_roles.yaml",
			Title:       "Unused Roles",
			Description: "Finds Role instances, across all namespaces, not referenced by any RoleBinding.",
		},
		{
			ID: "unused-service-accounts", Category: LibraryCategoryUnused, Group: LibraryGroupRBAC,
			sourcePath:  "service-accounts/unused_service-accounts.yaml",
			Title:       "Unused ServiceAccounts",
			Description: "Finds ServiceAccounts used by no Pod and referenced by no RoleBinding or ClusterRoleBinding.",
		},

		// --- unused-resources / Config & Secrets ---
		{
			ID: "unused-configmaps", Category: LibraryCategoryUnused, Group: LibraryGroupConfigSecrets,
			sourcePath:  "configmaps/orphaned_configmaps.yaml",
			Title:       "Orphaned ConfigMaps",
			Description: "Finds ConfigMaps not used by any Pod through volumes, environment variables, or envFrom.",
		},
		{
			ID: "unused-secrets", Category: LibraryCategoryUnused, Group: LibraryGroupConfigSecrets,
			sourcePath:  "secrets/orphaned_secrets.yaml",
			Title:       "Orphaned Secrets",
			Description: "Finds Secrets not used by Pods (volumes, env, image pulls), Ingress TLS, or ServiceAccounts.",
		},

		// --- unused-resources / Workloads ---
		{
			ID: "deployment-with-no-autoscaler", Category: LibraryCategoryUnused, Group: LibraryGroupWorkloads,
			sourcePath:  "deployments/deployment_with_no_autoscaler.yaml",
			Title:       "Deployments Without an Autoscaler",
			Description: "Finds Deployments, in any namespace, with no associated HorizontalPodAutoscaler.",
		},
		{
			ID: "deployment-with-zero-replicas", Category: LibraryCategoryUnused, Group: LibraryGroupWorkloads,
			sourcePath:  "deployments/deployment_with_replica_zero.yaml",
			Title:       "Deployments Scaled to Zero",
			Description: "Finds Deployments, in any namespace, with spec.replicas set to 0.",
		},
		{
			ID: "orphaned-deployments", Category: LibraryCategoryUnused, Group: LibraryGroupWorkloads,
			sourcePath:  "deployments/orphaned_deployment.yaml",
			Title:       "Orphaned Deployments",
			Description: "Finds Deployments with no Pods or Services associated with them.",
		},
		{
			ID: "unused-horizontal-pod-autoscalers", Category: LibraryCategoryUnused, Group: LibraryGroupWorkloads,
			sourcePath:  "horizontal-pod-autoscalers/unused-hpas.yaml",
			Title:       "Unused HorizontalPodAutoscalers",
			Description: "Finds HorizontalPodAutoscaler instances matching no Deployment or StatefulSet.",
		},
		{
			ID: "statefulset-with-no-autoscaler", Category: LibraryCategoryUnused, Group: LibraryGroupWorkloads,
			sourcePath:  "stateful-sets/statefulset_with_no_autoscaler.yaml",
			Title:       "StatefulSets Without an Autoscaler",
			Description: "Finds StatefulSets, in any namespace, with no associated HorizontalPodAutoscaler.",
		},
		{
			ID: "statefulset-with-zero-replicas", Category: LibraryCategoryUnused, Group: LibraryGroupWorkloads,
			sourcePath:  "stateful-sets/statefulset_with_no_replicas.yaml",
			Title:       "StatefulSets Scaled to Zero",
			Description: "Finds StatefulSets, in any namespace, with spec.replicas set to 0.",
		},
		{
			ID: "stale-pod-disruption-budgets", Category: LibraryCategoryUnused, Group: LibraryGroupWorkloads,
			sourcePath:  "pod-disruption-budgets/unused_pod-disruption-budgets.yaml",
			Title:       "Stale PodDisruptionBudgets",
			Description: "Finds PodDisruptionBudgets that match no Deployment or StatefulSet.",
		},
		{
			ID: "completed-jobs", Category: LibraryCategoryUnused, Group: LibraryGroupWorkloads,
			sourcePath:  "jobs/completed_jobs.yaml",
			Title:       "Completed Jobs",
			Description: "Finds Jobs that completed successfully and have no running or pending Pods left.",
		},
		{
			ID: "completed-pods", Category: LibraryCategoryUnused, Group: LibraryGroupWorkloads,
			sourcePath:  "pods/completed_pods.yaml",
			Title:       "Completed Pods",
			Description: "Finds Pods with condition PodCompleted set to true.",
		},

		// --- unused-resources / Networking ---
		{
			ID: "stale-ingresses", Category: LibraryCategoryUnused, Group: LibraryGroupNetworking,
			sourcePath:  "ingresses/unused_ingresses.yaml",
			Title:       "Stale Ingresses",
			Description: "Finds Ingress instances whose default backend or referenced Services no longer exist.",
		},

		// --- unused-resources / Storage ---
		{
			ID: "stale-persistent-volume-claim", Category: LibraryCategoryUnused, Group: LibraryGroupStorage,
			sourcePath:  "persistent-volume-claims/unused_persistent-volume-claims.yaml",
			Title:       "Unused PersistentVolumeClaims",
			Description: "Finds PersistentVolumeClaims, across all namespaces, not used by any Pod.",
		},
		{
			ID: "unbound-persistent-volumes", Category: LibraryCategoryUnused, Group: LibraryGroupStorage,
			sourcePath:  "persistent-volumes/unbound_persistent-volumes.yaml",
			Title:       "Unbound PersistentVolumes",
			Description: "Finds PersistentVolumes whose phase is anything other than Bound.",
		},

		// --- unused-resources / Cluster ---
		{
			ID: "unused-nodes", Category: LibraryCategoryUnused, Group: LibraryGroupCluster,
			sourcePath:  "nodes/unused_nodes.yaml",
			Title:       "Nodes With No Scheduled Pods",
			Description: "Finds Nodes with no Pods scheduled on them (DaemonSet Pods count as scheduled).",
		},

		// --- unused-resources / Time-Based ---
		{
			ID: "expire-date-based-cleaner", Category: LibraryCategoryUnused, Group: LibraryGroupTimeBased,
			sourcePath: "time_based_delete/delete_resource_based_on_expire_date.yaml",
			Title:      "Resources Past Their Expiry Date",
			Description: "Finds Deployments, StatefulSets, and Services carrying a cleaner/expires annotation " +
				"whose expiry date has passed. Opt in per-resource by setting the annotation, no editing required.",
		},
		{
			ID: "ttl-based-cleaner", Category: LibraryCategoryUnused, Group: LibraryGroupTimeBased,
			sourcePath: "time_based_delete/delete_resource_based_on_ttl_annotation.yaml",
			Title:      "Resources Past Their TTL",
			Description: "Finds Deployments, StatefulSets, and Services carrying a cleaner/ttl annotation " +
				"whose time-to-live has elapsed. Opt in per-resource by setting the annotation, no editing required.",
		},

		// --- unhealthy-resources / Pods ---
		{
			ID: "list-pods-with-expired-certificates", Category: LibraryCategoryUnhealthy, Group: LibraryGroupPods,
			sourcePath: "pod-with-expired-certificates/pods-with-expired-certificates.yaml",
			Title:      "Pods With Expired Certificates",
			Description: "Finds Pods mounting a cert-manager Certificate Secret that was renewed after the " +
				"Pod started, meaning the Pod is running with a stale certificate.",
		},
		{
			ID: "list-pods-with-outdated-secret-data", Category: LibraryCategoryUnhealthy, Group: LibraryGroupPods,
			sourcePath:  "pod-with-outdated-secrets/pods-with-outdated-secret-data.yaml",
			Title:       "Pods With Outdated Secret Data",
			Description: "Finds Pods mounting Secrets that have been modified since the Pod was created.",
		},
		{
			ID: "terminating-pods", Category: LibraryCategoryUnhealthy, Group: LibraryGroupPods,
			sourcePath:  "pod-with-terminating-state/pod-with-terminating-state.yaml",
			Title:       "Pods Stuck Terminating",
			Description: "Finds Pods that have a deletionTimestamp set, i.e. stuck in a terminating state.",
		},
		{
			ID: "error-state-pods", Category: LibraryCategoryUnhealthy, Group: LibraryGroupPods,
			sourcePath: "pod-with-error-state/pod-with-error-state.yml",
			Title:      "Pods in an Error State",
			Description: "Finds Pods in Failed phase with a container terminated with reason Error and a " +
				"non-zero exit code.",
		},
		{
			ID: "evicted-pods", Category: LibraryCategoryUnhealthy, Group: LibraryGroupPods,
			sourcePath:  "pod-with-evicted-state/pod-with-evicted-state.yml",
			Title:       "Evicted Pods",
			Description: "Finds Pods that failed with reason Evicted.",
		},

		// --- unhealthy-resources / Broken References ---
		{
			ID: "deployment-referencing-non-existent-resources", Category: LibraryCategoryUnhealthy,
			Group:       LibraryGroupBrokenReference,
			sourcePath:  "object-references/deployment-referencing-non-existent-resources.yaml",
			Title:       "Deployments Referencing Missing ConfigMaps or Secrets",
			Description: "Finds Deployments that mount a Secret or ConfigMap which no longer exists.",
		},
		{
			ID: "unhealthy-ingresses", Category: LibraryCategoryUnhealthy, Group: LibraryGroupBrokenReference,
			sourcePath: "object-references/ingress-referencing-non-existent-service.yaml",
			Title:      "Ingresses Referencing Missing Services",
			Description: "Finds Ingresses whose default backend, or one of the Services referenced via " +
				"spec.rules, no longer exists.",
		},
	}
)

// librarySummaryResponse is the JSON representation of a library entry for the card
// list. Selectors are included here (not just in the detail response) so the
// dashboard can render kind chips and support kind-based filtering without a
// second fetch per card.
type librarySummaryResponse struct {
	ID          string         `json:"id"`
	Category    string         `json:"category"`
	Group       string         `json:"group"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Schedule    string         `json:"schedule"`
	Selectors   []selectorInfo `json:"selectors"`
}

// libraryDetailResponse is the JSON representation of a library entry's full recipe.
type libraryDetailResponse struct {
	librarySummaryResponse
	LuaScript string `json:"luaScript,omitempty"`
}

// ListLibraryHandler returns every curated library recipe.
func ListLibraryHandler(log logr.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result := make([]librarySummaryResponse, 0, len(libraryManifest))
		for i := range libraryManifest {
			entry := &libraryManifest[i]
			cleaner, err := loadLibraryCleaner(entry)
			if err != nil {
				log.Error(err, "failed to load library recipe", "id", entry.ID)
				respondError(w, http.StatusInternalServerError, "failed to load library")
				return
			}
			result = append(result, toLibrarySummary(entry, cleaner))
		}
		respondJSON(w, http.StatusOK, result)
	}
}

// GetLibraryEntryHandler returns a single library recipe with full details, including its Lua script.
func GetLibraryEntryHandler(log logr.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		entry := findLibraryEntry(id)
		if entry == nil {
			respondError(w, http.StatusNotFound, "library recipe not found")
			return
		}

		cleaner, err := loadLibraryCleaner(entry)
		if err != nil {
			log.Error(err, "failed to load library recipe", "id", id)
			respondError(w, http.StatusInternalServerError, "failed to load library")
			return
		}

		respondJSON(w, http.StatusOK, toLibraryDetail(entry, cleaner))
	}
}

func findLibraryEntry(id string) *libraryManifestEntry {
	for i := range libraryManifest {
		if libraryManifest[i].ID == id {
			return &libraryManifest[i]
		}
	}
	return nil
}

// loadLibraryCleaner reads and parses the embedded YAML for a manifest entry.
func loadLibraryCleaner(entry *libraryManifestEntry) (*appsv1alpha1.Cleaner, error) {
	fsys, err := libraryFS(entry.Category)
	if err != nil {
		return nil, err
	}

	data, err := fsys.ReadFile(entry.sourcePath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", entry.sourcePath, err)
	}

	var cleaner appsv1alpha1.Cleaner
	if err := yaml.Unmarshal(data, &cleaner); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", entry.sourcePath, err)
	}
	return &cleaner, nil
}

func libraryFS(category string) (embed.FS, error) {
	switch category {
	case LibraryCategoryUnused:
		return unusedresources.FS(), nil
	case LibraryCategoryUnhealthy:
		return unhealthyresources.FS(), nil
	default:
		return embed.FS{}, fmt.Errorf("unknown library category %q", category)
	}
}

func toLibrarySummary(entry *libraryManifestEntry, cleaner *appsv1alpha1.Cleaner) librarySummaryResponse {
	return librarySummaryResponse{
		ID:          entry.ID,
		Category:    entry.Category,
		Group:       entry.Group,
		Title:       entry.Title,
		Description: entry.Description,
		Schedule:    cleaner.Spec.Schedule,
		Selectors:   selectorInfosFromCleaner(cleaner),
	}
}

func toLibraryDetail(entry *libraryManifestEntry, cleaner *appsv1alpha1.Cleaner) libraryDetailResponse {
	return libraryDetailResponse{
		librarySummaryResponse: toLibrarySummary(entry, cleaner),
		LuaScript:              primaryLuaScript(cleaner),
	}
}
