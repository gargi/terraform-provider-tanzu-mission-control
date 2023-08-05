/*
Copyright © 2023 VMware, Inc. All Rights Reserved.
SPDX-License-Identifier: MPL-2.0
Code generated by go-swagger; DO NOT EDIT.
*/

package policyrecipequotamodel

import "github.com/go-openapi/swag"

// VmwareTanzuManageV1alpha1CommonPolicySpecQuotaV1Custom Input schema for namespace quota policy custom recipe version v1
//
// # The input schema for namespace quota policy custom recipe version v1
//
// swagger:model VmwareTanzuManageV1alpha1CommonPolicySpecQuotaV1Custom
type VmwareTanzuManageV1alpha1CommonPolicySpecQuotaV1Custom struct {

	// The sum of CPU limits across all pods in a non-terminal state cannot exceed this value.
	LimitsCPU string `json:"limitsCpu,omitempty"`

	// The sum of memory limits across all pods in a non-terminal state cannot exceed this value.
	// Pattern: ^[0-9]+(E|P|T|G|M|K|Ei|Pi|Ti|Gi|Mi|Ki)?$
	LimitsMemory string `json:"limitsMemory,omitempty"`

	// The total number of PersistentVolumeClaims that can exist in a namespace.
	Persistentvolumeclaims int64 `json:"persistentvolumeclaims,omitempty"`

	// Across all persistent volume claims associated with each storage class, the total number of persistent volume claims that can exist in the namespace.
	PersistentvolumeclaimsPerClass map[string]int `json:"persistentvolumeclaimsPerClass,omitempty"`

	// The sum of CPU requests across all pods in a non-terminal state cannot exceed this value.
	RequestsCPU string `json:"requestsCpu,omitempty"`

	// The sum of memory requests across all pods in a non-terminal state cannot exceed this value.
	// Pattern: ^[0-9]+(E|P|T|G|M|K|Ei|Pi|Ti|Gi|Mi|Ki)?$
	RequestsMemory string `json:"requestsMemory,omitempty"`

	// The sum of storage requests across all persistent volume claims cannot exceed this value.
	// Pattern: ^[0-9]+(E|P|T|G|M|K|Ei|Pi|Ti|Gi|Mi|Ki)?$
	RequestsStorage string `json:"requestsStorage,omitempty"`

	// Across all persistent volume claims associated with each storage class, the sum of storage requests cannot exceed this value.
	RequestsStoragePerClass map[string]string `json:"requestsStoragePerClass,omitempty"`

	// The total number of Services of the given type that can exist in a namespace.
	ResourceCounts map[string]int `json:"resourceCounts,omitempty"`
}

// MarshalBinary interface implementation.
func (m *VmwareTanzuManageV1alpha1CommonPolicySpecQuotaV1Custom) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}

	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation.
func (m *VmwareTanzuManageV1alpha1CommonPolicySpecQuotaV1Custom) UnmarshalBinary(b []byte) error {
	var res VmwareTanzuManageV1alpha1CommonPolicySpecQuotaV1Custom
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}

	*m = res

	return nil
}
