/*
Copyright 2023 VMware, Inc. All Rights Reserved.
SPDX-License-Identifier: MPL-2.0
*/

package akscluster_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/suite"

	"github.com/vmware/terraform-provider-tanzu-mission-control/internal/authctx"
	"github.com/vmware/terraform-provider-tanzu-mission-control/internal/client"
	clienterrors "github.com/vmware/terraform-provider-tanzu-mission-control/internal/client/errors"
	models "github.com/vmware/terraform-provider-tanzu-mission-control/internal/models/akscluster"
	"github.com/vmware/terraform-provider-tanzu-mission-control/internal/resources/akscluster"
	"github.com/vmware/terraform-provider-tanzu-mission-control/internal/resources/common"
)

type mocks struct {
	clusterClient  *mockClusterClient
	nodepoolClient *mockNodepoolClient
}

func TestAKSClusterResource(t *testing.T) {
	suite.Run(t, &CreatClusterTestSuite{})
	suite.Run(t, &ReadClusterTestSuite{})
	suite.Run(t, &UpdateClusterTestSuite{})
	suite.Run(t, &DeleteClusterTestSuite{})
	suite.Run(t, &ImportClusterTestSuite{})
}

type CreatClusterTestSuite struct {
	suite.Suite
	ctx                context.Context
	mocks              mocks
	aksClusterResource *schema.Resource
	config             authctx.TanzuContext
}

func (s *CreatClusterTestSuite) SetupTest() {
	s.mocks.clusterClient = &mockClusterClient{
		createClusterResp: aTestCluster(),
		getClusterResp:    aTestCluster(withStatusSuccess),
	}
	s.mocks.nodepoolClient = &mockNodepoolClient{
		nodepoolListResp: []*models.VmwareTanzuManageV1alpha1AksclusterNodepoolNodepool{aTestNodePool()},
	}
	s.config = authctx.TanzuContext{
		TMCConnection: &client.TanzuMissionControl{
			AKSClusterResourceService:  s.mocks.clusterClient,
			AKSNodePoolResourceService: s.mocks.nodepoolClient,
		},
	}
	s.aksClusterResource = akscluster.ResourceTMCAKSCluster()
	s.ctx = context.WithValue(context.Background(), akscluster.RetryInterval, 10*time.Millisecond)
}

func (s *CreatClusterTestSuite) Test_resourceClusterCreate() {
	d := schema.TestResourceDataRaw(s.T(), akscluster.ClusterSchema, aTestClusterDataMap())
	expectedNP := aTestNodePool(forCluster(aTestCluster().FullName))

	result := s.aksClusterResource.CreateContext(s.ctx, d, s.config)

	s.Assert().False(result.HasError())
	s.Assert().True(s.mocks.clusterClient.AksCreateClusterWasCalled, "cluster create was not called")
	s.Assert().Equal(s.mocks.nodepoolClient.CreateNodepoolWasCalledWith, expectedNP, "nodepool create was not called ")
	s.Assert().Equal(expectedFullName(), s.mocks.clusterClient.AksClusterResourceServiceGetCalledWith)
	s.Assert().Equal(expectedFullName(), s.mocks.nodepoolClient.AksNodePoolResourceServiceListCalledWith)
	s.Assert().Equal("test-uid", d.Id())
}

func (s *CreatClusterTestSuite) Test_resourceClusterCreate_invalidConfig() {
	d := schema.TestResourceDataRaw(s.T(), akscluster.ClusterSchema, aTestClusterDataMap())

	result := s.aksClusterResource.CreateContext(s.ctx, d, nil)

	s.Assert().True(result.HasError())
}

func (s *CreatClusterTestSuite) Test_resourceClusterCreate_ClusterCreate_fails() {
	s.mocks.clusterClient.createErr = errors.New("create cluster failed")
	d := schema.TestResourceDataRaw(s.T(), akscluster.ClusterSchema, aTestClusterDataMap())

	result := s.aksClusterResource.CreateContext(s.ctx, d, s.config)

	s.Assert().True(result.HasError())
}

func (s *CreatClusterTestSuite) Test_resourceClusterCreate_NodepoolCreate_fails() {
	s.mocks.nodepoolClient.createErr = errors.New("create nodepool failed")
	d := schema.TestResourceDataRaw(s.T(), akscluster.ClusterSchema, aTestClusterDataMap())

	result := s.aksClusterResource.CreateContext(s.ctx, d, s.config)

	s.Assert().True(result.HasError())
}

func (s *CreatClusterTestSuite) Test_resourceClusterCreate_ClusterCreate_timeout() {
	s.mocks.clusterClient.getClusterResp = aTestCluster()
	d := schema.TestResourceDataRaw(s.T(), akscluster.ClusterSchema, aTestClusterDataMap(with5msTimeout))

	result := s.aksClusterResource.CreateContext(s.ctx, d, s.config)

	s.Assert().True(result.HasError())
}

func (s *CreatClusterTestSuite) Test_resourceClusterCreate_ClusterCreate_has_error_status() {
	s.mocks.clusterClient.getClusterResp = aTestCluster(withStatusError)
	d := schema.TestResourceDataRaw(s.T(), akscluster.ClusterSchema, aTestClusterDataMap())

	result := s.aksClusterResource.CreateContext(s.ctx, d, s.config)

	s.Assert().True(result.HasError())
}

func (s *CreatClusterTestSuite) Test_resourceClusterCreate_ClusterCreate_alreadyExists() {
	s.mocks.clusterClient.createErr = clienterrors.ErrorWithHTTPCode(http.StatusConflict, nil)
	d := schema.TestResourceDataRaw(s.T(), akscluster.ClusterSchema, aTestClusterDataMap())

	result := s.aksClusterResource.CreateContext(s.ctx, d, s.config)

	s.Assert().False(result.HasError())
	s.Assert().Equal("test-uid", d.Id())
}

func (s *CreatClusterTestSuite) Test_resourceClusterCreate_ClusterCreate_alreadyExists_but_notFound() {
	s.mocks.clusterClient.createErr = clienterrors.ErrorWithHTTPCode(http.StatusConflict, nil)
	s.mocks.clusterClient.getErr = clienterrors.ErrorWithHTTPCode(http.StatusNotFound, nil)
	d := schema.TestResourceDataRaw(s.T(), akscluster.ClusterSchema, aTestClusterDataMap())

	result := s.aksClusterResource.CreateContext(s.ctx, d, s.config)

	s.Assert().True(result.HasError())
	s.Assert().Empty(d.Id())
}

func (s *CreatClusterTestSuite) Test_resourceClusterCreate_ClusterCreate_succeeded_but_cluster_notFound() {
	s.mocks.clusterClient.getErr = clienterrors.ErrorWithHTTPCode(http.StatusNotFound, nil)
	d := schema.TestResourceDataRaw(s.T(), akscluster.ClusterSchema, aTestClusterDataMap())

	result := s.aksClusterResource.CreateContext(s.ctx, d, s.config)

	s.Assert().True(result.HasError())
	s.Assert().Empty(d.Id())
}

func (s *CreatClusterTestSuite) Test_resourceClusterCreate_ClusterCreate_succeeded_but_nodepools_notFound() {
	s.mocks.nodepoolClient.listErr = clienterrors.ErrorWithHTTPCode(http.StatusNotFound, nil)
	d := schema.TestResourceDataRaw(s.T(), akscluster.ClusterSchema, aTestClusterDataMap())

	result := s.aksClusterResource.CreateContext(s.ctx, d, s.config)

	s.Assert().True(result.HasError())
}

func (s *CreatClusterTestSuite) Test_resourceClusterCreate_ClusterCreate_all_system_pools_fail() {
	nodepools := []any{aTestNodepoolDataMap(withNodepoolMode("USER")), aTestNodepoolDataMap(withNodepoolMode("SYSTEM"))}
	cluster := aTestClusterDataMap(withNodepools(nodepools))
	s.mocks.nodepoolClient.failSystemPools = true
	d := schema.TestResourceDataRaw(s.T(), akscluster.ClusterSchema, cluster)

	result := s.aksClusterResource.CreateContext(s.ctx, d, s.config)

	s.Assert().True(result.HasError())
	s.Assert().Equal("no system nodepools were successfully created.", result[0].Summary)
}

func (s *CreatClusterTestSuite) Test_resourceClusterCreate_ClusterCreate_no_system_nodepool() {
	userpool := []any{aTestNodepoolDataMap(withNodepoolMode("USER"))}
	cluster := aTestClusterDataMap(withNodepools(userpool))
	d := schema.TestResourceDataRaw(s.T(), akscluster.ClusterSchema, cluster)

	result := s.aksClusterResource.CreateContext(s.ctx, d, s.config)

	s.Assert().True(result.HasError())
}

type ReadClusterTestSuite struct {
	suite.Suite
	ctx                context.Context
	mocks              mocks
	aksClusterResource *schema.Resource
	config             authctx.TanzuContext
}

func (s *ReadClusterTestSuite) SetupTest() {
	s.mocks.clusterClient = &mockClusterClient{
		createClusterResp: aTestCluster(),
		getClusterResp:    aTestCluster(withStatusSuccess),
	}
	s.mocks.nodepoolClient = &mockNodepoolClient{
		nodepoolListResp: []*models.VmwareTanzuManageV1alpha1AksclusterNodepoolNodepool{aTestNodePool()},
	}
	s.config = authctx.TanzuContext{
		TMCConnection: &client.TanzuMissionControl{
			AKSClusterResourceService:  s.mocks.clusterClient,
			AKSNodePoolResourceService: s.mocks.nodepoolClient,
		},
	}
	s.aksClusterResource = akscluster.ResourceTMCAKSCluster()
	s.ctx = context.WithValue(context.Background(), akscluster.RetryInterval, 10*time.Millisecond)
}

func (s *ReadClusterTestSuite) Test_resourceClusterRead() {
	d := schema.TestResourceDataRaw(s.T(), akscluster.ClusterSchema, aTestClusterDataMap())

	result := s.aksClusterResource.ReadContext(s.ctx, d, s.config)

	s.Assert().False(result.HasError())
	s.Assert().Equal(expectedFullName(), s.mocks.clusterClient.AksClusterResourceServiceGetCalledWith)
	s.Assert().Equal(expectedFullName(), s.mocks.nodepoolClient.AksNodePoolResourceServiceListCalledWith)
	s.Assert().Equal("test-uid", d.Id(), "expect id from REST request")
	s.Assert().NotNil(d.Get(common.MetaKey), "expected metadata from REST request")
	s.Assert().NotNil(d.Get("spec"), "expected cluster spec from REST request")
}

type UpdateClusterTestSuite struct {
	suite.Suite
	ctx                context.Context
	mocks              mocks
	aksClusterResource *schema.Resource
	config             authctx.TanzuContext
}

func (s *UpdateClusterTestSuite) SetupTest() {
	s.mocks.clusterClient = &mockClusterClient{
		createClusterResp: aTestCluster(),
		getClusterResp:    aTestCluster(withStatusSuccess),
	}
	s.mocks.nodepoolClient = &mockNodepoolClient{
		nodepoolListResp: []*models.VmwareTanzuManageV1alpha1AksclusterNodepoolNodepool{aTestNodePool(forCluster(aTestCluster().FullName))},
		nodepoolGetResp:  aTestNodePool(forCluster(aTestCluster().FullName), withNodepoolStatusSuccess),
	}
	s.config = authctx.TanzuContext{
		TMCConnection: &client.TanzuMissionControl{
			AKSClusterResourceService:  s.mocks.clusterClient,
			AKSNodePoolResourceService: s.mocks.nodepoolClient,
		},
	}
	s.aksClusterResource = akscluster.ResourceTMCAKSCluster()
	s.ctx = context.WithValue(context.Background(), akscluster.RetryInterval, 10*time.Millisecond)
}

func (s *UpdateClusterTestSuite) Test_resourceClusterUpdate_updateClusterConfig() {
	originalCluster := aTestClusterDataMap(withDNSPrefix("new-prefix1"))
	updatedCluster := aTestClusterDataMap(withDNSPrefix("new-prefix2"))
	d := dataDiffFrom(s.T(), originalCluster, updatedCluster)
	expected := aTestCluster()
	expected.Spec.Config.NetworkConfig.DNSPrefix = "new-prefix2"

	result := s.aksClusterResource.UpdateContext(s.ctx, d, s.config)

	s.Assert().False(result.HasError())
	s.Assert().Equal(expected, s.mocks.clusterClient.AksUpdateClusterWasCalledWith)
	s.Assert().Nil(s.mocks.nodepoolClient.UpdatedNodepoolWasCalledWith)
}

func (s *UpdateClusterTestSuite) Test_resourceClusterUpdate_updateNodepool() {
	originalNodepools := []any{aTestNodepoolDataMap(withNodepoolCount(1))}
	updatedNodepools := []any{aTestNodepoolDataMap(withNodepoolCount(5))}
	d := dataDiffFrom(s.T(), aTestClusterDataMap(withNodepools(originalNodepools)), aTestClusterDataMap(withNodepools(updatedNodepools)))
	expected := aTestNodePool(forCluster(aTestCluster().FullName))
	expected.Spec.Count = 5

	result := s.aksClusterResource.UpdateContext(s.ctx, d, s.config)

	s.Assert().False(result.HasError())
	s.Assert().Nil(s.mocks.clusterClient.AksUpdateClusterWasCalledWith)
	s.Assert().Equal(expected, s.mocks.nodepoolClient.UpdatedNodepoolWasCalledWith)
}

func (s *UpdateClusterTestSuite) Test_resourceClusterUpdate_addNodepool() {
	s.mocks.nodepoolClient.nodepoolListResp = []*models.VmwareTanzuManageV1alpha1AksclusterNodepoolNodepool{
		aTestNodePool(forCluster(aTestCluster().FullName), withNodepoolName("np1")),
	}
	originalNodepools := []any{aTestNodepoolDataMap(withName("np1"))}
	updatedNodepools := []any{aTestNodepoolDataMap(withName("np1")), aTestNodepoolDataMap(withName("np2"))}
	d := dataDiffFrom(s.T(), aTestClusterDataMap(withNodepools(originalNodepools)), aTestClusterDataMap(withNodepools(updatedNodepools)))
	expected := aTestNodePool(forCluster(aTestCluster().FullName), withNodepoolName("np2"))

	result := s.aksClusterResource.UpdateContext(s.ctx, d, s.config)

	s.Assert().False(result.HasError())
	s.Assert().Nil(s.mocks.clusterClient.AksUpdateClusterWasCalledWith)
	s.Assert().Equal(expected, s.mocks.nodepoolClient.CreateNodepoolWasCalledWith)
}

func (s *UpdateClusterTestSuite) Test_resourceClusterUpdate_deleteNodepool() {
	s.mocks.nodepoolClient.nodepoolListResp = []*models.VmwareTanzuManageV1alpha1AksclusterNodepoolNodepool{
		aTestNodePool(forCluster(aTestCluster().FullName), withNodepoolName("np1")),
		aTestNodePool(forCluster(aTestCluster().FullName), withNodepoolName("np2"))}
	originalNodepools := []any{aTestNodepoolDataMap(withName("np1")), aTestNodepoolDataMap(withName("np2"))}
	updatedNodepools := []any{aTestNodepoolDataMap(withName("np1")), nil}
	s.mocks.nodepoolClient.getErr = clienterrors.ErrorWithHTTPCode(http.StatusNotFound, nil)
	expected := aTestNodePool(forCluster(aTestCluster().FullName), withNodepoolName("np2")).FullName

	d := dataDiffFrom(s.T(), aTestClusterDataMap(withNodepools(originalNodepools)), aTestClusterDataMap(withNodepools(updatedNodepools)))

	result := s.aksClusterResource.UpdateContext(s.ctx, d, s.config)

	s.Assert().False(result.HasError())
	s.Assert().Nil(s.mocks.clusterClient.AksUpdateClusterWasCalledWith)
	s.Assert().Equal(expected, s.mocks.nodepoolClient.DeleteNodepoolWasCalledWith)
}

func (s *UpdateClusterTestSuite) Test_resourceClusterUpdate_invalidConfig() {
	d := schema.TestResourceDataRaw(s.T(), akscluster.ClusterSchema, aTestClusterDataMap())

	result := s.aksClusterResource.UpdateContext(s.ctx, d, "config")

	s.Assert().True(result.HasError())
}

func (s *UpdateClusterTestSuite) Test_resourceClusterUpdate_updateClusterFails() {
	originalCluster := aTestClusterDataMap(withDNSPrefix("new-prefix1"))
	updatedCluster := aTestClusterDataMap(withDNSPrefix("new-prefix2"))
	s.mocks.clusterClient.updateErr = errors.New("failed to update cluster")
	d := dataDiffFrom(s.T(), originalCluster, updatedCluster)

	result := s.aksClusterResource.UpdateContext(s.ctx, d, s.config)

	s.Assert().True(result.HasError())
}

func (s *UpdateClusterTestSuite) Test_resourceClusterUpdate_updateClusterTimeout() {
	originalCluster := aTestClusterDataMap(withDNSPrefix("new-prefix1"))
	updatedCluster := aTestClusterDataMap(withDNSPrefix("new-prefix2"), with5msTimeout)
	s.mocks.clusterClient.getClusterResp = aTestCluster() // without success status
	d := dataDiffFrom(s.T(), originalCluster, updatedCluster)

	result := s.aksClusterResource.UpdateContext(s.ctx, d, s.config)

	s.Assert().True(result.HasError())
}

func (s *UpdateClusterTestSuite) Test_resourceClusterUpdate_updateNodepoolFails() {
	s.mocks.nodepoolClient.nodepoolListResp = []*models.VmwareTanzuManageV1alpha1AksclusterNodepoolNodepool{
		aTestNodePool(forCluster(aTestCluster().FullName), withNodepoolName("np1"), withCount(1)),
	}
	originalNodepools := []any{aTestNodepoolDataMap(withName("np1"), withNodepoolCount(1))}
	updatedNodepools := []any{aTestNodepoolDataMap(withName("np1"), withNodepoolCount(5))}
	s.mocks.nodepoolClient.updateErr = errors.New("failed to update nodepool")
	d := dataDiffFrom(s.T(), aTestClusterDataMap(withNodepools(originalNodepools)), aTestClusterDataMap(withNodepools(updatedNodepools)))

	result := s.aksClusterResource.UpdateContext(s.ctx, d, s.config)

	s.Assert().True(result.HasError())
}

func (s *UpdateClusterTestSuite) Test_resourceClusterUpdate_updateNodepoolTimeout() {
	s.mocks.nodepoolClient.nodepoolListResp = []*models.VmwareTanzuManageV1alpha1AksclusterNodepoolNodepool{
		aTestNodePool(forCluster(aTestCluster().FullName), withNodepoolName("np1"), withCount(1)),
	}
	originalNodepools := []any{aTestNodepoolDataMap(withName("np1"), withNodepoolCount(1))}
	updatedNodepools := []any{aTestNodepoolDataMap(withName("np1"), withNodepoolCount(5))}
	s.mocks.nodepoolClient.nodepoolGetResp = aTestNodePool() // Without success
	d := dataDiffFrom(s.T(), aTestClusterDataMap(withNodepools(originalNodepools)), aTestClusterDataMap(withNodepools(updatedNodepools), with5msTimeout))

	result := s.aksClusterResource.UpdateContext(s.ctx, d, s.config)

	s.Assert().True(result.HasError())
}

func (s *UpdateClusterTestSuite) Test_resourceClusterUpdate_createNodepoolFails() {
	originalNodepools := []any{aTestNodepoolDataMap()}
	updatedNodepools := []any{aTestNodepoolDataMap(), aTestNodepoolDataMap(withName("np2"))}
	s.mocks.nodepoolClient.createErr = errors.New("failed to create nodepool")
	d := dataDiffFrom(s.T(), aTestClusterDataMap(withNodepools(originalNodepools)), aTestClusterDataMap(withNodepools(updatedNodepools)))

	result := s.aksClusterResource.UpdateContext(s.ctx, d, s.config)

	s.Assert().True(result.HasError())
}

func (s *UpdateClusterTestSuite) Test_resourceClusterUpdate_createNodepoolError() {
	originalNodepools := []any{aTestNodepoolDataMap()}
	updatedNodepools := []any{aTestNodepoolDataMap(), aTestNodepoolDataMap(withName("np2"))}
	s.mocks.nodepoolClient.nodepoolGetResp = aTestNodePool(withNodepoolStatusError())
	d := dataDiffFrom(s.T(), aTestClusterDataMap(withNodepools(originalNodepools)), aTestClusterDataMap(withNodepools(updatedNodepools)))

	result := s.aksClusterResource.UpdateContext(s.ctx, d, s.config)

	s.Assert().True(result.HasError())
}

func (s *UpdateClusterTestSuite) Test_resourceClusterUpdate_createNodepoolTimeout() {
	originalNodepools := []any{aTestNodepoolDataMap()}
	updatedNodepools := []any{aTestNodepoolDataMap(), aTestNodepoolDataMap(withName("np2"))}
	d := dataDiffFrom(s.T(), aTestClusterDataMap(withNodepools(originalNodepools)), aTestClusterDataMap(withNodepools(updatedNodepools), with5msTimeout))

	result := s.aksClusterResource.UpdateContext(s.ctx, d, s.config)

	s.Assert().True(result.HasError())
}

func (s *UpdateClusterTestSuite) Test_resourceClusterUpdate_deleteNodepoolFails() {
	s.mocks.nodepoolClient.nodepoolListResp = []*models.VmwareTanzuManageV1alpha1AksclusterNodepoolNodepool{
		aTestNodePool(forCluster(aTestCluster().FullName), withNodepoolName("np1")),
		aTestNodePool(forCluster(aTestCluster().FullName), withNodepoolName("np2"))}
	originalNodepools := []any{aTestNodepoolDataMap(withName("np1")), aTestNodepoolDataMap(withName("np2"))}
	updatedNodepools := []any{aTestNodepoolDataMap(withName("np1")), nil}
	s.mocks.nodepoolClient.DeleteErr = errors.New("failed to delete nodepool")
	d := dataDiffFrom(s.T(), aTestClusterDataMap(withNodepools(originalNodepools)), aTestClusterDataMap(withNodepools(updatedNodepools)))

	result := s.aksClusterResource.UpdateContext(s.ctx, d, s.config)

	s.Assert().True(result.HasError())
}

func (s *UpdateClusterTestSuite) Test_resourceClusterUpdate_deleteNodepoolTimeout() {
	s.mocks.nodepoolClient.nodepoolListResp = []*models.VmwareTanzuManageV1alpha1AksclusterNodepoolNodepool{
		aTestNodePool(forCluster(aTestCluster().FullName), withNodepoolName("np1")),
		aTestNodePool(forCluster(aTestCluster().FullName), withNodepoolName("np2"))}
	originalNodepools := []any{aTestNodepoolDataMap(withName("np1")), aTestNodepoolDataMap(withName("np2"))}
	updatedNodepools := []any{aTestNodepoolDataMap(withName("np1")), nil}

	d := dataDiffFrom(s.T(), aTestClusterDataMap(withNodepools(originalNodepools)), aTestClusterDataMap(withNodepools(updatedNodepools), with5msTimeout))

	result := s.aksClusterResource.UpdateContext(s.ctx, d, s.config)

	s.Assert().True(result.HasError())
}

func (s *UpdateClusterTestSuite) Test_resourceClusterUpdate_nodepoolOrderChange() {
	s.mocks.nodepoolClient.nodepoolListResp = []*models.VmwareTanzuManageV1alpha1AksclusterNodepoolNodepool{
		aTestNodePool(withNodepoolName("np1")),
		aTestNodePool(withNodepoolName("np2")),
	}
	originalNodepools := []any{aTestNodepoolDataMap(withName("np1")), aTestNodepoolDataMap(withName("np2"))}
	updatedNodepools := []any{aTestNodepoolDataMap(withName("np2")), aTestNodepoolDataMap(withName("np1"))}
	d := dataDiffFrom(s.T(), aTestClusterDataMap(withNodepools(originalNodepools)), aTestClusterDataMap(withNodepools(updatedNodepools)))

	result := s.aksClusterResource.UpdateContext(s.ctx, d, s.config)

	s.Assert().False(result.HasError())
	s.Assert().Nil(s.mocks.nodepoolClient.DeleteNodepoolWasCalledWith)
	s.Assert().Nil(s.mocks.nodepoolClient.CreateNodepoolWasCalledWith)
	s.Assert().Nil(s.mocks.nodepoolClient.UpdatedNodepoolWasCalledWith)
}

func (s *UpdateClusterTestSuite) Test_resourceClusterUpdate_nodepoolImmutableChange_recreate() {
	s.mocks.nodepoolClient.nodepoolListResp = []*models.VmwareTanzuManageV1alpha1AksclusterNodepoolNodepool{
		aTestNodePool(forCluster(aTestCluster().FullName), withNodepoolName("np1")),
	}
	s.mocks.nodepoolClient.getErr = clienterrors.ErrorWithHTTPCode(http.StatusNotFound, nil)

	originalNodepools := []any{aTestNodepoolDataMap(withName("np1"))}
	updatedNodepools := []any{aTestNodepoolDataMap(withName("np1"), withNodepoolVMSize("STANDARD_DS2v3"))}
	d := dataDiffFrom(s.T(), aTestClusterDataMap(withNodepools(originalNodepools)), aTestClusterDataMap(withNodepools(updatedNodepools)))
	expected := aTestNodePool(forCluster(aTestCluster().FullName), withNodepoolName("np1"))
	expected.Spec.VMSize = "STANDARD_DS2v3"

	result := s.aksClusterResource.UpdateContext(s.ctx, d, s.config)

	s.Assert().True(result.HasError())
	s.Assert().Equal(s.mocks.nodepoolClient.DeleteNodepoolWasCalledWith, expected.FullName)
	s.Assert().Equal(s.mocks.nodepoolClient.CreateNodepoolWasCalledWith, expected)
	s.Assert().Nil(s.mocks.nodepoolClient.UpdatedNodepoolWasCalledWith)
}

type DeleteClusterTestSuite struct {
	suite.Suite
	ctx                context.Context
	mocks              mocks
	aksClusterResource *schema.Resource
	config             authctx.TanzuContext
}

func (s *DeleteClusterTestSuite) SetupTest() {
	s.mocks.clusterClient = &mockClusterClient{
		createClusterResp: aTestCluster(),
		getClusterResp:    aTestCluster(withStatusSuccess),
	}
	s.config = authctx.TanzuContext{
		TMCConnection: &client.TanzuMissionControl{
			AKSClusterResourceService: s.mocks.clusterClient,
		},
	}
	s.aksClusterResource = akscluster.ResourceTMCAKSCluster()
	s.ctx = context.WithValue(context.Background(), akscluster.RetryInterval, 10*time.Millisecond)
}

func (s *DeleteClusterTestSuite) Test_resourceClusterDelete() {
	s.mocks.clusterClient.getErr = clienterrors.ErrorWithHTTPCode(http.StatusNotFound, nil)
	d := schema.TestResourceDataRaw(s.T(), akscluster.ClusterSchema, aTestClusterDataMap())

	result := s.aksClusterResource.DeleteContext(s.ctx, d, s.config)

	s.Assert().False(result.HasError())
	s.Assert().Equal(expectedFullName(), s.mocks.clusterClient.AksClusterResourceServiceDeleteCalledWith)
}

func (s *DeleteClusterTestSuite) Test_resourceClusterDelete_invalidConfig() {
	d := schema.TestResourceDataRaw(s.T(), akscluster.ClusterSchema, aTestClusterDataMap())

	result := s.aksClusterResource.DeleteContext(s.ctx, d, "config")

	s.Assert().True(result.HasError())
}

func (s *DeleteClusterTestSuite) Test_resourceClusterDelete_fails() {
	s.mocks.clusterClient.deleteErr = errors.New("cluster delete failed")
	d := schema.TestResourceDataRaw(s.T(), akscluster.ClusterSchema, aTestClusterDataMap())

	result := s.aksClusterResource.DeleteContext(s.ctx, d, s.config)

	s.Assert().True(result.HasError())
	s.Assert().Equal(expectedFullName(), s.mocks.clusterClient.AksClusterResourceServiceDeleteCalledWith)
}

func (s *DeleteClusterTestSuite) Test_resourceClusterDelete_timeout() {
	d := schema.TestResourceDataRaw(s.T(), akscluster.ClusterSchema, aTestClusterDataMap(with5msTimeout))

	result := s.aksClusterResource.DeleteContext(s.ctx, d, s.config)

	s.Assert().True(result.HasError())
	s.Assert().Equal(expectedFullName(), s.mocks.clusterClient.AksClusterResourceServiceDeleteCalledWith)
}

type ImportClusterTestSuite struct {
	suite.Suite
	ctx                context.Context
	mocks              mocks
	aksClusterResource *schema.Resource
	config             authctx.TanzuContext
}

func (s *ImportClusterTestSuite) SetupTest() {
	s.mocks.clusterClient = &mockClusterClient{
		getClusterByIDResp: aTestCluster(withStatusSuccess),
	}
	s.mocks.nodepoolClient = &mockNodepoolClient{
		nodepoolListResp: []*models.VmwareTanzuManageV1alpha1AksclusterNodepoolNodepool{aTestNodePool(forCluster(aTestCluster().FullName))},
	}
	s.config = authctx.TanzuContext{
		TMCConnection: &client.TanzuMissionControl{
			AKSClusterResourceService:  s.mocks.clusterClient,
			AKSNodePoolResourceService: s.mocks.nodepoolClient,
		},
	}
	s.aksClusterResource = akscluster.ResourceTMCAKSCluster()
	s.ctx = context.WithValue(context.Background(), akscluster.RetryInterval, 10*time.Millisecond)
}

func (s *ImportClusterTestSuite) Test_resourceClusterImport() {
	d := schema.TestResourceDataRaw(s.T(), akscluster.ClusterSchema, nil)
	d.SetId("test-id")

	result, err := s.aksClusterResource.Importer.StateContext(s.ctx, d, s.config)

	s.Assert().NoError(err)
	s.Assert().Len(result, 1)
	cluster := akscluster.ConstructCluster(result[0])
	s.Assert().Equal(cluster.FullName.Name, "test-cluster")
	s.Assert().Equal(cluster.FullName.CredentialName, "test-cred")
	s.Assert().Equal(cluster.FullName.SubscriptionID, "sub-id")
	s.Assert().Equal(cluster.FullName.ResourceGroupName, "resource-group")
	s.Assert().NotNil(cluster.Spec)
	s.Assert().NotNil(cluster.Meta)
}

func (s *ImportClusterTestSuite) Test_resourceClusterImport_GetClusterFails() {
	s.mocks.clusterClient.getErr = errors.New("failed to get cluster by ID")
	d := schema.TestResourceDataRaw(s.T(), akscluster.ClusterSchema, nil)
	d.SetId("test-id")

	_, err := s.aksClusterResource.Importer.StateContext(s.ctx, d, s.config)

	s.Assert().Error(err)
}

func (s *ImportClusterTestSuite) Test_resourceClusterImport_GetNodepoolsFails() {
	s.mocks.nodepoolClient.listErr = errors.New("failed to get nodepools")
	d := schema.TestResourceDataRaw(s.T(), akscluster.ClusterSchema, nil)
	d.SetId("test-id")

	_, err := s.aksClusterResource.Importer.StateContext(s.ctx, d, s.config)

	s.Assert().Error(err)
}
