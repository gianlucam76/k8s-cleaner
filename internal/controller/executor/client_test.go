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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"gianlucam76/k8s-cleaner/internal/controller/executor"
)

var _ = Describe("CleanerClient", func() {
	It("GetResult returns result when available", func() {
		cleanerName := randomString()

		d := executor.GetClient()
		defer d.ClearInternalStruct()

		r := map[string]error{cleanerName: nil}
		d.SetResults(r)
		Expect(len(d.GetResults())).To(Equal(1))

		result := d.GetResult(cleanerName)
		Expect(result.Err).To(BeNil())
		Expect(result.ResultStatus).To(Equal(executor.Processed))
	})

	It("GetResult returns result when available with error", func() {
		cleanerName := randomString()

		d := executor.GetClient()
		defer d.ClearInternalStruct()

		r := map[string]error{cleanerName: fmt.Errorf("failed to deploy")}
		d.SetResults(r)
		Expect(len(d.GetResults())).To(Equal(1))

		result := d.GetResult(cleanerName)
		Expect(result.Err).ToNot(BeNil())
		Expect(result.ResultStatus).To(Equal(executor.Failed))
	})

	It("GetResult returns InProgress when request is still queued (currently in progress)", func() {
		cleanerName := randomString()

		d := executor.GetClient()
		defer d.ClearInternalStruct()

		d.SetInProgress([]string{cleanerName})
		Expect(len(d.GetInProgress())).To(Equal(1))

		result := d.GetResult(cleanerName)
		Expect(result.Err).To(BeNil())
		Expect(result.ResultStatus).To(Equal(executor.InProgress))
	})

	It("GetResult returns InProgress when request is still queued (currently queued)", func() {
		cleanerName := randomString()

		d := executor.GetClient()
		defer d.ClearInternalStruct()

		d.SetJobQueue(cleanerName)
		Expect(len(d.GetJobQueue())).To(Equal(1))

		result := d.GetResult(cleanerName)
		Expect(result.Err).To(BeNil())
		Expect(result.ResultStatus).To(Equal(executor.InProgress))
	})

	It("GetResult returns Unavailable when request is not queued/in progress and result not available", func() {
		cleanerName := randomString()

		d := executor.GetClient()
		defer d.ClearInternalStruct()

		result := d.GetResult(cleanerName)
		Expect(result.Err).To(BeNil())
		Expect(result.ResultStatus).To(Equal(executor.Unavailable))
	})

	It("Process does nothing if already in the dirty set", func() {
		cleanerName := randomString()

		d := executor.GetClient()
		defer d.ClearInternalStruct()

		d.SetDirty([]string{cleanerName})
		Expect(len(d.GetDirty())).To(Equal(1))

		d.Process(context.TODO(), cleanerName)
		Expect(len(d.GetDirty())).To(Equal(1))
		Expect(len(d.GetInProgress())).To(Equal(0))
		Expect(len(d.GetJobQueue())).To(Equal(0))
	})

	It("Process adds to inProgress", func() {
		cleanerName := randomString()

		d := executor.GetClient()
		defer d.ClearInternalStruct()

		d.Process(context.TODO(), cleanerName)
		Expect(len(d.GetDirty())).To(Equal(1))
		Expect(len(d.GetInProgress())).To(Equal(0))
		Expect(len(d.GetJobQueue())).To(Equal(1))
	})

	It("Process if already in progress, does not add to jobQueue", func() {
		cleanerName := randomString()

		d := executor.GetClient()
		defer d.ClearInternalStruct()

		d.SetInProgress([]string{cleanerName})
		Expect(len(d.GetInProgress())).To(Equal(1))

		d.Process(context.TODO(), cleanerName)
		Expect(len(d.GetDirty())).To(Equal(1))
		Expect(len(d.GetInProgress())).To(Equal(1))
		Expect(len(d.GetJobQueue())).To(Equal(0))
	})

	It("Process removes existing result", func() {
		cleanerName := randomString()

		d := executor.GetClient()
		defer d.ClearInternalStruct()

		r := map[string]error{cleanerName: nil}
		d.SetResults(r)
		Expect(len(d.GetResults())).To(Equal(1))

		d.Process(context.TODO(), cleanerName)
		Expect(len(d.GetDirty())).To(Equal(1))
		Expect(len(d.GetInProgress())).To(Equal(0))
		Expect(len(d.GetJobQueue())).To(Equal(1))
		Expect(len(d.GetResults())).To(Equal(0))
	})
})
