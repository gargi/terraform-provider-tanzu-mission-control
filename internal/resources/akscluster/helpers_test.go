/*
Copyright 2023 VMware, Inc. All Rights Reserved.
SPDX-License-Identifier: MPL-2.0
*/

package akscluster_test

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	aksclusterclient "github.com/vmware/terraform-provider-tanzu-mission-control/internal/client/akscluster"
	aksnodepool "github.com/vmware/terraform-provider-tanzu-mission-control/internal/client/akscluster/nodepool"
	models "github.com/vmware/terraform-provider-tanzu-mission-control/internal/models/akscluster"
	objectmetamodel "github.com/vmware/terraform-provider-tanzu-mission-control/internal/models/objectmeta"
	"github.com/vmware/terraform-provider-tanzu-mission-control/internal/resources/akscluster"
)

func dataDiffFrom(t *testing.T, original map[string]any, updated map[string]any) *schema.ResourceData {
	originalData := schema.TestResourceDataRaw(t, akscluster.ClusterSchema, original)
	originalData.SetId("test-uid")
	state := originalData.State()

	sm := schema.InternalMap(akscluster.ClusterSchema)
	diff, _ := sm.Diff(context.Background(), state, terraform.NewResourceConfigRaw(updated), nil, nil, false)
	data, _ := sm.Data(state, diff)

	return data
}

func expectedFullName() *models.VmwareTanzuManageV1alpha1AksclusterFullName {
	return &models.VmwareTanzuManageV1alpha1AksclusterFullName{
		CredentialName:    "test-cred",
		SubscriptionID:    "sub-id",
		ResourceGroupName: "resource-group",
		Name:              "test-cluster",
	}
}

type clusterWither func(c *models.VmwareTanzuManageV1alpha1AksclusterAksCluster)

func withStatusSuccess(c *models.VmwareTanzuManageV1alpha1AksclusterAksCluster) {
	c.Status = &models.VmwareTanzuManageV1alpha1AksclusterStatus{
		Phase: models.VmwareTanzuManageV1alpha1AksclusterPhaseREADY.Pointer(),
	}
}

func withNodepoolStatusSuccess(c *models.VmwareTanzuManageV1alpha1AksclusterNodepoolNodepool) {
	c.Status = &models.VmwareTanzuManageV1alpha1AksclusterNodepoolStatus{
		Phase: models.VmwareTanzuManageV1alpha1AksclusterNodepoolPhaseREADY.Pointer(),
	}
}

func withStatusError(c *models.VmwareTanzuManageV1alpha1AksclusterAksCluster) {
	c.Status = &models.VmwareTanzuManageV1alpha1AksclusterStatus{
		Phase: models.VmwareTanzuManageV1alpha1AksclusterPhaseERROR.Pointer(),
	}
}

func aTestCluster(w ...clusterWither) *models.VmwareTanzuManageV1alpha1AksclusterAksCluster {
	c := &models.VmwareTanzuManageV1alpha1AksclusterAksCluster{
		FullName: &models.VmwareTanzuManageV1alpha1AksclusterFullName{
			CredentialName:    "test-cred",
			ResourceGroupName: "resource-group",
			SubscriptionID:    "sub-id",
			Name:              "test-cluster",
		},
		Meta: &objectmetamodel.VmwareTanzuCoreV1alpha1ObjectMeta{
			UID: "test-uid",
		},
		Spec: &models.VmwareTanzuManageV1alpha1AksclusterSpec{
			ClusterGroupName: "my-cluster-group",
			Config: &models.VmwareTanzuManageV1alpha1AksclusterClusterConfig{
				Location:              "eastus",
				Version:               "1.26.0",
				NodeResourceGroupName: "my-node-group",
				DiskEncryptionSetID:   "disk-encryption-set-id",
				Tags:                  map[string]string{"custom-tag": "tag-data"},
				Sku: &models.VmwareTanzuManageV1alpha1AksclusterClusterSKU{
					Name: models.VmwareTanzuManageV1alpha1AksclusterClusterSKUNameBASIC.Pointer(),
					Tier: models.VmwareTanzuManageV1alpha1AksclusterTierFREE.Pointer(),
				},
				AccessConfig: &models.VmwareTanzuManageV1alpha1AksclusterAccessConfig{
					AadConfig: &models.VmwareTanzuManageV1alpha1AksclusterAADConfig{
						AdminGroupObjectIds: []string{"admin-group1", "admin-group-2"},
						EnableAzureRbac:     true,
						Managed:             true,
						TenantID:            "my-tenant-id",
					},
					DisableLocalAccounts: true,
					EnableRbac:           true,
				},
				APIServerAccessConfig: &models.VmwareTanzuManageV1alpha1AksclusterAPIServerAccessConfig{
					AuthorizedIPRanges:   []string{"127.0.0.1", "127.0.0.2"},
					EnablePrivateCluster: true,
				},
				LinuxConfig: &models.VmwareTanzuManageV1alpha1AksclusterLinuxConfig{
					AdminUsername: "my-admin",
					SSHKeys:       []string{"key1", "key2"},
				},
				NetworkConfig: &models.VmwareTanzuManageV1alpha1AksclusterNetworkConfig{
					DNSPrefix:        "net-prefix",
					DNSServiceIP:     "127.0.0.1",
					DockerBridgeCidr: "127.0.0.2",
					LoadBalancerSku:  "load-balancer",
					NetworkPlugin:    "azure",
					NetworkPolicy:    "policy",
					PodCidrs:         []string{"127.0.0.3"},
					ServiceCidrs:     []string{"127.0.0.4"},
				},
				StorageConfig: &models.VmwareTanzuManageV1alpha1AksclusterStorageConfig{
					EnableDiskCsiDriver:      true,
					EnableFileCsiDriver:      true,
					EnableSnapshotController: true,
				},
				AddonsConfig: &models.VmwareTanzuManageV1alpha1AksclusterAddonsConfig{
					AzureKeyvaultSecretsProviderConfig: &models.VmwareTanzuManageV1alpha1AksclusterAzureKeyvaultSecretsProviderAddonConfig{
						Enabled:              true,
						EnableSecretRotation: true,
						RotationPoolInterval: "5m",
					},
					MonitoringConfig: &models.VmwareTanzuManageV1alpha1AksclusterMonitoringAddonConfig{
						Enabled:                 true,
						LogAnalyticsWorkspaceID: "workspace-id",
					},
					AzurePolicyConfig: &models.VmwareTanzuManageV1alpha1AksclusterAzurePolicyAddonConfig{
						Enabled: true,
					},
				},
				AutoUpgradeConfig: &models.VmwareTanzuManageV1alpha1AksclusterAutoUpgradeConfig{
					Channel: models.VmwareTanzuManageV1alpha1AksclusterChannelSTABLE.Pointer(),
				},
			},
			ProxyName:  "my-proxy",
			AgentName:  "my-agent-name",
			ResourceID: "my-resource-id",
		},
	}

	for _, f := range w {
		f(c)
	}

	return c
}

type mapWither func(map[string]any)

func withoutNodepools(m map[string]any) {
	spec := m["spec"].([]any)
	spec[0].(map[string]any)["nodepool"] = []any{}
}

func withoutNodepoolType(m map[string]any) {
	specs := m["spec"].([]any)
	spec := specs[0].(map[string]any)
	nps := spec["nodepool"].([]any)
	np := nps[0].(map[string]any)
	npSpecs := np["spec"].([]any)
	npSpec := npSpecs[0].(map[string]any)
	delete(npSpec, "type")
}

func with5msTimeout(m map[string]any) {
	m["ready_wait_timeout"] = (5 * time.Millisecond).String()
}

func withDNSPrefix(prefix string) mapWither {
	return func(m map[string]any) {
		specs := m["spec"].([]any)
		spec := specs[0].(map[string]any)
		configs := spec["config"].([]any)
		config := configs[0].(map[string]any)
		networks := config["network_config"].([]any)
		network := networks[0].(map[string]any)
		network["dns_prefix"] = prefix
	}
}

func withNodepools(nps []any) mapWither {
	return func(m map[string]any) {
		specs := m["spec"].([]any)
		spec := specs[0].(map[string]any)
		spec["nodepool"] = nps
	}
}

func withName(name string) mapWither {
	return func(m map[string]any) {
		m["name"] = name
	}
}

func withNodepoolCount(count int) mapWither {
	return func(m map[string]any) {
		specs := m["spec"].([]any)
		spec := specs[0].(map[string]any)
		spec["count"] = count
	}
}

func withNodepoolMode(mode string) mapWither {
	return func(m map[string]any) {
		specs := m["spec"].([]any)
		spec := specs[0].(map[string]any)
		spec["mode"] = mode
	}
}

func aTestClusterDataMap(w ...mapWither) map[string]any {
	m := map[string]any{
		"credential_name": "test-cred",
		"subscription_id": "sub-id",
		"resource_group":  "resource-group",
		"name":            "test-cluster",
		"spec": []any{map[string]any{
			"cluster_group": "my-cluster-group",
			"proxy":         "my-proxy",
			"config": []any{map[string]any{
				"location":                 "eastus",
				"kubernetes_version":       "1.26.0",
				"node_resource_group_name": "my-node-group",
				"disk_encryption_set":      "disk-encryption-set-id",
				"tags": map[string]any{
					"custom-tag": "tag-data",
				},
				"sku": []any{map[string]any{
					"name": "BASIC",
					"tier": "FREE",
				}},
				"access_config": []any{map[string]any{
					"enable_rbac":            true,
					"disable_local_accounts": true,
					"aad_config": []any{map[string]any{
						"managed":           true,
						"tenant_id":         "my-tenant-id",
						"admin_group_ids":   []any{"admin-group1", "admin-group-2"},
						"enable_azure_rbac": true,
					}},
				}},
				"api_server_access_config": []any{map[string]any{
					"authorized_ip_ranges":   []any{"127.0.0.1", "127.0.0.2"},
					"enable_private_cluster": true,
				}},
				"linux_config": []any{map[string]any{
					"admin_username": "my-admin",
					"ssh_keys":       []any{"key1", "key2"},
				}},
				"network_config": []any{map[string]any{
					"load_balancer_sku":  "load-balancer",
					"network_plugin":     "azure",
					"network_policy":     "policy",
					"dns_service_ip":     "127.0.0.1",
					"docker_bridge_cidr": "127.0.0.2",
					"pod_cidr":           []any{"127.0.0.3"},
					"service_cidr":       []any{"127.0.0.4"},
					"dns_prefix":         "net-prefix",
				}},
				"storage_config": []any{map[string]any{
					"enable_disk_csi_driver":     true,
					"enable_file_csi_driver":     true,
					"enable_snapshot_controller": true,
				}},
				"addon_config": []any{map[string]any{
					"azure_keyvault_secrets_provider_addon_config": []any{map[string]any{
						"enable":                 true,
						"enable_secret_rotation": true,
						"rotation_poll_interval": "5m",
					}},
					"monitor_addon_config": []any{map[string]any{
						"enable":                     true,
						"log_analytics_workspace_id": "workspace-id",
					}},
					"azure_policy_addon_config": []any{map[string]any{
						"enable": true,
					}},
				}},
				"auto_upgrade_config": []any{map[string]any{
					"upgrade_channel": "STABLE",
				}},
			}},
			"nodepool": []any{
				aTestNodepoolDataMap(),
			},
			"agent_name":  "my-agent-name",
			"resource_id": "my-resource-id",
		}},
	}

	for _, f := range w {
		f(m)
	}

	return m
}

type nodepoolWither func(np *models.VmwareTanzuManageV1alpha1AksclusterNodepoolNodepool)

func withNodepoolName(name string) nodepoolWither {
	return func(np *models.VmwareTanzuManageV1alpha1AksclusterNodepoolNodepool) {
		np.FullName.Name = name
	}
}

func forCluster(c *models.VmwareTanzuManageV1alpha1AksclusterAksCluster) nodepoolWither {
	return func(np *models.VmwareTanzuManageV1alpha1AksclusterNodepoolNodepool) {
		np.FullName.CredentialName = c.FullName.CredentialName
		np.FullName.SubscriptionID = c.FullName.SubscriptionID
		np.FullName.ResourceGroupName = c.FullName.ResourceGroupName
		np.FullName.AksClusterName = c.FullName.Name
	}
}

func withNodepoolStatusError() nodepoolWither {
	return func(np *models.VmwareTanzuManageV1alpha1AksclusterNodepoolNodepool) {
		np.Status = &models.VmwareTanzuManageV1alpha1AksclusterNodepoolStatus{
			Phase: models.VmwareTanzuManageV1alpha1AksclusterNodepoolPhaseERROR.Pointer(),
		}
	}
}

func aTestNodePool(w ...nodepoolWither) *models.VmwareTanzuManageV1alpha1AksclusterNodepoolNodepool {
	np := &models.VmwareTanzuManageV1alpha1AksclusterNodepoolNodepool{
		FullName: &models.VmwareTanzuManageV1alpha1AksclusterNodepoolFullName{
			Name: "system-np",
		},
		Spec: &models.VmwareTanzuManageV1alpha1AksclusterNodepoolSpec{
			Mode:              models.VmwareTanzuManageV1alpha1AksclusterNodepoolModeSYSTEM.Pointer(),
			Type:              models.VmwareTanzuManageV1alpha1AksclusterNodepoolTypeVIRTUALMACHINESCALESETS.Pointer(),
			AvailabilityZones: []string{"1", "2", "3"},
			Count:             1,
			VMSize:            "STANDARD_DS2v2",
			AutoScaling: &models.VmwareTanzuManageV1alpha1AksclusterNodepoolAutoScalingConfig{
				Enabled:                true,
				MinCount:               1,
				MaxCount:               10,
				ScaleSetEvictionPolicy: models.VmwareTanzuManageV1alpha1AksclusterNodepoolScaleSetEvictionPolicyDELETE.Pointer(),
				ScaleSetPriority:       models.VmwareTanzuManageV1alpha1AksclusterNodepoolScaleSetPriorityREGULAR.Pointer(),
				SpotMaxPrice:           1.5,
			},
			EnableNodePublicIP: true,
			MaxPods:            110,
			NodeLabels:         map[string]string{"label": "val"},
			NodeTaints: []*models.VmwareTanzuManageV1alpha1AksclusterNodepoolTaint{{
				Effect: models.VmwareTanzuManageV1alpha1AksclusterNodepoolTaintEffectNOSCHEDULE.Pointer(),
				Key:    "tkey",
				Value:  "tval",
			}},
			OsDiskSizeGb: 30,
			OsDiskType:   models.VmwareTanzuManageV1alpha1AksclusterNodepoolOsDiskTypeEPHEMERAL.Pointer(),
			OsType:       models.VmwareTanzuManageV1alpha1AksclusterNodepoolOsTypeLINUX.Pointer(),
			Tags:         map[string]string{"tmc.node.tag": "val"},
			UpgradeConfig: &models.VmwareTanzuManageV1alpha1AksclusterNodepoolUpgradeConfig{
				MaxSurge: "50%",
			},
			VnetSubnetID: "subnet-1",
		},
	}

	for _, f := range w {
		f(np)
	}

	return np
}

func aTestNodepoolDataMap(w ...mapWither) map[string]any {
	m := map[string]any{
		"name": "system-np",
		"spec": []any{map[string]any{
			"mode":                  "SYSTEM",
			"type":                  "VIRTUAL_MACHINE_SCALE_SETS",
			"availability_zones":    []any{"1", "2", "3"},
			"count":                 1,
			"vm_size":               "STANDARD_DS2v2",
			"os_type":               "LINUX",
			"os_disk_type":          "EPHEMERAL",
			"os_disk_size_gb":       30,
			"max_pods":              110,
			"enable_node_public_ip": true,
			"taints": []any{
				map[string]any{
					"effect": "NO_SCHEDULE",
					"key":    "tkey",
					"value":  "tval",
				},
			},
			"vnet_subnet_id": "subnet-1",
			"node_labels":    map[string]any{"label": "val"},
			"tags":           map[string]any{"tmc.node.tag": "val"},
			"auto_scaling_config": []any{map[string]any{
				"enable":                    true,
				"min_count":                 1,
				"max_count":                 10,
				"scale_set_priority":        "REGULAR",
				"scale_set_eviction_policy": "DELETE",
				"spot_max_price":            1.5,
			}},
			"upgrade_config": []any{map[string]any{
				"max_surge": "50%",
			}},
		}},
	}

	for _, f := range w {
		f(m)
	}

	return m
}

var _ aksclusterclient.ClientService = &mockClusterClient{}

type mockClusterClient struct {
	AksClusterResourceServiceGetCalledWith    *models.VmwareTanzuManageV1alpha1AksclusterFullName
	getClusterResp                            *models.VmwareTanzuManageV1alpha1AksclusterAksCluster
	createClusterResp                         *models.VmwareTanzuManageV1alpha1AksclusterAksCluster
	AksClusterResourceServiceDeleteCalledWith *models.VmwareTanzuManageV1alpha1AksclusterFullName
	AksUpdateClusterWasCalledWith             *models.VmwareTanzuManageV1alpha1AksclusterAksCluster
	AksClusterResourceServiceGetCallCount     int
	AksCreateClusterWasCalled                 bool
	createErr                                 error
	getErr                                    error
	updateErr                                 error
	deleteErr                                 error
}

func (m *mockClusterClient) AksClusterResourceServiceCreate(_ *models.VmwareTanzuManageV1alpha1AksclusterCreateAksClusterRequest) (*models.VmwareTanzuManageV1alpha1AksclusterCreateAksClusterResponse, error) {
	m.AksCreateClusterWasCalled = true

	return &models.VmwareTanzuManageV1alpha1AksclusterCreateAksClusterResponse{
		AksCluster: m.createClusterResp,
	}, m.createErr
}

func (m *mockClusterClient) AksClusterResourceServiceGet(fn *models.VmwareTanzuManageV1alpha1AksclusterFullName) (*models.VmwareTanzuManageV1alpha1AksclusterGetAksClusterResponse, error) {
	m.AksClusterResourceServiceGetCalledWith = fn
	m.AksClusterResourceServiceGetCallCount += 1

	return &models.VmwareTanzuManageV1alpha1AksclusterGetAksClusterResponse{
		AksCluster: m.getClusterResp,
	}, m.getErr
}

func (m *mockClusterClient) AksClusterResourceServiceUpdate(ucr *models.VmwareTanzuManageV1alpha1AksclusterUpdateAksClusterRequest) (*models.VmwareTanzuManageV1alpha1AksclusterUpdateAksClusterResponse, error) {
	m.AksUpdateClusterWasCalledWith = ucr.AksCluster

	return nil, m.updateErr
}

func (m *mockClusterClient) AksClusterResourceServiceDelete(fn *models.VmwareTanzuManageV1alpha1AksclusterFullName, _ string) error {
	m.AksClusterResourceServiceDeleteCalledWith = fn

	return m.deleteErr
}

var _ aksnodepool.ClientService = &mockNodepoolClient{}

type mockNodepoolClient struct {
	AksNodePoolResourceServiceListCalledWith *models.VmwareTanzuManageV1alpha1AksclusterFullName
	CreateNodepoolWasCalledWith              *models.VmwareTanzuManageV1alpha1AksclusterNodepoolNodepool
	UpdatedNodepoolWasCalledWith             *models.VmwareTanzuManageV1alpha1AksclusterNodepoolNodepool
	DeleteNodepoolWasCalledWith              *models.VmwareTanzuManageV1alpha1AksclusterNodepoolFullName
	GetNodepoolCalledWith                    *models.VmwareTanzuManageV1alpha1AksclusterNodepoolFullName
	nodepoolListResp                         []*models.VmwareTanzuManageV1alpha1AksclusterNodepoolNodepool
	nodepoolGetResp                          *models.VmwareTanzuManageV1alpha1AksclusterNodepoolNodepool
	createErr                                error
	listErr                                  error
	updateErr                                error
	getErr                                   error
	DeleteErr                                error
}

func (m *mockNodepoolClient) AksNodePoolResourceServiceCreate(req *models.VmwareTanzuManageV1alpha1AksclusterNodepoolCreateNodepoolRequest) (*models.VmwareTanzuManageV1alpha1AksclusterNodepoolCreateNodepoolResponse, error) {
	m.CreateNodepoolWasCalledWith = req.Nodepool

	return nil, m.createErr
}

func (m *mockNodepoolClient) AksNodePoolResourceServiceList(fn *models.VmwareTanzuManageV1alpha1AksclusterFullName) (*models.VmwareTanzuManageV1alpha1AksclusterNodepoolListNodepoolsResponse, error) {
	m.AksNodePoolResourceServiceListCalledWith = fn

	return &models.VmwareTanzuManageV1alpha1AksclusterNodepoolListNodepoolsResponse{
		Nodepools:  m.nodepoolListResp,
		TotalCount: "1",
	}, m.listErr
}

func (m *mockNodepoolClient) AksNodePoolResourceServiceGet(fn *models.VmwareTanzuManageV1alpha1AksclusterNodepoolFullName) (*models.VmwareTanzuManageV1alpha1AksclusterNodepoolGetNodepoolResponse, error) {
	m.GetNodepoolCalledWith = fn

	return &models.VmwareTanzuManageV1alpha1AksclusterNodepoolGetNodepoolResponse{Nodepool: m.nodepoolGetResp}, m.getErr
}

func (m *mockNodepoolClient) AksNodePoolResourceServiceUpdate(req *models.VmwareTanzuManageV1alpha1AksclusterNodepoolUpdateNodepoolRequest) (*models.VmwareTanzuManageV1alpha1AksclusterNodepoolCreateNodepoolResponse, error) {
	m.UpdatedNodepoolWasCalledWith = req.Nodepool

	return nil, m.updateErr
}

func (m *mockNodepoolClient) AksNodePoolResourceServiceDelete(req *models.VmwareTanzuManageV1alpha1AksclusterNodepoolFullName) error {
	m.DeleteNodepoolWasCalledWith = req

	return m.DeleteErr
}
