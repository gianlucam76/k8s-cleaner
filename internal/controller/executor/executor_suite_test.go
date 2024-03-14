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
	"gianlucam76/k8s-cleaner/internal/controller/executor"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var (
	testEnv   *envtest.Environment
	config    *rest.Config
	k8sClient client.Client
	scheme    *runtime.Scheme
)

var (
	cacheSyncBackoff = wait.Backoff{
		Duration: 100 * time.Millisecond,
		Factor:   1.5,
		Steps:    8,
		Jitter:   0.4,
	}
)

func TestExecutor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Executor Suite")
}

var _ = BeforeSuite(func() {
	By("bootstrapping test environment")

	var err error
	scheme, err = setupScheme()
	Expect(err).To(BeNil())

	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "..", "..", "config", "crd", "bases")},
	}

	config, err = testEnv.Start()
	Expect(err).To(BeNil())

	time.Sleep(time.Second)

	k8sClient, err = client.New(config, client.Options{Scheme: scheme})
	Expect(err).To(BeNil())

	// Create a zap logger
	logger, err := zap.NewDevelopment()
	Expect(err).To(BeNil())

	executor.InitializeClient(context.TODO(), zapr.NewLogger(logger), config, k8sClient, scheme, 10)

	By("bootstrapping completed")
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

func setupScheme() (*runtime.Scheme, error) {
	s := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(s); err != nil {
		return nil, err
	}

	return s, nil
}

func randomString() string {
	const length = 10
	return util.RandomString(length)
}

// waitForObject waits for the cache to be updated helps in preventing test flakes due to the cache sync delays.
func waitForObject(ctx context.Context, c client.Client, obj client.Object) error {
	// Makes sure the cache is updated with the new object
	objCopy := obj.DeepCopyObject().(client.Object)
	key := client.ObjectKeyFromObject(obj)
	if err := wait.ExponentialBackoff(
		cacheSyncBackoff,
		func() (done bool, err error) {
			if err := c.Get(ctx, key, objCopy); err != nil {
				if apierrors.IsNotFound(err) {
					return false, nil
				}
				return false, err
			}
			return true, nil
		}); err != nil {
		return errors.Wrapf(err, "object %s, %s is not being added to the testenv client cache", obj.GetObjectKind().GroupVersionKind().String(), key)
	}
	return nil
}
