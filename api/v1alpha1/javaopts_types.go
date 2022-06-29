/*
Copyright 2022.

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
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// JavaOptsSpec defines the desired state of JavaOpts
type JavaOptsSpec struct {
	//JVMName       string `json:"jvmName,omitempty"`
	Command       string `json:"command"`
	RevName       string `json:"revName"`
	DriverImage   string `json:"driverImage,omitempty"`
	ConfigMapName string `json:"configMapName,omitempty"`
	GCType        string `json:"gcType,omitempty"`
	MaxHeap       string `json:"maxHeapSize,omitempty"`
	MinHeap       string `json:"minHeapSize,omitempty"`
	GCThreads     string `json:"gcThreads,omitempty"`
	Escape        string `json:"escapeAnalysis,omitempty"`
	DefaultOpts   string `json:"defaultOpts,omitempty"`
}

// JavaOptsStatus defines the observed state of JavaOpts
type JavaOptsStatus struct {
	Conditions     []batchv1.JobCondition `json:"conditions"`
	StartTime      *metav1.Time           `json:"startTime,omitempty"`
	CompletionTime *metav1.Time           `json:"completionTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// JavaOpts is the Schema for the javaopts API
type JavaOpts struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JavaOptsSpec   `json:"spec,omitempty"`
	Status JavaOptsStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// JavaOptsList contains a list of JavaOpts
type JavaOptsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JavaOpts `json:"items"`
}

func init() {
	SchemeBuilder.Register(&JavaOpts{}, &JavaOptsList{})
}
