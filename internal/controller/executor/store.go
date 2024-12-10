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
	"fmt"
	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
	"os"
	"path"
	"path/filepath"

	"github.com/go-logr/logr"
	"gopkg.in/yaml.v3"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	logs "github.com/projectsveltos/libsveltos/lib/logsettings"
)

const (
	permission0600 = 0600
	permission0644 = 0644
	permission0755 = 0755
)

func storeResources(processedResources []ResourceResult, scheme *runtime.Scheme, cleaner *appsv1alpha1.Cleaner,
	logger logr.Logger) error {

	if cleaner.Spec.StoreResourcePath == "" {
		return nil
	}

	folder, err := getFolder(cleaner.Spec.StoreResourcePath, cleaner.Name, logger)
	if err != nil {
		return err
	}

	for i := range processedResources {
		err = dumpObject(processedResources[i].Resource, scheme, *folder, logger)
		if err != nil {
			logger.V(logs.LogInfo).Info(fmt.Sprintf("failed to store object %s %s/%s: %v",
				processedResources[i].Resource.GetKind(), processedResources[i].Resource.GetNamespace(),
				processedResources[i].Resource.GetName(), err))
			// Error is ignored as Cleaner tries to store other resources
		}
	}

	return nil
}

func getFolder(storage, cleanerName string, logger logr.Logger) (*string, error) {
	l := logger.WithValues("cleaner", cleanerName)
	l.V(logs.LogDebug).Info("getting directory containing collections for instance")

	if _, err := os.Stat(storage); os.IsNotExist(err) {
		logger.V(logs.LogInfo).Info(fmt.Sprintf("directory %s not found", storage))
		return nil, err
	}

	artifactFolder := filepath.Join(storage, cleanerName)

	return &artifactFolder, nil
}

// dumpObject is a helper function to generically dump resource definition
// given the resource reference and file path for dumping location.
func dumpObject(resource *unstructured.Unstructured, scheme *runtime.Scheme, logPath string, logger logr.Logger) error {
	// Do not store resource version
	resource.SetResourceVersion("")
	err := addTypeInformationToObject(scheme, resource)
	if err != nil {
		return err
	}

	logger = logger.WithValues("kind", resource.GetObjectKind())
	logger = logger.WithValues("resource", fmt.Sprintf("%s %s",
		resource.GetNamespace(), resource.GetName()))

	if !resource.GetDeletionTimestamp().IsZero() {
		logger.V(logs.LogDebug).Info("resource is marked for deletion. Do not collect it.")
	}

	resourceYAML, err := yaml.Marshal(resource.UnstructuredContent())
	if err != nil {
		return err
	}

	metaObj, err := apimeta.Accessor(resource)
	if err != nil {
		return err
	}

	kind := resource.GetObjectKind().GroupVersionKind().Kind
	namespace := metaObj.GetNamespace()
	name := metaObj.GetName()

	resourceFilePath := path.Join(logPath, namespace, kind, name+".yaml")
	err = os.MkdirAll(filepath.Dir(resourceFilePath), permission0755)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(resourceFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, permission0644)
	if err != nil {
		return err
	}
	defer f.Close()

	logger.V(logs.LogDebug).Info(fmt.Sprintf("storing resource in %s", resourceFilePath))
	return os.WriteFile(f.Name(), resourceYAML, permission0600)
}

func addTypeInformationToObject(scheme *runtime.Scheme, obj client.Object) error {
	gvks, _, err := scheme.ObjectKinds(obj)
	if err != nil {
		return err
	}

	for _, gvk := range gvks {
		if gvk.Kind == "" {
			continue
		}
		if gvk.Version == "" || gvk.Version == runtime.APIVersionInternal {
			continue
		}
		obj.GetObjectKind().SetGroupVersionKind(gvk)
		break
	}

	return nil
}
