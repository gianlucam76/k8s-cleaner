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
	"fmt"
	"time"

	"github.com/go-logr/logr"
	lua "github.com/yuin/gopher-lua"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "gianlucam76/k8s-pruner/api/v1alpha1"

	libsveltosv1alpha1 "github.com/projectsveltos/libsveltos/api/v1alpha1"
	logs "github.com/projectsveltos/libsveltos/lib/logsettings"
)

// A "request" represents a Pruner instance that needs to be processed.
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

type responseParams struct {
	prunerName string
	err        error
}

var (
	k8sClient client.Client
	config    *rest.Config
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

func processRequests(ctx context.Context, i int, logger logr.Logger) {
	id := i
	var prunerName *string

	logger.V(logs.LogDebug).Info(fmt.Sprintf("started worker %d", id))

	for {
		if prunerName != nil {
			l := logger.WithValues("pruner", prunerName)
			// Get error only from getIsCleanupFromKey as same key is always used
			l.Info(fmt.Sprintf("worker: %d processing request", id))
			err := processPrunerInstance(ctx, *prunerName, l)
			storeResult(*prunerName, err, l)
			l.Info(fmt.Sprintf("worker: %d request processed", id))
		}
		prunerName = nil
		select {
		case <-time.After(1 * time.Second):
			managerInstance.mu.Lock()
			if len(managerInstance.jobQueue) > 0 {
				// take a request from queue and remove it from queue
				prunerName = &managerInstance.jobQueue[0]
				managerInstance.jobQueue = managerInstance.jobQueue[1:]
				l := logger.WithValues("pruner", prunerName)
				l.V(logs.LogDebug).Info("take from jobQueue")
				// Add to inProgress
				l.V(logs.LogDebug).Info("add to inProgress")
				key := *prunerName
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

func processPrunerInstance(ctx context.Context, prunerName string, logger logr.Logger) error {
	pruner, err := getPrunerInstance(ctx, prunerName)
	if err != nil {
		logger.Info(fmt.Sprintf("failed to get pruner instance: %v", err))
		return err
	}
	if pruner == nil {
		logger.V(logs.LogDebug).Info("pruner instance not found")
		return nil
	}

	for i := range pruner.Spec.StaleResources {
		sr := &pruner.Spec.StaleResources[i]
		var resources []*unstructured.Unstructured
		resources, err = getStaleResources(ctx, sr, pruner.Spec.DryRun, logger)
		if err != nil {
			logger.Info(fmt.Sprintf("failed to fetch resource (gvk: %s): %v",
				fmt.Sprintf("%s:%s:%s", sr.Group, sr.Version, sr.Kind), err))
			return err
		}
		if sr.Action == appsv1alpha1.ActionDelete {
			return deleteStaleResources(ctx, resources, logger)
		} else {
			return updateStaleResources(ctx, resources, sr.Transform, logger)
		}
	}

	return nil
}

func getStaleResources(ctx context.Context, sr *appsv1alpha1.Resources, dryRun bool, logger logr.Logger,
) ([]*unstructured.Unstructured, error) {

	resources, err := fetchResources(ctx, sr)
	if err != nil {
		logger.Info(fmt.Sprintf("failed to fetch resources: %v", err))
		return nil, err
	}

	if resources == nil {
		return nil, nil
	}

	results := make([]*unstructured.Unstructured, 0)
	for i := range resources.Items {
		resource := &resources.Items[i]
		if !resource.GetDeletionTimestamp().IsZero() {
			continue
		}
		l := logger.WithValues("resource", fmt.Sprintf("%s:%s/%s",
			resource.GetKind(), resource.GetNamespace(), resource.GetName()))
		l.V(logs.LogDebug).Info("considering resource for deletion")

		isMatch, err := isMatch(resource, sr.Evaluate, l)
		if err != nil {
			return nil, err
		}
		if isMatch {
			l.Info("resource is a match for pruner")
			results = append(results, resource)
		}
	}

	return results, nil
}

func deleteStaleResources(ctx context.Context, resources []*unstructured.Unstructured,
	logger logr.Logger) error {

	for i := range resources {
		resource := resources[i]
		l := logger.WithValues("resource", fmt.Sprintf("%s:%s/%s",
			resource.GetKind(), resource.GetNamespace(), resource.GetName()))
		l.Info("deleting resource")
		if err := k8sClient.Delete(ctx, resource); err != nil {
			l.Info(fmt.Sprintf("failed to delete resource: %v", err))
			return err
		}
	}

	return nil
}

func updateStaleResources(ctx context.Context, resources []*unstructured.Unstructured,
	transform string, logger logr.Logger) error {

	for i := range resources {
		resource := resources[i]
		l := logger.WithValues("resource", fmt.Sprintf("%s:%s/%s",
			resource.GetKind(), resource.GetNamespace(), resource.GetName()))
		l.Info("updating resource")
		newResource, err := Transform(resource, transform, l)
		if err != nil {
			l.Info(fmt.Sprintf("failed to transform resource: %v", err))
			return err
		}
		if err := k8sClient.Update(ctx, newResource); err != nil {
			l.Info(fmt.Sprintf("failed to update resource: %v", err))
			return err
		}
	}

	return nil
}

func fetchResources(ctx context.Context, sr *appsv1alpha1.Resources) (*unstructured.UnstructuredList, error) {
	gvk := schema.GroupVersionKind{
		Group:   sr.Group,
		Version: sr.Version,
		Kind:    sr.Kind,
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

	if len(sr.LabelFilters) > 0 {
		labelFilter := ""
		for i := range sr.LabelFilters {
			if labelFilter != "" {
				labelFilter += ","
			}
			f := sr.LabelFilters[i]
			if f.Operation == libsveltosv1alpha1.OperationEqual {
				labelFilter += fmt.Sprintf("%s=%s", f.Key, f.Value)
			} else {
				labelFilter += fmt.Sprintf("%s!=%s", f.Key, f.Value)
			}
		}

		options.LabelSelector = labelFilter
	}

	if sr.Namespace != "" {
		options.FieldSelector += fmt.Sprintf("metadata.namespace=%s", sr.Namespace)
	}

	d := dynamic.NewForConfigOrDie(config)
	var list *unstructured.UnstructuredList
	list, err = d.Resource(resourceId).List(ctx, options)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func isMatch(resource *unstructured.Unstructured, script string, logger logr.Logger) (bool, error) {
	if script == "" {
		return true, nil
	}

	l := lua.NewState()
	defer l.Close()

	obj := mapToTable(resource.UnstructuredContent())

	if err := l.DoString(script); err != nil {
		logger.Info(fmt.Sprintf("doString failed: %v", err))
		return false, err
	}

	l.SetGlobal("obj", obj)

	if err := l.CallByParam(lua.P{
		Fn:      l.GetGlobal("evaluate"), // name of Lua function
		NRet:    1,                       // number of returned values
		Protect: true,                    // return err or panic
	}, obj); err != nil {
		logger.Info(fmt.Sprintf("failed to evaluate health for resource: %v", err))
		return false, err
	}

	lv := l.Get(-1)
	tbl, ok := lv.(*lua.LTable)
	if !ok {
		logger.Info(luaTableError)
		return false, fmt.Errorf("%s", luaTableError)
	}

	goResult := toGoValue(tbl)
	resultJson, err := json.Marshal(goResult)
	if err != nil {
		logger.Info(fmt.Sprintf("failed to marshal result: %v", err))
		return false, err
	}

	var result evaluateStatus
	err = json.Unmarshal(resultJson, &result)
	if err != nil {
		logger.Info(fmt.Sprintf("failed to marshal result: %v", err))
		return false, err
	}

	if result.Message != "" {
		logger.Info(fmt.Sprintf("message: %s", result.Message))
	}

	logger.V(logs.LogDebug).Info(fmt.Sprintf("is a match: %t", result.Matching))

	return result.Matching, nil
}

func Transform(resource *unstructured.Unstructured, script string, logger logr.Logger,
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

func getPrunerInstance(ctx context.Context, prunerName string) (*appsv1alpha1.Pruner, error) {
	pruner := &appsv1alpha1.Pruner{}
	err := k8sClient.Get(ctx, types.NamespacedName{Name: prunerName}, pruner)

	if apierrors.IsNotFound(err) {
		err = nil
	}

	return pruner, err
}

// storeResult does following:
// - set results for further in time lookup
// - remove request from inProgress
// - if request is in dirty, remove it from there and add it to the back of the jobQueue
func storeResult(prunerName string, err error, logger logr.Logger) {
	managerInstance.mu.Lock()
	defer managerInstance.mu.Unlock()

	key := prunerName

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
func getRequestStatus(prunerName string) (*responseParams, error) {
	logger := managerInstance.log.WithValues("pruner", prunerName)
	managerInstance.mu.Lock()
	defer managerInstance.mu.Unlock()

	key := prunerName

	logger.V(logs.LogDebug).Info("searching result")
	if _, ok := managerInstance.results[key]; ok {
		logger.V(logs.LogDebug).Info("request already processed, result present. returning result.")
		if managerInstance.results[key] != nil {
			logger.V(logs.LogDebug).Info("returning a response with an error")
		}
		resp := responseParams{
			prunerName: key,
			err:        managerInstance.results[key],
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
