package ixcloud

import (
	"fmt"
	"io"

	"gopkg.in/gcfg.v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/ixcloud/iksclient"

	"k8s.io/autoscaler/cluster-autoscaler/config"
)

// ixCloudManager is an interface for the basic interactions with the cluster.
type ixCloudManager interface {
	nodeGroupSize(nodeGroupID string) (int, error)
	updateNodeCount(nodeGroupID string, nodes int) error
	getNodes(nodeGroupID string) ([]cloudprovider.Instance, error)
	deleteNodes(nodeGroupID string, nodes []NodeRef, updatedNodeCount int) error
	uniqueNameAndIDForNodeGroup(nodeGroupID string) (string, string, error)
	nodeGroupForNode(node *apiv1.Node) (string, error)
	autoDiscoverNodeGroups(configs []ixCloudAutoDiscoveryConfig) ([]*iksclient.NodeGroup, error)
}

func createIxCloudManager(configReader io.Reader, opts config.AutoscalingOptions) (ixCloudManager, error) {
	cfg, err := readConfig(configReader)
	if err != nil {
		return nil, err
	}

	iksApiClient, err := iksclient.CreateIksApiClient(cfg, opts)
	if err != nil {
		return nil, fmt.Errorf("could not create iks api client: %v", err)
	}

	return createIxCloudManagerImpl(iksApiClient, opts)
}

// readConfig parses an OpenStack cloud-config file from an io.Reader.
func readConfig(configReader io.Reader) (*iksclient.Config, error) {
	var cfg iksclient.Config
	if configReader != nil {
		if err := gcfg.ReadInto(&cfg, configReader); err != nil {
			return nil, fmt.Errorf("couldn't read cloud config: %v", err)
		}
	}
	return &cfg, nil
}
