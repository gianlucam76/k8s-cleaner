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

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"

	logs "github.com/projectsveltos/libsveltos/lib/logsettings"
)

// RollbackResourceResult is the outcome of rolling back a single resource.
type RollbackResourceResult struct {
	Kind      string `json:"kind"`
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name"`
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
}

// Rollback reverts the most recent Delete or Transform execution recorded for
// cleanerName, using the pre-action resource state captured in the Report
// instance's ResourceInfo.FullResource. It is a best-effort, per-resource
// operation: a failure on one resource does not stop the others from being
// attempted.
func Rollback(ctx context.Context, c client.Client, cleanerName string, logger logr.Logger,
) ([]RollbackResourceResult, error) {

	report := &appsv1alpha1.Report{}
	if err := c.Get(ctx, types.NamespacedName{Name: cleanerName}, report); err != nil {
		return nil, err
	}

	if report.Spec.Action == appsv1alpha1.ActionScan {
		return nil, fmt.Errorf("nothing to roll back: last execution's action was Scan")
	}

	results := make([]RollbackResourceResult, 0, len(report.Spec.ResourceInfo))
	for i := range report.Spec.ResourceInfo {
		resourceInfo := &report.Spec.ResourceInfo[i]
		results = append(results, rollbackResource(ctx, c, report.Spec.Action, resourceInfo, logger))
	}

	return results, nil
}

func rollbackResource(ctx context.Context, c client.Client, action appsv1alpha1.Action,
	resourceInfo *appsv1alpha1.ResourceInfo, logger logr.Logger) RollbackResourceResult {

	ref := resourceInfo.Resource
	result := RollbackResourceResult{
		Kind:      ref.Kind,
		Namespace: ref.Namespace,
		Name:      ref.Name,
	}

	if len(resourceInfo.FullResource) == 0 {
		result.Message = "no rollback data captured for this resource"
		return result
	}

	obj := &unstructured.Unstructured{}
	if err := obj.UnmarshalJSON(resourceInfo.FullResource); err != nil {
		result.Message = fmt.Sprintf("failed to parse captured resource: %v", err)
		return result
	}
	if obj.GroupVersionKind().Empty() {
		obj.SetAPIVersion(ref.APIVersion)
		obj.SetKind(ref.Kind)
	}

	l := logger.WithValues("resource", fmt.Sprintf("%s:%s/%s", ref.Kind, ref.Namespace, ref.Name))

	var err error
	switch action {
	case appsv1alpha1.ActionDelete:
		err = recreateResource(ctx, c, obj)
	case appsv1alpha1.ActionTransform:
		err = restoreResource(ctx, c, obj)
	case appsv1alpha1.ActionScan:
		// Nothing is ever captured for a Scan action.
	}

	if err != nil {
		l.V(logs.LogInfo).Info(fmt.Sprintf("failed to roll back resource: %v", err))
		result.Message = err.Error()
		return result
	}

	result.Success = true
	return result
}

// recreateResource re-creates a resource that was deleted, using its captured
// pre-deletion state.
func recreateResource(ctx context.Context, c client.Client, obj *unstructured.Unstructured) error {
	obj.SetResourceVersion("")
	obj.SetUID("")
	obj.SetCreationTimestamp(metav1.Time{})
	obj.SetManagedFields(nil)

	if err := c.Create(ctx, obj); err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil
		}
		return err
	}
	return nil
}

// restoreResource updates a live resource back to its captured pre-transform
// state. It fetches the live object first to pick up the current
// resourceVersion required for the update.
func restoreResource(ctx context.Context, c client.Client, obj *unstructured.Unstructured) error {
	live := &unstructured.Unstructured{}
	live.SetGroupVersionKind(obj.GroupVersionKind())
	if err := c.Get(ctx, types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}, live); err != nil {
		return err
	}

	obj.SetResourceVersion(live.GetResourceVersion())
	return c.Update(ctx, obj)
}
