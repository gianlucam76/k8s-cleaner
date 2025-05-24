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
	"fmt"
	"sync"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	logs "github.com/projectsveltos/libsveltos/lib/logsettings"
)

var (
	getClientLock   = &sync.Mutex{}
	managerInstance *Manager
)

const (
	unavailable = "unavailable"
)

type ResultStatus int64

const (
	Processed ResultStatus = iota
	InProgress
	Failed
	Unavailable
)

func (r ResultStatus) String() string {
	switch r {
	case Processed:
		return "processed"
	case InProgress:
		return "in-progress"
	case Failed:
		return "failed"
	case Unavailable:
		return unavailable
	}
	return unavailable
}

type Result struct {
	ResultStatus
	Message string
	Err     error
}

type Manager struct {
	log logr.Logger
	client.Client
	config *rest.Config
	scheme *runtime.Scheme

	mu *sync.Mutex

	// A request represents a request to process a Cleaner instance

	// dirty contains all requests (cleaner names) which are currently waiting to be served.
	dirty []string

	// inProgress contains all requests (cleaner names) that are currently being served.
	inProgress []string

	// jobQueue contains all requests (cleaner names) that needs to be served.
	jobQueue []string

	// results contains results for processed requests (cleaner names)
	results map[string]error

	eventRecorder record.EventRecorder
}

// InitializeClient initializes a client
func InitializeClient(ctx context.Context, l logr.Logger, config *rest.Config,
	c client.Client, scheme *runtime.Scheme, eventRecorder record.EventRecorder, numOfWorker int) {

	if managerInstance == nil {
		getClientLock.Lock()
		defer getClientLock.Unlock()
		if managerInstance == nil {
			// Create a zap logger
			logger, err := zap.NewDevelopment()
			if err != nil {
				panic(1)
			}

			l.V(logs.LogInfo).Info(fmt.Sprintf("Creating instance now. Number of workers: %d", numOfWorker))
			managerInstance = &Manager{log: l, Client: c, config: config}
			managerInstance.scheme = scheme
			managerInstance.log = zapr.NewLogger(logger)
			managerInstance.startWorkloadWorkers(ctx, numOfWorker, l)
			managerInstance.eventRecorder = eventRecorder
		}
	}
}

// startWorkloadWorkers initializes all internal structures and starts
// pool of workers
// - numWorker is number of requested workers
// - c is the kubernetes client to access control cluster
func (m *Manager) startWorkloadWorkers(ctx context.Context, numOfWorker int, logger logr.Logger) {
	m.mu = &sync.Mutex{}
	m.dirty = make([]string, 0)
	m.inProgress = make([]string, 0)
	m.jobQueue = make([]string, 0)
	m.results = make(map[string]error)
	k8sClient = m.Client
	config = m.config
	scheme = m.scheme

	for i := 0; i < numOfWorker; i++ {
		go processRequests(ctx, i, logger.WithValues("worker", fmt.Sprintf("%d", i)))
	}
}

// GetClient return a collector client
func GetClient() *Manager {
	getClientLock.Lock()
	defer getClientLock.Unlock()
	return managerInstance
}

func (m *Manager) Process(ctx context.Context, cleanerName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	l := m.log.WithValues("cleaner", cleanerName)
	key := cleanerName

	// Search if request is in dirty. Drop it if already there
	for i := range m.dirty {
		if m.dirty[i] == key {
			l.V(logs.LogDebug).Info("request is already present in dirty")
			return
		}
	}

	// Since we got a new request, if a result was saved, clear it.
	l.V(logs.LogDebug).Info("removing result from previous request if any")
	delete(m.results, key)

	m.log.V(logs.LogDebug).Info("request added to dirty")
	m.dirty = append(m.dirty, key)

	// Push to queue if not already in progress
	for i := range m.inProgress {
		if m.inProgress[i] == key {
			m.log.V(logs.LogDebug).Info("request is already in inProgress")
			return
		}
	}

	m.log.V(logs.LogDebug).Info("request added to jobQueue")
	m.jobQueue = append(m.jobQueue, cleanerName)
}

func (m *Manager) GetResult(cleanerName string) Result {
	responseParam, err := getRequestStatus(cleanerName)
	if err != nil {
		return Result{
			ResultStatus: Unavailable,
			Err:          nil,
		}
	}

	if responseParam == nil {
		return Result{
			ResultStatus: InProgress,
			Err:          nil,
		}
	}

	if responseParam.err != nil {
		return Result{
			ResultStatus: Failed,
			Err:          responseParam.err,
		}
	}

	return Result{
		ResultStatus: Processed,
	}
}

func (m *Manager) RemoveEntries(cleanerName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := cleanerName

	for i := range m.inProgress {
		if m.inProgress[i] == key {
			removeFromSlice(m.inProgress, i)
			break
		}
	}

	for i := range m.dirty {
		if m.dirty[i] == key {
			removeFromSlice(m.dirty, i)
			break
		}
	}

	for i := range m.jobQueue {
		if m.jobQueue[i] == key {
			removeFromSlice(m.jobQueue, i)
			break
		}
	}

	delete(m.results, key)
}
