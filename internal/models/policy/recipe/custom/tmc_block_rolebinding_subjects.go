/*
Copyright © 2022 VMware, Inc. All Rights Reserved.
SPDX-License-Identifier: MPL-2.0
Code generated by go-swagger; DO NOT EDIT.
*/

package policyrecipecustommodel

import (
	"github.com/go-openapi/swag"

	policyrecipecustomcommonmodel "github.com/vmware/terraform-provider-tanzu-mission-control/internal/models/policy/recipe/custom/common"
)

// VmwareTanzuManageV1alpha1CommonPolicySpecCustomV1TMCBlockRoleBindingSubjects tmc-block-rolebinding-subjects recipe schema.
//
// The input schema for tmc-block-rolebinding-subjects recipe.
//
// swagger:model VmwareTanzuManageV1alpha1CommonPolicySpecCustomV1TMCBlockRoleBindingSubjects
type VmwareTanzuManageV1alpha1CommonPolicySpecCustomV1TMCBlockRoleBindingSubjects struct {

	// Audit (dry-run).
	// Creates this policy for dry-run. Violations will be logged but not denied. Defaults to false (deny).
	Audit bool `json:"audit,omitempty"`

	// Parameters.
	Parameters *VmwareTanzuManageV1alpha1CommonPolicySpecCustomV1TMCBlockRoleBindingSubjectsParameters `json:"parameters,omitempty"`

	// TargetKubernetesResources is a list of kubernetes api resources on which the policy will be enforced, identified using apiGroups and kinds. You can use 'kubectl api-resources' to view the list of available api resources on your cluster.
	// Required: true
	// Min Items: 1
	TargetKubernetesResources []*policyrecipecustomcommonmodel.VmwareTanzuManageV1alpha1CommonPolicySpecCustomV1TargetKubernetesResources `json:"targetKubernetesResources"`
}

// MarshalBinary interface implementation.
func (m *VmwareTanzuManageV1alpha1CommonPolicySpecCustomV1TMCBlockRoleBindingSubjects) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}

	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation.
func (m *VmwareTanzuManageV1alpha1CommonPolicySpecCustomV1TMCBlockRoleBindingSubjects) UnmarshalBinary(b []byte) error {
	var res VmwareTanzuManageV1alpha1CommonPolicySpecCustomV1TMCBlockRoleBindingSubjects
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}

	*m = res

	return nil
}

// VmwareTanzuManageV1alpha1CommonPolicySpecCustomV1TMCBlockRoleBindingSubjectsParameters Parameters.
//
// swagger:model VmwareTanzuManageV1alpha1CommonPolicySpecCustomV1TMCBlockRoleBindingSubjectsParameters
type VmwareTanzuManageV1alpha1CommonPolicySpecCustomV1TMCBlockRoleBindingSubjectsParameters struct {

	// Disallowed Subjects
	// Min Items: 1
	DisallowedSubjects []*VmwareTanzuManageV1alpha1CommonPolicySpecCustomV1TMCBlockRoleBindingSubjectsParametersDisallowedSubjects `json:"disallowedSubjects"`
}

// MarshalBinary interface implementation.
func (m *VmwareTanzuManageV1alpha1CommonPolicySpecCustomV1TMCBlockRoleBindingSubjectsParameters) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}

	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation.
func (m *VmwareTanzuManageV1alpha1CommonPolicySpecCustomV1TMCBlockRoleBindingSubjectsParameters) UnmarshalBinary(b []byte) error {
	var res VmwareTanzuManageV1alpha1CommonPolicySpecCustomV1TMCBlockRoleBindingSubjectsParameters
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}

	*m = res

	return nil
}

// VmwareTanzuManageV1alpha1CommonPolicySpecCustomV1TMCBlockRoleBindingSubjectsParametersDisallowedSubjects Disallowed Subjects.
//
// swagger:model VmwareTanzuManageV1alpha1CommonPolicySpecCustomV1TMCBlockRoleBindingSubjectsParametersDisallowedSubjects
type VmwareTanzuManageV1alpha1CommonPolicySpecCustomV1TMCBlockRoleBindingSubjectsParametersDisallowedSubjects struct {

	// The kind of subject to disallow, can be User/Group/ServiceAccount
	// Pattern: User|Group|ServiceAccount
	Kind string `json:"kind,omitempty"`

	// The name of the subject to disallow
	Name string `json:"name,omitempty"`
}

// MarshalBinary interface implementation.
func (m *VmwareTanzuManageV1alpha1CommonPolicySpecCustomV1TMCBlockRoleBindingSubjectsParametersDisallowedSubjects) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}

	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation.
func (m *VmwareTanzuManageV1alpha1CommonPolicySpecCustomV1TMCBlockRoleBindingSubjectsParametersDisallowedSubjects) UnmarshalBinary(b []byte) error {
	var res VmwareTanzuManageV1alpha1CommonPolicySpecCustomV1TMCBlockRoleBindingSubjectsParametersDisallowedSubjects
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}

	*m = res

	return nil
}
