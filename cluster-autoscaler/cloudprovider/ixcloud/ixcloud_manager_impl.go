package ixcloud

import (
	"fmt"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/ixcloud/iksclient"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/klog/v2"
)

// ixCloudManagerImpl implements the iksManager interface.
type ixCloudManagerImpl struct {
	iksApiClient *iksclient.IksApiClient

	clusterName string

	providerIDToNodeGroupCache map[string]string
}

func createIxCloudManagerImpl(iksApiClient *iksclient.IksApiClient, opts config.AutoscalingOptions) (*ixCloudManagerImpl, error) {
	manager := ixCloudManagerImpl{
		iksApiClient: iksApiClient,
		clusterName:  opts.ClusterName,
	}

	return &manager, nil
}

func uniqueName(ng iksclient.NodeGroup) string {
	name := ng.Name
	id := ng.ID

	idSegment := strings.Split(id, "-")[0]
	uniqueName := fmt.Sprintf("%s-%s", name, idSegment)

	return uniqueName
}

// TODO: nodeGroupID가 uuid가 아니라 노드네임으로 검색할 수 있도록 해야함 아니면 전체를 불러오고 검색해서 찾는 방법 사용
func (mgr *ixCloudManagerImpl) uniqueNameAndIDForNodeGroup(nodeGroupID string) (string, string, error) {
	ng, err := iksclient.GetNodeGroup(mgr.iksApiClient, nodeGroupID)
	if err != nil {
		return "", "", fmt.Errorf("could not get node group: %v", err)
	}

	uniqueName := uniqueName(ng)

	return uniqueName, ng.ID, nil
}

// nodeGroupSize gets the current node count of the given node group.
func (mgr *ixCloudManagerImpl) nodeGroupSize(nodeGroupID string) (int, error) {
	ng, err := iksclient.GetNodeGroup(mgr.iksApiClient, nodeGroupID)
	if err != nil {
		return 0, fmt.Errorf("could not get node group: %v", err)
	}
	return ng.CurrentSize, nil
}

// getNodes returns Instances with ProviderIDs and running states
// of all nodes that exist in OpenStack for a node group.
func (mgr *ixCloudManagerImpl) getNodes(nodeGroupID string) ([]cloudprovider.Instance, error) {
	var nodes []cloudprovider.Instance

	//TODO: implement getNodes
	//TODO: /k8s/node_groups/{nodeGroupId} API 에서 NODE 정보 추출 (status 확인하는 이유 체크 필요)

	return nodes, nil
}

// deleteNodes deletes nodes by resizing the cluster to a smaller size
// and specifying which nodes should be removed.
//
// The nodes are referenced by server ID for nodes which have them,
// or by the minion index for nodes which are creating or in an error state.
func (mgr *ixCloudManagerImpl) deleteNodes(nodeGroupID string, nodes []NodeRef, updatedNodeCount int) error {
	var nodesToRemove []string
	for _, nodeRef := range nodes {
		if nodeRef.IsFake {
			_, index, err := parseFakeProviderID(nodeRef.Name)
			if err != nil {
				return fmt.Errorf("error handling fake node: %v", err)
			}
			nodesToRemove = append(nodesToRemove, index)
			continue
		}
		klog.V(2).Infof("manager deleting node: %s", nodeRef.Name)
		nodesToRemove = append(nodesToRemove, nodeRef.SystemUUID)
	}

	resizeOpts := iksclient.ResizeOpts{
		NodeCount:     &updatedNodeCount,
		NodesToRemove: nodesToRemove,
		NodeGroup:     nodeGroupID,
	}

	_, err := iksclient.Resize(mgr.iksApiClient, nodeGroupID, resizeOpts)
	if err != nil {
		return fmt.Errorf("could not resize cluster: %v", err)
	}

	return nil
}

// nodeGroupForNode returns the UUID of the node group that the given node is a member of.
func (mgr *ixCloudManagerImpl) nodeGroupForNode(node *apiv1.Node) (string, error) {
	//TODO: implement nodeGroupForNode
	if groupUUID, ok := mgr.providerIDToNodeGroupCache[node.Spec.ProviderID]; ok {
		klog.V(5).Infof("nodeGroupForNode: already cached %s in node group %s", node.Spec.ProviderID, groupUUID)
		return groupUUID, nil
	}

	//모든 노드그룹 조회(/k8s/node_groups) -> 노드그룹의 노드 조회(/k8s/node_groups/{nodeGroupId}/nodes) -> 일치하는 노드그룹아이디 리턴
	//providerIDToNodeGroupCache 정보도 갱신
	return "", nil
}

// updateNodeCount performs a node group resize targeting the given node group.
func (mgr *ixCloudManagerImpl) updateNodeCount(nodeGroupID string, nodes int) error {
	resizeOpts := iksclient.ResizeOpts{
		NodeCount: &nodes,
		NodeGroup: nodeGroupID,
	}

	_, err := iksclient.Resize(mgr.iksApiClient, mgr.clusterName, resizeOpts)
	if err != nil {
		return fmt.Errorf("could not resize cluster: %v", err)
	}
	return nil
}

// autoDiscoverNodeGroups lists all node groups that belong to this cluster
// and finds the ones which are valid for autoscaling and that match the
// auto discovery configuration.
func (mgr *ixCloudManagerImpl) autoDiscoverNodeGroups(cgfs []ixCloudAutoDiscoveryConfig) ([]*iksclient.NodeGroup, error) {
	ngs := []*iksclient.NodeGroup{}

	// TODO: implement autoDiscoverNodeGroups
	// nodegroups.List
	return ngs, nil
}
