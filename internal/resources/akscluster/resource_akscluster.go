/*
Copyright 2023 VMware, Inc. All Rights Reserved.
SPDX-License-Identifier: MPL-2.0
*/

package akscluster

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/pkg/errors"

	"github.com/vmware/terraform-provider-tanzu-mission-control/internal/authctx"
	"github.com/vmware/terraform-provider-tanzu-mission-control/internal/client"
	"github.com/vmware/terraform-provider-tanzu-mission-control/internal/client/akscluster"
	clienterrors "github.com/vmware/terraform-provider-tanzu-mission-control/internal/client/errors"
	models "github.com/vmware/terraform-provider-tanzu-mission-control/internal/models/akscluster"
)

func ResourceTMCAKSCluster() *schema.Resource {
	return &schema.Resource{
		Schema:        ClusterSchema,
		CreateContext: resourceClusterCreate,
		ReadContext:   resourceClusterRead,
		UpdateContext: resourceClusterInPlaceUpdate,
		DeleteContext: resourceClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceClusterImporter,
		},
		Description: "Tanzu Mission Control AKS Cluster Resource",
	}
}

// resourceClusterCreate will create an AKS cluster and any assigned nodepool.
func resourceClusterCreate(ctx context.Context, data *schema.ResourceData, config any) diag.Diagnostics {
	tc, ok := config.(authctx.TanzuContext)
	if !ok {
		return diag.Errorf("error while retrieving Tanzu auth config")
	}

	nodepools := ConstructNodepools(data)
	if err := validate(nodepools); err != nil {
		return diag.FromErr(err)
	}

	if err := createOrUpdateCluster(data, tc.TMCConnection.AKSClusterResourceService); err != nil {
		return diag.FromErr(err)
	}

	if err := createNodepools(ctx, nodepools, tc.TMCConnection.AKSNodePoolResourceService); err != nil {
		return diag.FromErr(err)
	}

	ctx, cancel := context.WithTimeout(ctx, getTimeOut(data))
	defer cancel()

	if err := pollUntilReady(ctx, data, tc.TMCConnection, getPollInterval(ctx)); err != nil {
		return diag.FromErr(err)
	}

	return dataSourceTMCAKSClusterRead(ctx, data, tc)
}

// resourceClusterRead read state of existing AKS cluster and assigned nodepools.
func resourceClusterRead(ctx context.Context, data *schema.ResourceData, config any) diag.Diagnostics {
	return dataSourceTMCAKSClusterRead(ctx, data, config)
}

// resourceClusterInPlaceUpdate updates AKS cluster settings in place then manually reconciles nodepool changes
// updating, creating, and deleting where appropriate.
func resourceClusterInPlaceUpdate(ctx context.Context, data *schema.ResourceData, config any) diag.Diagnostics {
	tc, ok := config.(authctx.TanzuContext)
	if !ok {
		return diag.Errorf("error while retrieving Tanzu auth config")
	}

	clusterResp, nodepoolResp, getErr := getClusterAndNodepools(ctx, data, tc.TMCConnection)
	if getErr != nil || clusterResp == nil || nodepoolResp == nil {
		return diag.FromErr(errors.Errorf("Unable to get Tanzu Mission Control AKS cluster entry, name : %s", data.Get(NameKey)))
	}

	// Make changes to the cluster config.
	if clusterChange := data.HasChange("spec.0.config.0"); clusterChange {
		if updateErr := updateClusterConfig(ctx, data, clusterResp, tc); updateErr != nil {
			return diag.FromErr(updateErr)
		}
	}

	// Make changes to cluster nodepools.
	if data.HasChange("spec.0.nodepool") {
		if npChangeErr := handleNodepoolChanges(ctx, nodepoolResp.Nodepools, data, tc.TMCConnection); npChangeErr != nil {
			return diag.FromErr(npChangeErr)
		}
	}

	return dataSourceTMCAKSClusterRead(ctx, data, config)
}

// resourceClusterDelete deletes an AKS cluster and all associated node pools.
func resourceClusterDelete(ctx context.Context, data *schema.ResourceData, config any) diag.Diagnostics {
	tc, ok := config.(authctx.TanzuContext)
	if !ok {
		return diag.Errorf("error while retrieving Tanzu auth config")
	}

	fn := extractClusterFullName(data)
	if err := tc.TMCConnection.AKSClusterResourceService.AksClusterResourceServiceDelete(fn, "false"); err != nil && !clienterrors.IsNotFoundError(err) {
		return diag.FromErr(errors.Wrapf(err, "Unable to delete Tanzu Mission Control AKS cluster entry, name : %s", data.Get(NameKey)))
	}

	ctx, cancel := context.WithTimeout(ctx, getTimeOut(data))
	defer cancel()

	if err := pollUntilClusterDeleted(ctx, data, tc.TMCConnection.AKSClusterResourceService, getPollInterval(ctx)); err != nil {
		return diag.FromErr(err)
	}

	data.SetId("") // explicitly delete

	return diag.Diagnostics{}
}

func resourceClusterImporter(_ context.Context, data *schema.ResourceData, config any) ([]*schema.ResourceData, error) {
	tc, ok := config.(authctx.TanzuContext)
	if !ok {
		return nil, errors.New("error while retrieving Tanzu auth config")
	}

	id := data.Id()
	if id == "" {
		return nil, errors.New("ID is needed to import an TMC AKS cluster")
	}

	resp, err := tc.TMCConnection.AKSClusterResourceService.AksClusterResourceServiceGetByID(id)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to get Tanzu Mission Control AKS cluster entry for id %s", id)
	}

	npresp, err := tc.TMCConnection.AKSNodePoolResourceService.AksNodePoolResourceServiceList(resp.AksCluster.FullName)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to get Tanzu Mission Control AKS nodepools for cluster %s", resp.AksCluster.FullName.Name)
	}

	if err = data.Set(CredentialNameKey, resp.AksCluster.FullName.CredentialName); err != nil {
		return nil, errors.Wrapf(err, "Failed to set credential name for the cluster %s", resp.AksCluster.FullName.Name)
	}

	if err = data.Set(SubscriptionIDKey, resp.AksCluster.FullName.SubscriptionID); err != nil {
		return nil, errors.Wrapf(err, "Failed to set subscription for the cluster %s", resp.AksCluster.FullName.Name)
	}

	if err = data.Set(ResourceGroupNameKey, resp.AksCluster.FullName.ResourceGroupName); err != nil {
		return nil, errors.Wrapf(err, "Failed to set resource group for the cluster %s", resp.AksCluster.FullName.Name)
	}

	if err = data.Set(NameKey, resp.AksCluster.FullName.Name); err != nil {
		return nil, errors.Wrapf(err, "Failed to set name for the cluster %s", resp.AksCluster.FullName.Name)
	}

	if err = setResourceState(data, resp.AksCluster, npresp.Nodepools); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{data}, nil
}

// validate returns an error configuration will result in a cluster that will fail to create.
func validate(nodepools []*models.VmwareTanzuManageV1alpha1AksclusterNodepoolNodepool) error {
	for _, n := range nodepools {
		if *n.Spec.Mode == models.VmwareTanzuManageV1alpha1AksclusterNodepoolModeSYSTEM {
			return nil
		}
	}

	return errors.New("AKS cluster must contain at least 1 SYSTEM nodepool")
}

// createOrUpdateCluster creates an AKS cluster in TMC.  It is possible the cluster already exists in which case the
// existing cluster is updated with any node pools defined in the configuration.
func createOrUpdateCluster(data *schema.ResourceData, client akscluster.ClientService) error {
	cluster := ConstructCluster(data)
	clusterReq := &models.VmwareTanzuManageV1alpha1AksclusterCreateAksClusterRequest{AksCluster: cluster}
	createResp, err := client.AksClusterResourceServiceCreate(clusterReq)

	if clienterrors.IsAlreadyExistsError(err) {
		if getErr := getExistingCluster(data, client, clusterReq); getErr != nil {
			return errors.Wrapf(getErr, "Failed to created cluster do to conflict but conflicting cluster not found")
		}

		return nil
	}

	if err != nil {
		return err
	}

	data.SetId(createResp.AksCluster.Meta.UID)

	return nil
}

func getExistingCluster(data *schema.ResourceData, client akscluster.ClientService, clusterReq *models.VmwareTanzuManageV1alpha1AksclusterCreateAksClusterRequest) error {
	getResp, getErr := client.AksClusterResourceServiceGet(clusterReq.AksCluster.FullName)
	if getErr != nil {
		return getErr
	}

	data.SetId(getResp.AksCluster.Meta.UID)

	return nil
}

func updateClusterConfig(ctx context.Context, data *schema.ResourceData, clusterResp *models.VmwareTanzuManageV1alpha1AksclusterGetAksClusterResponse, tc authctx.TanzuContext) error {
	cluster := ConstructCluster(data)
	cluster.Meta = clusterResp.AksCluster.Meta
	updateReq := &models.VmwareTanzuManageV1alpha1AksclusterUpdateAksClusterRequest{AksCluster: cluster}

	if _, updateErr := tc.TMCConnection.AKSClusterResourceService.AksClusterResourceServiceUpdate(updateReq); updateErr != nil {
		return errors.Wrapf(updateErr, "Unable to update Tanzu Mission Control AKS cluster entry, name : %s", data.Get(NameKey))
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, getTimeOut(data))
	defer cancel()

	return pollUntilReady(ctxTimeout, data, tc.TMCConnection, getPollInterval(ctx))
}

func pollUntilReady(ctx context.Context, data *schema.ResourceData, mc *client.TanzuMissionControl, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	fn := extractClusterFullName(data)

	for {
		select {
		case <-ctx.Done():
			return errors.New("Timed out waiting for READY")
		case <-ticker.C:
			aksClusterResp, err := mc.AKSClusterResourceService.AksClusterResourceServiceGet(fn)
			if clienterrors.IsNotFoundError(err) {
				_ = schema.RemoveFromState(data, nil)
				return errors.Errorf("Unable to get Tanzu Mission Control AKS cluster entry, name : %s", data.Get(NameKey))
			}

			if clusterHasFatalError(aksClusterResp) {
				return errors.Errorf("Cluster creation failed: %s", getErrorReason(aksClusterResp.AksCluster.Status.Conditions))
			}

			if clusterIsReady(aksClusterResp) {
				_, npErr := mc.AKSNodePoolResourceService.AksNodePoolResourceServiceList(fn)
				if clienterrors.IsNotFoundError(npErr) {
					return errors.Errorf("Unable to get Tanzu Mission Control AKS nodepools for entry, name : %s", data.Get(NameKey))
				}

				if npErr == nil {
					return nil
				}
			}
		}
	}
}

func pollUntilClusterDeleted(ctx context.Context, data *schema.ResourceData, client akscluster.ClientService, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	fn := extractClusterFullName(data)

	for {
		select {
		case <-ctx.Done():
			return errors.New("timed out waiting for delete")
		case <-ticker.C:
			_, err := client.AksClusterResourceServiceGet(fn)
			if clienterrors.IsNotFoundError(err) {
				return nil
			}
		}
	}
}
