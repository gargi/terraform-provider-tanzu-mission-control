/*
Copyright 2022 VMware, Inc. All Rights Reserved.
SPDX-License-Identifier: MPL-2.0
*/

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/swag"
)

// VmwareTanzuManageV1alpha1EksclusterLogging EKS logging configuration.
// Refer https://docs.aws.amazon.com/eks/latest/userguide/control-plane-logs.html for more info
//
// swagger:model vmware.tanzu.manage.v1alpha1.ekscluster.Logging
type VmwareTanzuManageV1alpha1EksclusterLogging struct {

	// Enable API server logs.
	APIServer bool `json:"apiServer,omitempty"`

	// Enable audit logs.
	Audit bool `json:"audit,omitempty"`

	// Enable authenticator logs.
	Authenticator bool `json:"authenticator,omitempty"`

	// Enable controller manager logs.
	ControllerManager bool `json:"controllerManager,omitempty"`

	// Enable scheduler logs.
	Scheduler bool `json:"scheduler,omitempty"`
}

// MarshalBinary interface implementation.
func (m *VmwareTanzuManageV1alpha1EksclusterLogging) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}

	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation.
func (m *VmwareTanzuManageV1alpha1EksclusterLogging) UnmarshalBinary(b []byte) error {
	var res VmwareTanzuManageV1alpha1EksclusterLogging
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}

	*m = res

	return nil
}
