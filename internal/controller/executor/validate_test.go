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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2/textlogger"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
	"gianlucam76/k8s-cleaner/internal/controller/executor"

	"github.com/projectsveltos/libsveltos/lib/k8s_utils"
)

const (
	fileName = "cleaner.yaml"

	matchingFileName    = "matching.yaml"
	nonMatchingFileName = "non-matching.yaml"
	updatedFileName     = "updated.yaml"
	allResourceFileName = "resources.yaml"
)

var _ = Describe("Cleaner with Lua", Label("VERIFY_LUA"), func() {
	It("Verify all resourceSelectors", func() {
		const eventDir = "./validate_resourceselector"

		dirs, err := os.ReadDir(eventDir)
		Expect(err).To(BeNil())

		for i := range dirs {
			if dirs[i].IsDir() {
				verifyCleanerResourceSelector(filepath.Join(eventDir, dirs[i].Name()))
			}
		}
	})

	It("Verify all transforms", func() {
		const eventDir = "./validate_transform"

		dirs, err := os.ReadDir(eventDir)
		Expect(err).To(BeNil())

		for i := range dirs {
			if dirs[i].IsDir() {
				verifyCleanerTransforms(filepath.Join(eventDir, dirs[i].Name()))
			}
		}
	})

	It("Verify all aggregatedselections", func() {
		const eventDir = "./validate_aggregatedselection"

		dirs, err := os.ReadDir(eventDir)
		Expect(err).To(BeNil())

		for i := range dirs {
			if dirs[i].IsDir() {
				verifyCleanerAggregatedSelections(filepath.Join(eventDir, dirs[i].Name()))
			}
		}
	})
})

func verifyCleanerResourceSelectors(dirName string) {
	By(fmt.Sprintf("Verifying verify Cleaner ResourceSelector in directory %s", dirName))

	dirs, err := os.ReadDir(dirName)
	Expect(err).To(BeNil())

	fileCount := 0

	for i := range dirs {
		if dirs[i].IsDir() {
			verifyCleanerResourceSelectors(fmt.Sprintf("%s/%s", dirName, dirs[i].Name()))
		} else {
			fileCount++
		}
	}

	if fileCount > 0 {
		verifyCleanerResourceSelector(dirName)
	}
}

func verifyCleanerResourceSelector(dirName string) {
	logger := textlogger.NewLogger(textlogger.NewConfig())

	files, err := os.ReadDir(dirName)
	Expect(err).To(BeNil())

	for i := range files {
		if files[i].IsDir() {
			verifyCleanerResourceSelectors(filepath.Join(dirName, files[i].Name()))
		}
	}

	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	executor.InitializeClient(context.TODO(), logger, nil, c, nil, nil, 10)
	client := executor.GetClient()
	Expect(client).ToNot(BeNil())

	By(fmt.Sprintf("Validating cleaner in dir: %s", dirName))
	cleaner := getCleaner(dirName)
	Expect(cleaner).ToNot(BeNil())

	matchingResource := getResource(dirName, matchingFileName)
	if matchingResource == nil {
		By(fmt.Sprintf("%s file not present", matchingFileName))
	} else {
		By("Verifying matching content")
		isMatch := false
		for i := range cleaner.Spec.ResourcePolicySet.ResourceSelectors {
			rs := &cleaner.Spec.ResourcePolicySet.ResourceSelectors[i]
			tmpIsMatch, _, err := executor.IsMatch(matchingResource, rs.Evaluate, logger)
			Expect(err).To(BeNil())
			if tmpIsMatch {
				isMatch = true
			}
		}
		Expect(isMatch).To(BeTrue())
	}

	nonMatchingResource := getResource(dirName, nonMatchingFileName)
	if nonMatchingResource == nil {
		By(fmt.Sprintf("%s file not present", nonMatchingFileName))
	} else {
		By("Verifying non-matching content")
		isMatch := false
		for i := range cleaner.Spec.ResourcePolicySet.ResourceSelectors {
			rs := &cleaner.Spec.ResourcePolicySet.ResourceSelectors[i]
			tmpIsMatch, _, err := executor.IsMatch(nonMatchingResource, rs.Evaluate, logger)
			Expect(err).To(BeNil())
			if tmpIsMatch {
				isMatch = true
			}
		}
		Expect(isMatch).To(BeFalse())
	}
}

func verifyCleanerTransforms(dirName string) {
	By(fmt.Sprintf("Verifying verify Cleaner Transforms in directory %s", dirName))

	dirs, err := os.ReadDir(dirName)
	Expect(err).To(BeNil())

	fileCount := 0

	for i := range dirs {
		if dirs[i].IsDir() {
			verifyCleanerTransforms(fmt.Sprintf("%s/%s", dirName, dirs[i].Name()))
		} else {
			fileCount++
		}
	}

	if fileCount > 0 {
		verifyCleanerTransform(dirName)
	}
}

func verifyCleanerTransform(dirName string) {
	logger := textlogger.NewLogger(textlogger.NewConfig())

	files, err := os.ReadDir(dirName)
	Expect(err).To(BeNil())

	for i := range files {
		if files[i].IsDir() {
			verifyCleanerTransforms(filepath.Join(dirName, files[i].Name()))
			continue
		}
	}

	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	executor.InitializeClient(context.TODO(), logger, nil, c, nil, nil, 10)
	client := executor.GetClient()
	Expect(client).ToNot(BeNil())

	By(fmt.Sprintf("Validating cleaner in dir: %s", dirName))
	cleaner := getCleaner(dirName)
	Expect(cleaner).ToNot(BeNil())
	matchingResource := getResource(dirName, matchingFileName)
	if matchingResource == nil {
		By(fmt.Sprintf("%s file not present", matchingFileName))
	} else {
		By("Verifying matching content")
		isMatch := false
		for i := range cleaner.Spec.ResourcePolicySet.ResourceSelectors {
			rs := &cleaner.Spec.ResourcePolicySet.ResourceSelectors[i]
			var tmpIsMatch bool
			tmpIsMatch, _, err = executor.IsMatch(matchingResource, rs.Evaluate, logger)
			Expect(err).To(BeNil())
			if tmpIsMatch {
				isMatch = true
			}
		}
		Expect(isMatch).To(BeTrue())
	}

	var updatedResource *unstructured.Unstructured
	updatedResource, err = executor.Transform(matchingResource, cleaner.Spec.Transform, logger)
	Expect(err).To(BeNil())

	expectedUpdatedResource := getResource(dirName, updatedFileName)
	if expectedUpdatedResource == nil {
		By(fmt.Sprintf("%s file not present", updatedFileName))
	} else {
		gvk := updatedResource.GroupVersionKind()
		updatedTypedObj, err := scheme.New(gvk)
		Expect(err).To(BeNil())
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(updatedResource.Object, updatedTypedObj)
		Expect(err).To(BeNil())

		gvk = expectedUpdatedResource.GroupVersionKind()
		expectedTypedObj, err := scheme.New(gvk)
		Expect(err).To(BeNil())
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(expectedUpdatedResource.Object, expectedTypedObj)
		Expect(err).To(BeNil())

		diff := cmp.Diff(updatedTypedObj, expectedTypedObj)
		Expect(diff).To(BeEmpty())
	}
}

func verifyCleanerAggregatedSelections(dirName string) {
	By(fmt.Sprintf("Verifying verify Cleaner Transforms in directory %s", dirName))

	dirs, err := os.ReadDir(dirName)
	Expect(err).To(BeNil())

	fileCount := 0

	for i := range dirs {
		if dirs[i].IsDir() {
			verifyCleanerAggregatedSelections(fmt.Sprintf("%s/%s", dirName, dirs[i].Name()))
		} else {
			fileCount++
		}
	}

	if fileCount > 0 {
		verifyCleanerAggregatedSelection(dirName)
	}
}

func verifyCleanerAggregatedSelection(dirName string) {
	logger := textlogger.NewLogger(textlogger.NewConfig())

	files, err := os.ReadDir(dirName)
	Expect(err).To(BeNil())

	for i := range files {
		if files[i].IsDir() {
			verifyCleanerAggregatedSelections(filepath.Join(dirName, files[i].Name()))
			continue
		}
	}

	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	executor.InitializeClient(context.TODO(), logger, nil, c, nil, nil, 10)
	client := executor.GetClient()
	Expect(client).ToNot(BeNil())

	By(fmt.Sprintf("Validating cleaner in dir: %s", dirName))
	cleaner := getCleaner(dirName)
	Expect(cleaner).ToNot(BeNil())

	var result []executor.ResourceResult
	resources := getResources(dirName, allResourceFileName)
	matchingResources := getResources(dirName, matchingFileName)
	if resources == nil {
		By(fmt.Sprintf("%s file not present", matchingFileName))
	} else {
		result, err = executor.AggregatedSelection(cleaner.Spec.ResourcePolicySet.AggregatedSelection,
			resources, logger)
		Expect(err).To(BeNil())
		verifyMatchingResources(result, matchingResources)
	}

	Expect(len(matchingResources) < len(resources))
	for i := range resources {
		if !isPresent(resources[i], matchingResources) {
			// resource is supposed to be non matching. Verify it is not present
			// in the cleaner result
			Expect(isPresent(resources[i], result)).To(BeFalse())
		}
	}
}

func getCleaner(dirName string) *appsv1alpha1.Cleaner {
	cleanerFileName := filepath.Join(dirName, fileName)
	content, err := os.ReadFile(cleanerFileName)
	Expect(err).To(BeNil())

	u, err := k8s_utils.GetUnstructured(content)
	Expect(err).To(BeNil())

	var cleaner appsv1alpha1.Cleaner
	err = runtime.DefaultUnstructuredConverter.
		FromUnstructured(u.UnstructuredContent(), &cleaner)
	Expect(err).To(BeNil())
	return &cleaner
}

func getResource(dirName, fileName string) *unstructured.Unstructured {
	resourceFileName := filepath.Join(dirName, fileName)

	_, err := os.Stat(resourceFileName)
	if os.IsNotExist(err) {
		return nil
	}
	Expect(err).To(BeNil())

	content, err := os.ReadFile(resourceFileName)
	Expect(err).To(BeNil())

	u, err := k8s_utils.GetUnstructured(content)
	Expect(err).To(BeNil())

	return u
}

func getResources(dirName, fileName string) []executor.ResourceResult {
	resourceFileName := filepath.Join(dirName, fileName)

	_, err := os.Stat(resourceFileName)
	if os.IsNotExist(err) {
		return nil
	}
	Expect(err).To(BeNil())

	content, err := os.ReadFile(resourceFileName)
	Expect(err).To(BeNil())

	resources := make([]executor.ResourceResult, 0)
	elements := strings.Split(string(content), "---")
	for i := range elements {
		u, err := k8s_utils.GetUnstructured([]byte(elements[i]))
		Expect(err).To(BeNil())
		resources = append(resources, executor.ResourceResult{
			Resource: u,
		})
	}

	return resources
}

func getKey(u *unstructured.Unstructured) string {
	return fmt.Sprintf("%s:%s/%s", u.GetKind(), u.GetNamespace(), u.GetName())
}

func verifyMatchingResources(result, matchingResources []executor.ResourceResult) {
	// This is used to keep track of resources that are expected to match
	expected := map[string]bool{}

	for i := range matchingResources {
		By(fmt.Sprintf("Verify matchingResources %s %s:%s",
			matchingResources[i].Resource.GroupVersionKind().Kind,
			matchingResources[i].Resource.GetNamespace(), matchingResources[i].Resource.GetName()))
		Expect(isPresent(matchingResources[i], result)).To(BeTrue())
		expected[getKey(matchingResources[i].Resource)] = true
	}

	// verify only expected matching objects are found
	for i := range result {
		key := getKey(result[i].Resource)
		if ok := expected[key]; !ok {
			// Print the resource that is not expected to be a match
			Expect(key).To(BeEmpty())
		}
	}
}

func isPresent(r executor.ResourceResult, resources []executor.ResourceResult) bool {
	for i := range resources {
		if r.Resource.GroupVersionKind().Kind == resources[i].Resource.GroupVersionKind().Kind &&
			r.Resource.GetNamespace() == resources[i].Resource.GetNamespace() &&
			r.Resource.GetName() == resources[i].Resource.GetName() {

			return true
		}
	}

	return false
}
