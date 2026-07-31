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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
	"gianlucam76/k8s-cleaner/internal/controller/executor"
)

var _ = Describe("Blast Radius Limit", func() {
	It("checkBlastRadiusLimit allows the run when no limit is set", func() {
		Expect(executor.CheckBlastRadiusLimit(nil, 1000, 1000)).To(Succeed())
	})

	It("checkBlastRadiusLimit allows the run when matched count is within MaxCount", func() {
		maxCount := 10
		limit := &appsv1alpha1.BlastRadiusLimit{MaxCount: &maxCount}
		Expect(executor.CheckBlastRadiusLimit(limit, 10, 10)).To(Succeed())
	})

	It("checkBlastRadiusLimit aborts the run when matched count exceeds MaxCount", func() {
		maxCount := 10
		limit := &appsv1alpha1.BlastRadiusLimit{MaxCount: &maxCount}
		err := executor.CheckBlastRadiusLimit(limit, 11, 100)
		Expect(err).To(HaveOccurred())
	})

	It("checkBlastRadiusLimit allows the run when matched percentage is within MaxPercentage", func() {
		maxPercentage := 50
		limit := &appsv1alpha1.BlastRadiusLimit{MaxPercentage: &maxPercentage}
		Expect(executor.CheckBlastRadiusLimit(limit, 5, 10)).To(Succeed())
	})

	It("checkBlastRadiusLimit aborts the run when matched percentage exceeds MaxPercentage", func() {
		maxPercentage := 50
		limit := &appsv1alpha1.BlastRadiusLimit{MaxPercentage: &maxPercentage}
		err := executor.CheckBlastRadiusLimit(limit, 6, 10)
		Expect(err).To(HaveOccurred())
	})

	It("checkBlastRadiusLimit ignores MaxPercentage when totalScanned is zero", func() {
		maxPercentage := 50
		limit := &appsv1alpha1.BlastRadiusLimit{MaxPercentage: &maxPercentage}
		Expect(executor.CheckBlastRadiusLimit(limit, 0, 0)).To(Succeed())
	})

	It("checkBlastRadiusLimit aborts when either MaxCount or MaxPercentage is exceeded", func() {
		maxCount := 100
		maxPercentage := 10
		limit := &appsv1alpha1.BlastRadiusLimit{MaxCount: &maxCount, MaxPercentage: &maxPercentage}
		// under MaxCount, but over MaxPercentage
		err := executor.CheckBlastRadiusLimit(limit, 20, 100)
		Expect(err).To(HaveOccurred())
	})
})
