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

package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	lua "github.com/yuin/gopher-lua"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"

	libsveltosv1beta1 "github.com/projectsveltos/libsveltos/api/v1beta1"
	logs "github.com/projectsveltos/libsveltos/lib/logsettings"
)

// A "request" represents a Cleaner instance that needs to be processed.
//
// The flow is following:
// - when a request arrives, it is first added to the dirty set or dropped if it already
// present in the dirty set;
// - pushed to the jobQueue only if it is not presented in inProgress (we don't want
// to process same request in parallel)
//
// When a worker is ready to serve a request, it gets the request from the
// front of the jobQueue.
// The request is also added to the inProgress set and removed from the dirty set.
//
// If a request, currently in the inProgress arrives again, such request is only added
// to the dirty set, not to the queue. This guarantees that same request to collect resources
// is never process more than once in parallel.
//
// When worker is done, the request is removed from the inProgress set.
// If the same request is also present in the dirty set, it is added back to the back of the jobQueue.

type ResourceResult struct {
	// Resource identify a Kubernetes resource
	Resource *unstructured.Unstructured `json:"resource,omitempty"`

	// Message is an optional field.
	// +optional
	Message string `json:"message,omitempty"`
}

type responseParams struct {
	cleanerName string
	err         error
}

var (
	k8sClient client.Client
	config    *rest.Config
	scheme    *runtime.Scheme
)

const (
	luaTableError = "lua script output is not a lua table"
	luaBoolError  = "lua script output is not a lua bool"
)

type evaluateStatus struct {
	Matching bool   `json:"matching"`
	Message  string `json:"message"`
}

type transformStatus struct {
	Resource *unstructured.Unstructured `json:"resource"`
	Message  string                     `json:"message"`
}

type aggregatedStatus struct {
	Resources []ResourceResult `json:"resources,omitempty"`
}

func processRequests(ctx context.Context, i int, logger logr.Logger) {
	id := i
	var cleanerName *string

	logger.V(logs.LogDebug).Info(fmt.Sprintf("started worker %d", id))

	for {
		if cleanerName != nil {
			l := logger.WithValues("cleaner", cleanerName)
			// Get error only from getIsCleanupFromKey as same key is always used
			l.Info(fmt.Sprintf("worker: %d processing request", id))
			err := processCleanerInstance(ctx, *cleanerName, l)
			storeResult(*cleanerName, err, l)
			l.Info(fmt.Sprintf("worker: %d request processed", id))
		}
		cleanerName = nil
		select {
		case <-time.After(1 * time.Second):
			managerInstance.mu.Lock()
			if len(managerInstance.jobQueue) > 0 {
				// take a request from queue and remove it from queue
				cleanerName = &managerInstance.jobQueue[0]
				managerInstance.jobQueue = managerInstance.jobQueue[1:]
				l := logger.WithValues("cleaner", cleanerName)
				l.V(logs.LogDebug).Info("take from jobQueue")
				// Add to inProgress
				l.V(logs.LogDebug).Info("add to inProgress")
				key := *cleanerName
				managerInstance.inProgress = append(managerInstance.inProgress, key)
				// If present remove from dirty
				for i := range managerInstance.dirty {
					if managerInstance.dirty[i] == key {
						l.V(logs.LogDebug).Info("remove from dirty")
						managerInstance.dirty = removeFromSlice(managerInstance.dirty, i)
						break
					}
				}
			}
			managerInstance.mu.Unlock()
		case <-ctx.Done():
			logger.V(logs.LogDebug).Info("context canceled")
			return
		}
	}
}

func processCleanerInstance(ctx context.Context, cleanerName string, logger logr.Logger) error {
	cleaner, err := getCleanerInstance(ctx, cleanerName)
	if err != nil {
		logger.Info(fmt.Sprintf("failed to get cleaner instance: %v", err))
		return err
	}
	if cleaner == nil {
		logger.V(logs.LogDebug).Info("cleaner instance not found")
		return nil
	}

	resources := make([]ResourceResult, 0)
	for i := range cleaner.Spec.ResourcePolicySet.ResourceSelectors {
		selector := &cleaner.Spec.ResourcePolicySet.ResourceSelectors[i]
		var tmpResources []ResourceResult
		tmpResources, err = getMatchingResources(ctx, selector, logger)
		if err != nil {
			logger.Info(fmt.Sprintf("failed to fetch resource (gvk: %s): %v",
				fmt.Sprintf("%s:%s:%s", selector.Group, selector.Version, selector.Kind), err))
			return err
		}
		resources = append(resources, tmpResources...)
	}

	if cleaner.Spec.ResourcePolicySet.AggregatedSelection != "" {
		resources, err = aggregatedSelection(cleaner.Spec.ResourcePolicySet.AggregatedSelection, resources,
			logger)
		if err != nil {
			logger.Info(fmt.Sprintf("failed to filter aggregated resources: %v", err))
			return err
		}
	}

	var processedResources []ResourceResult
	switch cleaner.Spec.Action {
	case appsv1alpha1.ActionDelete:
		processedResources, err = deleteMatchingResources(ctx, resources, cleaner.Spec.DeleteOptions, logger)
	case appsv1alpha1.ActionTransform:
		processedResources, err = updateMatchingResources(ctx, resources, cleaner.Spec.Transform, logger)
	case appsv1alpha1.ActionScan:
		printMatchingResources(resources, logger)
		processedResources = resources
	}

	// Send notification irrespective of err
	sendErr := sendNotifications(ctx, processedResources, cleaner, logger)
	if sendErr != nil {
		return sendErr
	}

	// Store resources before any action was taken irrespective of err
	storeErr := storeResources(processedResources, scheme, cleaner, logger)
	if storeErr != nil {
		return storeErr
	}

	return err
}

func getMatchingResources(ctx context.Context, sr *appsv1alpha1.ResourceSelector, logger logr.Logger,
) ([]ResourceResult, error) {

	resources, err := fetchResources(ctx, sr, logger)
	if err != nil {
		logger.Info(fmt.Sprintf("failed to fetch resources: %v", err))
		return nil, err
	}

	if resources == nil {
		return nil, nil
	}

	results := make([]ResourceResult, 0)
	for i := range resources {
		resource := &resources[i]
		if sr.ExcludeDeleted && !resource.GetDeletionTimestamp().IsZero() {
			continue
		}
		l := logger.WithValues("resource", fmt.Sprintf("%s:%s/%s",
			resource.GetKind(), resource.GetNamespace(), resource.GetName()))
		l.V(logs.LogDebug).Info("considering resource for deletion")

		isMatch, message, err := isMatch(resource, sr.Evaluate, l)
		if err != nil {
			return nil, err
		}
		if isMatch {
			l.Info(fmt.Sprintf("getMatchingResources: found a match %q", message))
			resourceInfo := ResourceResult{
				Resource: resource,
				Message:  message,
			}
			results = append(results, resourceInfo)
		}
	}

	return results, nil
}

func deleteMatchingResources(ctx context.Context, resources []ResourceResult,
	deleteOptions *appsv1alpha1.DeleteOptions, logger logr.Logger) ([]ResourceResult, error) {

	processedResources := make([]ResourceResult, 0)
	var failedActions []error // Slice to store all errors

	for i := range resources {
		resource := resources[i]
		l := logger.WithValues("resource", fmt.Sprintf("%s:%s/%s",
			resource.Resource.GetKind(),
			resource.Resource.GetNamespace(),
			resource.Resource.GetName()))
		l.Info("deleting resource")

		options := &client.DeleteOptions{}
		if deleteOptions != nil {
			options.GracePeriodSeconds = deleteOptions.GracePeriodSeconds
			options.PropagationPolicy = deleteOptions.PropagationPolicy
		}

		if err := k8sClient.Delete(ctx, resource.Resource, options); err != nil {
			if apierrors.IsNotFound(err) {
				// Cleaner was about to delete a resource, but the resource is gone
				// Ignore this error as outcome is that resource is gone either way.
				continue
			}
			l.Info(fmt.Sprintf("failed to delete resource: %v", err))
			failedActions = append(failedActions, err)
		} else {
			processedResources = append(processedResources, resource)
		}
	}

	if len(failedActions) > 0 {
		// Use errors.Join to combine all collected errors into a single error
		return processedResources, errors.Join(failedActions...)
	}

	return processedResources, nil
}

func updateMatchingResources(ctx context.Context, resources []ResourceResult,
	transformFunction string, logger logr.Logger) ([]ResourceResult, error) {

	processedResources := make([]ResourceResult, 0)
	var failedActions []error // Slice to store all errors

	for i := range resources {
		resource := resources[i]
		l := logger.WithValues("resource", fmt.Sprintf("%s:%s/%s",
			resource.Resource.GetKind(),
			resource.Resource.GetNamespace(),
			resource.Resource.GetName()))
		l.Info("updating resource")
		newResource, err := transform(resource.Resource, transformFunction, l)
		if err != nil {
			l.Info(fmt.Sprintf("failed to transform resource: %v", err))
			failedActions = append(failedActions, err)
			continue
		}
		if err := k8sClient.Update(ctx, newResource); err != nil {
			l.Info(fmt.Sprintf("failed to update resource: %v", err))
			failedActions = append(failedActions, err)
			continue
		}
		processedResources = append(processedResources, resource)
	}

	if len(failedActions) > 0 {
		// Use errors.Join to combine all collected errors into a single error
		return processedResources, errors.Join(failedActions...)
	}

	return processedResources, nil
}

func fetchResources(ctx context.Context, resourceSelector *appsv1alpha1.ResourceSelector,
	logger logr.Logger) ([]unstructured.Unstructured, error) {

	gvk := schema.GroupVersionKind{
		Group:   resourceSelector.Group,
		Version: resourceSelector.Version,
		Kind:    resourceSelector.Kind,
	}

	dc := discovery.NewDiscoveryClientForConfigOrDie(config)
	groupResources, err := restmapper.GetAPIGroupResources(dc)
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDiscoveryRESTMapper(groupResources)

	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		if meta.IsNoMatchError(err) {
			return nil, nil
		}
		return nil, err
	}

	resourceId := schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: mapping.Resource.Resource,
	}

	options := metav1.ListOptions{}

	if len(resourceSelector.LabelFilters) > 0 {
		options.LabelSelector = addLabelFilters(resourceSelector.LabelFilters)
	}

	var namespaces []string
	namespaces, err = getNamespaces(ctx, resourceSelector, logger)
	if err != nil {
		return nil, err
	}

	var result []unstructured.Unstructured
	if len(namespaces) > 0 {
		result, err = collectFromNamespaces(ctx, config, namespaces, &resourceId, &options)
	} else {
		result, err = collectWithOptions(ctx, config, &resourceId, &options)
	}

	return result, err
}

func addLabelFilters(labelFilters []libsveltosv1beta1.LabelFilter) string {
	labelFilter := ""
	if len(labelFilters) > 0 {
		for i := range labelFilters {
			if labelFilter != "" {
				labelFilter += ","
			}
			f := labelFilters[i]
			switch f.Operation {
			case libsveltosv1beta1.OperationEqual:
				labelFilter += fmt.Sprintf("%s=%s", f.Key, f.Value)
			case libsveltosv1beta1.OperationDifferent:
				labelFilter += fmt.Sprintf("%s!=%s", f.Key, f.Value)
			case libsveltosv1beta1.OperationHas:
				// Key exists, value is not checked
				labelFilter += f.Key
			case libsveltosv1beta1.OperationDoesNotHave:
				// Key does not exist
				labelFilter += fmt.Sprintf("!%s", f.Key)
			}
		}
	}

	return labelFilter
}

func collectFromNamespaces(ctx context.Context, config *rest.Config, namespaces []string,
	resourceId *schema.GroupVersionResource, options *metav1.ListOptions) ([]unstructured.Unstructured, error) {

	result := make([]unstructured.Unstructured, 0)
	for i := range namespaces {
		tmpOptions := *options
		tmpOptions.FieldSelector += fmt.Sprintf("metadata.namespace=%s", namespaces[i])
		tmpResult, err := collectWithOptions(ctx, config, resourceId, &tmpOptions)
		if err != nil {
			return nil, err
		}

		result = append(result, tmpResult...)
	}

	return result, nil
}

func collectWithOptions(ctx context.Context, config *rest.Config,
	resourceId *schema.GroupVersionResource, options *metav1.ListOptions) ([]unstructured.Unstructured, error) {

	d := dynamic.NewForConfigOrDie(config)
	list, err := d.Resource(*resourceId).List(ctx, *options)
	if err != nil {
		return nil, err
	}

	result := make([]unstructured.Unstructured, len(list.Items))

	copy(result, list.Items)

	return result, nil
}

// getNamespaces returns all namespaces to consider:
// - if resourceSelector.Namespace is defined, such namespace is considered
// - if resourceSelector.NamespaceSelector is defined, all matching namespaces are also considered
func getNamespaces(ctx context.Context, resourceSelector *appsv1alpha1.ResourceSelector,
	logger logr.Logger) ([]string, error) {

	matchingNamespaces := make([]string, 0)

	addedNamespace := make(map[string]bool)
	if resourceSelector.NamespaceSelector != "" {
		parsedSelector, err := labels.Parse(resourceSelector.NamespaceSelector)
		if err != nil {
			logger.V(logs.LogInfo).Info(fmt.Sprintf("failed to parse NamespaceSelector: %v", err))
			return nil, err
		}

		namespaces := &corev1.NamespaceList{}
		err = k8sClient.List(ctx, namespaces)
		if err != nil {
			logger.Error(err, "failed to list all namespaces")
			return nil, err
		}

		for i := range namespaces.Items {
			ns := &namespaces.Items[i]

			if !ns.DeletionTimestamp.IsZero() {
				// Only existing namespaces can match
				continue
			}

			err = addTypeInformationToObject(scheme, ns)
			if err != nil {
				return nil, err
			}
			if parsedSelector.Matches(labels.Set(ns.Labels)) {
				matchingNamespaces = append(matchingNamespaces, ns.Name)
				addedNamespace[ns.Name] = true
			}
		}
	}

	if resourceSelector.Namespace != "" {
		// if resourceSelector.Namespace was already a match for resourceSelector.NamespaceSelector
		// do not add it again
		if isPresent := addedNamespace[resourceSelector.Namespace]; !isPresent {
			matchingNamespaces = append(matchingNamespaces, resourceSelector.Namespace)
		}
	}

	return matchingNamespaces, nil
}

func isMatch(resource *unstructured.Unstructured, script string, logger logr.Logger,
) (matching bool, message string, err error) {

	if script == "" {
		return true, "", nil
	}

	l := lua.NewState()
	defer l.Close()

	obj := mapToTable(resource.UnstructuredContent())

	if err = l.DoString(script); err != nil {
		logger.Info(fmt.Sprintf("doString failed: %v", err))
		return false, "", err
	}

	l.SetGlobal("obj", obj)

	if err = l.CallByParam(lua.P{
		Fn:      l.GetGlobal("evaluate"), // name of Lua function
		NRet:    1,                       // number of returned values
		Protect: true,                    // return err or panic
	}, obj); err != nil {
		logger.Info(fmt.Sprintf("failed to evaluate health for resource: %v", err))
		return false, "", err
	}

	lv := l.Get(-1)
	tbl, ok := lv.(*lua.LTable)
	if !ok {
		logger.Info(luaTableError)
		return false, "", fmt.Errorf("%s", luaTableError)
	}

	goResult := toGoValue(tbl)
	var resultJson []byte
	resultJson, err = json.Marshal(goResult)
	if err != nil {
		logger.Info(fmt.Sprintf("failed to marshal result: %v", err))
		return false, "", err
	}

	var result evaluateStatus
	err = json.Unmarshal(resultJson, &result)
	if err != nil {
		logger.Info(fmt.Sprintf("failed to marshal result: %v", err))
		return false, "", err
	}

	if result.Message != "" {
		logger.Info(fmt.Sprintf("message: %s", result.Message))
	}

	logger.V(logs.LogDebug).Info(fmt.Sprintf("is a match: %t", result.Matching))

	return result.Matching, result.Message, nil
}

func transform(resource *unstructured.Unstructured, script string, logger logr.Logger,
) (*unstructured.Unstructured, error) {

	if script == "" {
		return resource, nil
	}

	l := lua.NewState()
	defer l.Close()

	obj := mapToTable(resource.UnstructuredContent())

	if err := l.DoString(script); err != nil {
		logger.Info(fmt.Sprintf("doString failed: %v", err))
		return nil, err
	}

	l.SetGlobal("obj", obj)

	if err := l.CallByParam(lua.P{
		Fn:      l.GetGlobal("transform"), // name of Lua function
		NRet:    1,                        // number of returned values
		Protect: true,                     // return err or panic
	}, obj); err != nil {
		logger.Info(fmt.Sprintf("failed to evaluate health for resource: %v", err))
		return nil, err
	}

	lv := l.Get(-1)
	tbl, ok := lv.(*lua.LTable)
	if !ok {
		logger.Info(luaTableError)
		return nil, fmt.Errorf("%s", luaTableError)
	}

	goResult := toGoValue(tbl)
	resultJson, err := json.Marshal(goResult)
	if err != nil {
		logger.Info(fmt.Sprintf("failed to marshal result: %v", err))
		return nil, err
	}

	var result transformStatus
	err = json.Unmarshal(resultJson, &result)
	if err != nil {
		logger.Info(fmt.Sprintf("failed to marshal result: %v", err))
		return nil, err
	}

	if result.Message != "" {
		logger.Info(fmt.Sprintf("message: %s", result.Message))
	}

	return result.Resource, nil
}

func aggregatedSelection(luaScript string, resources []ResourceResult, logger logr.Logger) ([]ResourceResult, error) {
	if luaScript == "" {
		return resources, nil
	}

	// Create a new Lua state
	l := lua.NewState()
	defer l.Close()

	// Load the Lua script
	if err := l.DoString(luaScript); err != nil {
		logger.V(logs.LogInfo).Info(fmt.Sprintf("doString failed: %v", err))
		return nil, err
	}

	// Create an argument table
	argTable := l.NewTable()
	for _, resource := range resources {
		obj := mapToTable(resource.Resource.UnstructuredContent())
		argTable.Append(obj)
	}

	l.SetGlobal("resources", argTable)

	if err := l.CallByParam(lua.P{
		Fn:      l.GetGlobal("evaluate"), // name of Lua function
		NRet:    1,                       // number of returned values
		Protect: true,                    // return err or panic
	}, argTable); err != nil {
		logger.V(logs.LogInfo).Info(fmt.Sprintf("failed to call evaluate function: %s", err.Error()))
		return nil, err
	}

	lv := l.Get(-1)
	tbl, ok := lv.(*lua.LTable)
	if !ok {
		logger.V(logs.LogInfo).Info(luaTableError)
		return nil, fmt.Errorf("%s", luaTableError)
	}

	goResult := toGoValue(tbl)
	resultJson, err := json.Marshal(goResult)
	if err != nil {
		logger.V(logs.LogInfo).Info(fmt.Sprintf("failed to marshal result: %v", err))
		return nil, err
	}

	var result aggregatedStatus
	err = json.Unmarshal(resultJson, &result)
	if err != nil {
		logger.V(logs.LogInfo).Info(fmt.Sprintf("failed to marshal result: %v", err))
		return nil, err
	}

	for i := range result.Resources {
		l := logger.WithValues("resource", fmt.Sprintf("%s:%s/%s",
			result.Resources[i].Resource.GetKind(),
			result.Resources[i].Resource.GetNamespace(),
			result.Resources[i].Resource.GetName()))
		if result.Resources[i].Message != "" {
			l.V(logs.LogInfo).Info(fmt.Sprintf("message: %s", result.Resources[i].Message))
		}

		l.Info("aggregatedSelection: found a match")
	}

	return result.Resources, nil
}

func getCleanerInstance(ctx context.Context, cleanerName string) (*appsv1alpha1.Cleaner, error) {
	cleaner := &appsv1alpha1.Cleaner{}
	err := k8sClient.Get(ctx, types.NamespacedName{Name: cleanerName}, cleaner)

	if apierrors.IsNotFound(err) {
		err = nil
	}

	return cleaner, err
}

// storeResult does following:
// - set results for further in time lookup
// - remove request from inProgress
// - if request is in dirty, remove it from there and add it to the back of the jobQueue
func storeResult(cleanerName string, err error, logger logr.Logger) {
	managerInstance.mu.Lock()
	defer managerInstance.mu.Unlock()

	key := cleanerName

	// Remove from inProgress
	for i := range managerInstance.inProgress {
		if managerInstance.inProgress[i] != key {
			continue
		}
		logger.V(logs.LogDebug).Info("remove from inProgress")
		managerInstance.inProgress = removeFromSlice(managerInstance.inProgress, i)
		break
	}

	if err != nil {
		logger.V(logs.LogDebug).Info(fmt.Sprintf("added to result with err %s", err.Error()))
	} else {
		logger.V(logs.LogDebug).Info("added to result")
	}
	managerInstance.results[key] = err

	// if key is in dirty, remove from there and push to jobQueue
	for i := range managerInstance.dirty {
		if managerInstance.dirty[i] != key {
			continue
		}
		logger.V(logs.LogDebug).Info("add to jobQueue")
		managerInstance.jobQueue = append(managerInstance.jobQueue, key)
		logger.V(logs.LogDebug).Info("remove from dirty")
		managerInstance.dirty = removeFromSlice(managerInstance.dirty, i)
		logger.V(logs.LogDebug).Info("remove result")
		delete(managerInstance.results, key)
		break
	}
}

// getRequestStatus gets requests status.
// If result is available it returns the result.
// If request is still queued, responseParams is nil and an error is nil.
// If result is not available and request is neither queued nor already processed, it returns an error to indicate that.
func getRequestStatus(cleanerName string) (*responseParams, error) {
	logger := managerInstance.log.WithValues("cleaner", cleanerName)
	managerInstance.mu.Lock()
	defer managerInstance.mu.Unlock()

	key := cleanerName

	logger.V(logs.LogDebug).Info("searching result")
	if _, ok := managerInstance.results[key]; ok {
		logger.V(logs.LogDebug).Info("request already processed, result present. returning result.")
		if managerInstance.results[key] != nil {
			logger.V(logs.LogDebug).Info("returning a response with an error")
		}
		resp := responseParams{
			cleanerName: key,
			err:         managerInstance.results[key],
		}
		logger.V(logs.LogDebug).Info("removing result")
		delete(managerInstance.results, key)
		return &resp, nil
	}

	for i := range managerInstance.inProgress {
		if managerInstance.inProgress[i] == key {
			logger.V(logs.LogDebug).Info("request is still in inProgress, so being processed")
			return nil, nil
		}
	}

	for i := range managerInstance.jobQueue {
		if managerInstance.jobQueue[i] == key {
			logger.V(logs.LogDebug).Info("request is still in jobQueue, so waiting to be processed.")
			return nil, nil
		}
	}

	// if we get here it means, we have no response for this request, nor the
	// request is queued or being processed
	logger.V(logs.LogDebug).Info("request has not been processed nor is currently queued.")
	return nil, fmt.Errorf("request has not been processed nor is currently queued")
}

func removeFromSlice(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

// mapToTable converts a Go map to a lua table
// credit to: https://github.com/yuin/gopher-lua/issues/160#issuecomment-447608033
func mapToTable(m map[string]interface{}) *lua.LTable {
	// Main table pointer
	resultTable := &lua.LTable{}

	// Loop map
	for key, element := range m {
		switch element := element.(type) {
		case float64:
			resultTable.RawSetString(key, lua.LNumber(element))
		case int64:
			resultTable.RawSetString(key, lua.LNumber(element))
		case string:
			resultTable.RawSetString(key, lua.LString(element))
		case bool:
			resultTable.RawSetString(key, lua.LBool(element))
		case []byte:
			resultTable.RawSetString(key, lua.LString(string(element)))
		case map[string]interface{}:

			// Get table from map
			tble := mapToTable(element)

			resultTable.RawSetString(key, tble)

		case time.Time:
			resultTable.RawSetString(key, lua.LNumber(element.Unix()))

		case []map[string]interface{}:

			// Create slice table
			sliceTable := &lua.LTable{}

			// Loop element
			for _, s := range element {
				// Get table from map
				tble := mapToTable(s)

				sliceTable.Append(tble)
			}

			// Set slice table
			resultTable.RawSetString(key, sliceTable)

		case []interface{}:

			// Create slice table
			sliceTable := &lua.LTable{}

			// Loop interface slice
			for _, s := range element {
				// Switch interface type
				switch s := s.(type) {
				case map[string]interface{}:

					// Convert map to table
					t := mapToTable(s)

					// Append result
					sliceTable.Append(t)

				case float64:

					// Append result as number
					sliceTable.Append(lua.LNumber(s))

				case string:

					// Append result as string
					sliceTable.Append(lua.LString(s))

				case bool:

					// Append result as bool
					sliceTable.Append(lua.LBool(s))
				}
			}

			// Append to main table
			resultTable.RawSetString(key, sliceTable)
		}
	}

	return resultTable
}

// toGoValue converts the given LValue to a Go object.
// Credit to: https://github.com/yuin/gluamapper/blob/master/gluamapper.go
func toGoValue(lv lua.LValue) interface{} {
	switch v := lv.(type) {
	case *lua.LNilType:
		return nil
	case lua.LBool:
		return bool(v)
	case lua.LString:
		return string(v)
	case lua.LNumber:
		return float64(v)
	case *lua.LTable:
		maxn := v.MaxN()
		if maxn == 0 { // table
			ret := make(map[string]interface{})
			v.ForEach(func(key, value lua.LValue) {
				keystr := fmt.Sprint(toGoValue(key))
				ret[keystr] = toGoValue(value)
			})
			return ret
		} else { // array
			ret := make([]interface{}, 0, maxn)
			for i := 1; i <= maxn; i++ {
				ret = append(ret, toGoValue(v.RawGetInt(i)))
			}
			return ret
		}
	default:
		return v
	}
}

func printMatchingResources(resources []ResourceResult, logger logr.Logger) {
	for i := range resources {
		resource := resources[i]
		l := logger.WithValues("resource", fmt.Sprintf("%s:%s/%s %q",
			resource.Resource.GetKind(),
			resource.Resource.GetNamespace(),
			resource.Resource.GetName(),
			resource.Message))
		l = l.WithValues("message", resource.Message)
		l.Info("resource is a match for cleaner")
	}
}
