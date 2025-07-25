/*
Copyright 2025.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ConfigMapSyncSpec defines the desired state of ConfigMapSync
type ConfigMapSyncSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// foo is an example field of ConfigMapSync. Edit configmapsync_types.go to remove/update
	// +optional
	SourceNamespace      string `json:"sourceNamespace"`
	DestinationNamespace string `json:"destinationNamespace"`
	ConfigMapName        string `json:"configMapName"`
}

// ConfigMapSyncStatus defines the observed state of ConfigMapSync.
type ConfigMapSyncStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	LastSyncTime      string `json:"lastSyncTime,omitempty"`
	SyncStatus        string `json:"syncStatus,omitempty"` // "Success", "Failed", "InProgress"
	Message           string `json:"message,omitempty"`    // Human readable message
	SourceExists      bool   `json:"sourceExists"`         // Source ConfigMap found
	DestinationExists bool   `json:"destinationExists"`    // Destination ConfigMap exists

	Conditions []metav1.Condition `json:"conditions,omitempty"`
	RetryCount int                `json:"retryCount,omitempty"` // Track retry attempts
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ConfigMapSync is the Schema for the configmapsyncs API
type ConfigMapSync struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of ConfigMapSync
	// +required
	Spec ConfigMapSyncSpec `json:"spec"`

	// status defines the observed state of ConfigMapSync
	// +optional
	Status ConfigMapSyncStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// ConfigMapSyncList contains a list of ConfigMapSync
type ConfigMapSyncList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConfigMapSync `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ConfigMapSync{}, &ConfigMapSyncList{})
}
