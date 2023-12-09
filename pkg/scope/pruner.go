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

package scope

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "gianlucam76/k8s-cleaner/api/v1alpha1"
)

// CleanerScopeParams defines the input parameters used to create a new Cleaner Scope.
type CleanerScopeParams struct {
	Client         client.Client
	Logger         logr.Logger
	Cleaner        *appsv1alpha1.Cleaner
	ControllerName string
}

// NewCleanerScope creates a new Cleaner Scope from the supplied parameters.
// This is meant to be called for each reconcile iteration.
func NewCleanerScope(params CleanerScopeParams) (*CleanerScope, error) {
	if params.Client == nil {
		return nil, errors.New("client is required when creating a CleanerScope")
	}
	if params.Cleaner == nil {
		return nil, errors.New("failed to generate new scope from nil Cleaner")
	}

	helper, err := patch.NewHelper(params.Cleaner, params.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}
	return &CleanerScope{
		Logger:         params.Logger,
		client:         params.Client,
		Cleaner:        params.Cleaner,
		patchHelper:    helper,
		controllerName: params.ControllerName,
	}, nil
}

// CleanerScope defines the basic context for an actuator to operate upon.
type CleanerScope struct {
	logr.Logger
	client         client.Client
	patchHelper    *patch.Helper
	Cleaner        *appsv1alpha1.Cleaner
	controllerName string
}

// PatchObject persists the feature configuration and status.
func (s *CleanerScope) PatchObject(ctx context.Context) error {
	return s.patchHelper.Patch(
		ctx,
		s.Cleaner,
	)
}

// Close closes the current scope persisting the Cleaner configuration and status.
func (s *CleanerScope) Close(ctx context.Context) error {
	return s.PatchObject(ctx)
}

// SetLastRunTime set LastRunTime field
func (s *CleanerScope) SetLastRunTime(lastRunTime *metav1.Time) {
	s.Cleaner.Status.LastRunTime = lastRunTime
}

// SetNextScheduleTime sets NextScheduleTime field
func (s *CleanerScope) SetNextScheduleTime(lastRunTime *metav1.Time) {
	s.Cleaner.Status.NextScheduleTime = lastRunTime
}

// SetFailureMessage sets FasilureMessage field
func (s *CleanerScope) SetFailureMessage(failureMessage *string) {
	s.Cleaner.Status.FailureMessage = failureMessage
}
