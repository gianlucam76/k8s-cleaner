/*
Copyright 2025. projectsveltos.io. All rights reserved.

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
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"

	"gianlucam76/k8s-cleaner/internal/controller/executor"
)

// descFQName extracts the fully-qualified metric name from a collector's descriptor.
func descFQName(col prometheus.Collector) string {
	ch := make(chan *prometheus.Desc, 1)
	col.Describe(ch)
	return (<-ch).String()
}

var _ = Describe("metrics", func() {
	It("each counter vec getter returns its own distinct instance", func() {
		deleted := executor.GetDeletedResourcesCounterVec()
		updated := executor.GetUpdatedResourcesCounterVec()
		scan := executor.GetScanResourcesCounterVec()
		errVec := executor.GetErrorResourcesCounterVec()

		Expect(updated).ToNot(BeIdenticalTo(deleted))
		Expect(scan).ToNot(BeIdenticalTo(deleted))
		Expect(errVec).ToNot(BeIdenticalTo(deleted))
	})

	DescribeTable("metric names do not include a namespace prefix",
		func(col prometheus.Collector, expectedName string) {
			desc := descFQName(col)
			Expect(desc).To(ContainSubstring(fmt.Sprintf(`fqName: %q`, expectedName)))
		},
		Entry("deleted counter", executor.GetDeletedResourcesCounterVec(), "k8s_cleaner_deleted_resources_total"),
		Entry("updated counter", executor.GetUpdatedResourcesCounterVec(), "k8s_cleaner_updated_resources_total"),
		Entry("scan counter", executor.GetScanResourcesCounterVec(), "k8s_cleaner_scan_resources_total"),
		Entry("error counter", executor.GetErrorResourcesCounterVec(), "k8s_cleaner_error_resources_total"),
		Entry("deleted gauge", executor.GetDeletedResourcesGaugeVec(), "k8s_cleaner_current_deleted_resources_count"),
		Entry("updated gauge", executor.GetUpdatedResourcesGaugeVec(), "k8s_cleaner_current_updated_resources_count"),
		Entry("scan gauge", executor.GetScanResourcesGaugeVec(), "k8s_cleaner_current_resources_count"),
		Entry("error gauge", executor.GetErrorResourcesGaugeVec(), "k8s_cleaner_current_error_resources_count"),
	)
})
