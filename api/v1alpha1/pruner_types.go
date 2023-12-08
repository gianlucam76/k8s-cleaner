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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	libsveltosv1alpha1 "github.com/projectsveltos/libsveltos/api/v1alpha1"
)

// Action specifies the action to take on matching resources
// +kubebuilder:validation:Enum:=Delete;Transform
type Action string

const (
	// ActionDelete will delete the resource
	ActionDelete = Action("Delete")

	// ActionTransform will update object
	ActionTransform = Action("Transform")
)

const (
	// PrunerFinalizer allows Reconciler to clean up resources associated with
	// Pruner instance before removing it from the apiserver.
	PrunerFinalizer = "prunerfinalizer.projectsveltos.io"
)

type Resources struct {
	// Namespace of the resource deployed in the  Cluster.
	// Empty for resources scoped at cluster level.
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Group of the resource deployed in the Cluster.
	Group string `json:"group"`

	// Version of the resource deployed in the Cluster.
	Version string `json:"version"`

	// Kind of the resource deployed in the Cluster.
	// +kubebuilder:validation:MinLength=1
	Kind string `json:"kind"`

	// LabelFilters allows to filter resources based on current labels.
	LabelFilters []libsveltosv1alpha1.LabelFilter `json:"labelFilters,omitempty"`

	// Evaluate contains a function "evaluate" in lua language.
	// The function will be passed one of the object selected based on
	// above criteria.
	// Must return struct with field "matching"
	// representing whether object is a match.
	// +optional
	Evaluate string `json:"evaluate,omitempty"`

	// Action indicates the action to take. Default action
	// is to delete object. If set to transform, the transform function
	// will be invoked and then object will be updated.
	// +kubebuilder:default:=Delete
	Action Action `json:"action,omitempty"`

	// Transform contains a function "transform" in lua language.
	// When Action is set to *Transform*, this function will be invoked
	// and be passed one of the object selected based on
	// above criteria.
	// Must the new object that will be applied
	// +optional
	Transform string `json:"transform,omitempty"`
}

// PrunerSpec defines the desired state of Pruner
type PrunerSpec struct {
	StaleResources []Resources `json:"staleResources"`

	// Schedule in Cron format, see https://en.wikipedia.org/wiki/Cron.
	Schedule string `json:"schedule"`

	// Optional deadline in seconds for starting the job if it misses scheduled
	// time for any reason.  Missed jobs executions will be counted as failed ones.
	// +optional
	StartingDeadlineSeconds *int64 `json:"startingDeadlineSeconds,omitempty"`

	// DryRun if set to true, will have controller delete no resource.
	// All matching resources will be listed in status section
	// +kubebuilder:default:=false
	// +optional
	DryRun bool `json:"dryRune,omitempty"`
}

// PrunerStatus defines the observed state of Pruner
type PrunerStatus struct {
	// Information when next snapshot is scheduled
	// +optional
	NextScheduleTime *metav1.Time `json:"nextScheduleTime,omitempty"`

	// Information when was the last time a snapshot was successfully scheduled.
	// +optional
	LastRunTime *metav1.Time `json:"lastRunTime,omitempty"`

	// FailureMessage provides more information about the error, if
	// any occurred
	FailureMessage *string `json:"failureMessage,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:path=pruners,scope=Cluster
//+kubebuilder:subresource:status

// Pruner is the Schema for the pruners API
type Pruner struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PrunerSpec   `json:"spec,omitempty"`
	Status PrunerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PrunerList contains a list of Pruner
type PrunerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Pruner `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Pruner{}, &PrunerList{})
}
