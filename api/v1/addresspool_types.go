/*

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

// AddressPoolSpec defines the desired state of AddressPool
type AddressPoolSpec struct {
	Protocol  string   `json:"protocol"`
	Addresses []string `json:"addresses"`
}

// AddressPoolStatus defines the observed state of AddressPool
type AddressPoolStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// AddressPool is the Schema for the addresspools API
type AddressPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AddressPoolSpec   `json:"spec,omitempty"`
	Status AddressPoolStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AddressPoolList contains a list of AddressPool
type AddressPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AddressPool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AddressPool{}, &AddressPoolList{})
}
