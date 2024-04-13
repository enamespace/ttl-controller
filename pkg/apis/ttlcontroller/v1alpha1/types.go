package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TTL is a specification for a TTL resource
type TTL struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TTLSpec   `json:"spec"`
	Status TTLStatus `json:"status"`
}

// TTLTarget is the target reference for a TTL resource
type TTLTargetRef struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
}

// TTLSpec is the spec for a TTL resource
type TTLSpec struct {
	TTLTargetRef TTLTargetRef `json:"ttlTargetRef"`
	After        string       `json:"after"`
}

// TTLStatus is the status for a TTL resource
type TTLStatus struct {
	Remaining string `json:"remaining"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TTLList is a list of TTL resources
type TTLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []TTL `json:"items"`
}
