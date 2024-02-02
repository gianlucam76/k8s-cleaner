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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ResourceInfo struct {
	// Resource identify a Kubernetes resource
	Resource corev1.ObjectReference `json:"resource,omitempty"`

	// FullResource contains full resources before
	// before Cleaner took an action on it
	// +optional
	FullResource []byte `json:"fullResource,omitempty"`

	// Message is an optional field.
	// +optional
	Message string `json:"message,omitempty"`
}

// ReportSpec defines the desired state of Report
type ReportSpec struct {
	// Resources identify a set of Kubernetes resource
	ResourceInfo []ResourceInfo `json:"resourceInfo"`

	// Action indicates the action to take on selected object.
	Action Action `json:"action"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:path=reports,scope=Cluster

// Report is the Schema for the reports API
type Report struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ReportSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// ReportList contains a list of Report
type ReportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Report `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Report{}, &ReportList{})
}
