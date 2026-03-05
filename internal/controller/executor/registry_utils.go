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
	"context"
	"fmt"
	"os"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
)

// 1. Key Generation Logic
// getResourceKey returns a unique identifier: <Kind>/<UID>
func getResourceKey(obj *unstructured.Unstructured) string {
	return fmt.Sprintf("%s__%s", obj.GetKind(), string(obj.GetUID()))
}

// getConfigMapInfo generates the unique name and namespace for the registry.
func getConfigMapInfo(cleaner *appsv1alpha1.Cleaner) types.NamespacedName {
	return types.NamespacedName{
		Namespace: os.Getenv(namespace),
		Name:      fmt.Sprintf("cleaner-%s", cleaner.Name),
	}
}

// 2. Fetching Logic
// getThrottledResources parses the ConfigMap Data into a map of [Key]Count.
func getThrottledResources(ctx context.Context, cleaner *appsv1alpha1.Cleaner,
) (map[string]int, error) {

	registry := make(map[string]int)
	if cleaner.Spec.OccurrenceThreshold <= 1 {
		return registry, nil
	}

	configMap, err := getRegistryConfigMap(ctx, cleaner)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return registry, nil
		}
		return nil, err
	}

	if configMap.Data == nil {
		return registry, nil
	}

	for key, val := range configMap.Data {
		count, err := strconv.Atoi(val)
		if err != nil {
			continue
		}
		registry[key] = count
	}
	return registry, nil
}

// getRegistryConfigMap fetches the ConfigMap from the cluster.
func getRegistryConfigMap(ctx context.Context, cleaner *appsv1alpha1.Cleaner,
) (*corev1.ConfigMap, error) {

	info := getConfigMapInfo(cleaner)
	configMap := &corev1.ConfigMap{}
	err := k8sClient.Get(ctx, info, configMap)
	return configMap, err
}

// 3. Filtering Logic
// filterResourcesByThreshold returns only the ResourceResults that have reached threshold.
func filterResourcesByThreshold(
	results []ResourceResult,
	throttledResources map[string]int,
	threshold int,
) []ResourceResult {

	if threshold <= 1 {
		return results
	}

	var toProcess []ResourceResult
	for i := range results {
		obj := results[i].Resource
		if obj == nil {
			continue
		}

		key := getResourceKey(obj)
		if count, ok := throttledResources[key]; ok {
			if count >= threshold {
				toProcess = append(toProcess, results[i])
			}
		}
	}
	return toProcess
}

// 4. Persistence & Sync Logic
// updateRegistry generates the NEW state and persists it to the ConfigMap.
func updateRegistry(ctx context.Context, cleaner *appsv1alpha1.Cleaner,
	results []ResourceResult, oldRegistry map[string]int) error {

	if cleaner.Spec.OccurrenceThreshold <= 1 {
		return nil
	}

	// Build the new data map from current matches
	newData := make(map[string]string)
	for _, res := range results {
		if res.Resource == nil {
			continue
		}
		key := getResourceKey(res.Resource)
		count := 1
		if val, ok := oldRegistry[key]; ok {
			count = val + 1
		}
		newData[key] = strconv.Itoa(count)
	}

	info := getConfigMapInfo(cleaner)
	configMap, err := getRegistryConfigMap(ctx, cleaner)

	isNew := false
	if err != nil {
		if apierrors.IsNotFound(err) {
			isNew = true
			configMap = &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      info.Name,
					Namespace: info.Namespace,
					OwnerReferences: []metav1.OwnerReference{
						*metav1.NewControllerRef(cleaner, appsv1alpha1.GroupVersion.WithKind("Cleaner")),
					},
				},
			}
		} else {
			return err
		}
	}

	configMap.Data = newData

	if isNew {
		return k8sClient.Create(ctx, configMap)
	}
	return k8sClient.Update(ctx, configMap)
}

// 5. Cleanup Logic
// DeleteConfigMap removes the registry ConfigMap from the cluster.
func DeleteConfigMap(ctx context.Context, cleaner *appsv1alpha1.Cleaner) error {
	configMap := &corev1.ConfigMap{}
	err := k8sClient.Get(ctx, getConfigMapInfo(cleaner), configMap)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	return k8sClient.Delete(ctx, configMap)
}
