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

	appsv1alpha1 "gianlucam76/k8s-pruner/api/v1alpha1"
)

// PrunerScopeParams defines the input parameters used to create a new Pruner Scope.
type PrunerScopeParams struct {
	Client         client.Client
	Logger         logr.Logger
	Pruner         *appsv1alpha1.Pruner
	ControllerName string
}

// NewPrunerScope creates a new Pruner Scope from the supplied parameters.
// This is meant to be called for each reconcile iteration.
func NewPrunerScope(params PrunerScopeParams) (*PrunerScope, error) {
	if params.Client == nil {
		return nil, errors.New("client is required when creating a PrunerScope")
	}
	if params.Pruner == nil {
		return nil, errors.New("failed to generate new scope from nil Pruner")
	}

	helper, err := patch.NewHelper(params.Pruner, params.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}
	return &PrunerScope{
		Logger:         params.Logger,
		client:         params.Client,
		Pruner:         params.Pruner,
		patchHelper:    helper,
		controllerName: params.ControllerName,
	}, nil
}

// PrunerScope defines the basic context for an actuator to operate upon.
type PrunerScope struct {
	logr.Logger
	client         client.Client
	patchHelper    *patch.Helper
	Pruner         *appsv1alpha1.Pruner
	controllerName string
}

// PatchObject persists the feature configuration and status.
func (s *PrunerScope) PatchObject(ctx context.Context) error {
	return s.patchHelper.Patch(
		ctx,
		s.Pruner,
	)
}

// Close closes the current scope persisting the Pruner configuration and status.
func (s *PrunerScope) Close(ctx context.Context) error {
	return s.PatchObject(ctx)
}

// SetLastRunTime set LastRunTime field
func (s *PrunerScope) SetLastRunTime(lastRunTime *metav1.Time) {
	s.Pruner.Status.LastRunTime = lastRunTime
}

// SetNextScheduleTime sets NextScheduleTime field
func (s *PrunerScope) SetNextScheduleTime(lastRunTime *metav1.Time) {
	s.Pruner.Status.NextScheduleTime = lastRunTime
}

// SetFailureMessage sets FasilureMessage field
func (s *PrunerScope) SetFailureMessage(failureMessage *string) {
	s.Pruner.Status.FailureMessage = failureMessage
}
