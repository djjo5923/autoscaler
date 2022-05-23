package ixcloud

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/klog/v2"
)

const (
	// GPULabel is the label added to nodes with GPU resource.
	// TODO: fix gpu label
	GPULabel = "ixcloud.openstack.org/gpu"

	// Refresh interval for node group auto discovery
	discoveryRefreshInterval = 1 * time.Minute
)

type ixCloudProvider struct {
	ixCloudManager  ixCloudManager
	resourceLimiter *cloudprovider.ResourceLimiter

	nodeGroups []*ixCloudNodeGroup

	// To be locked when modifying or reading the node groups slice.
	nodeGroupsLock *sync.Mutex

	// To be locked when modifying or reading the cluster state from ixcloud.
	clusterUpdateLock *sync.Mutex

	usingAutoDiscovery   bool
	autoDiscoveryConfigs []ixCloudAutoDiscoveryConfig
	lastDiscoveryRefresh time.Time
}

func buildIxCloudProvider(ixCloudManager ixCloudManager, resourceLimiter *cloudprovider.ResourceLimiter) (*ixCloudProvider, error) {
	ixcp := &ixCloudProvider{
		ixCloudManager:  ixCloudManager,
		resourceLimiter: resourceLimiter,
		nodeGroups:      []*ixCloudNodeGroup{},
		nodeGroupsLock:  &sync.Mutex{},
	}
	return ixcp, nil
}

func (ixcp *ixCloudProvider) Name() string {
	return cloudprovider.IxCloudProviderName
}

// NodeGroups returns all node groups managed by this cloud provider.
func (ixcp *ixCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	ixcp.nodeGroupsLock.Lock()
	defer ixcp.nodeGroupsLock.Unlock()

	// Have to convert to a slice of the NodeGroup interface type.
	groups := make([]cloudprovider.NodeGroup, len(ixcp.nodeGroups))
	for i, group := range ixcp.nodeGroups {
		groups[i] = group
	}
	return groups
}

// AddNodeGroup appends a node group to the list of node groups managed by this cloud provider.
func (ixcp *ixCloudProvider) AddNodeGroup(group *ixCloudNodeGroup) {
	ixcp.nodeGroupsLock.Lock()
	defer ixcp.nodeGroupsLock.Unlock()
	ixcp.nodeGroups = append(ixcp.nodeGroups, group)
}

func (ixcp *ixCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	ixcp.nodeGroupsLock.Lock()
	defer ixcp.nodeGroupsLock.Unlock()

	// Ignore master node
	if _, found := node.ObjectMeta.Labels["node-role.kubernetes.io/master"]; found {
		return nil, nil
	}

	ngUUID, err := ixcp.ixCloudManager.nodeGroupForNode(node)
	if err != nil {
		return nil, fmt.Errorf("error finding node group UUID for node %s: %v", node.Spec.ProviderID, err)
	}

	for _, group := range ixcp.nodeGroups {
		if group.UUID == ngUUID {
			klog.V(4).Infof("Node %s belongs to node group %s", node.Spec.ProviderID, group.Id())
			return group, nil
		}
	}

	klog.V(4).Infof("Node %s is not part of an autoscaled node group", node.Spec.ProviderID)

	return nil, nil
}

// Pricing is not implemented.
func (ixcp *ixCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes is not implemented.
func (ixcp *ixCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup is not implemented.
func (ixcp *ixCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string, taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns resource constraints for the cloud provider
func (ixcp *ixCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return ixcp.resourceLimiter, nil
}

func (ixcp *ixCloudProvider) GPULabel() string {
	return GPULabel
}

// GetAvailableGPUTypes is not implemented.
func (ixcp *ixCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return map[string]struct{}{}
}

// Cleanup currently does nothing.
func (ixcp *ixCloudProvider) Cleanup() error {
	return nil
}

// Refresh is called before every autoscaler main loop.
//
// Debug information for each node group is printed with logging level >= 5.
// Every 60 seconds the node group state on the ixcloud side is checked,
// to see if there are any node groups that need to be added/removed/updated.
func (ixcp *ixCloudProvider) Refresh() error {
	ixcp.nodeGroupsLock.Lock()
	for _, nodegroup := range ixcp.nodeGroups {
		klog.V(5).Info(nodegroup.Debug())
	}
	ixcp.nodeGroupsLock.Unlock()

	if ixcp.usingAutoDiscovery {
		if time.Since(ixcp.lastDiscoveryRefresh) > discoveryRefreshInterval {
			ixcp.lastDiscoveryRefresh = time.Now()
			err := ixcp.refreshNodeGroups()
			if err != nil {
				return fmt.Errorf("error refreshing node groups: %v", err)
			}
		}
	}

	return nil
}

// refreshNodeGroups gets the list of node groups which meet the requirements for autoscaling,
// creates ixCloudNodeGroups for any that do not exist in the cloud provider,
// and drops any node groups which are present in the cloud provider but not in the
// list of node groups that should be autoscaled.
//
// Any node groups which have had their min/max node count updated in ixcloud
// are updated with the new limits.
func (ixcp *ixCloudProvider) refreshNodeGroups() error {
	ixcp.clusterUpdateLock.Lock()
	defer ixcp.clusterUpdateLock.Unlock()

	// Get the list of node groups that match the auto discovery configuration and
	// meet the requirements for autoscaling.
	nodeGroups, err := ixcp.ixCloudManager.autoDiscoverNodeGroups(ixcp.autoDiscoveryConfigs)
	if err != nil {
		return fmt.Errorf("could not discover node groups: %v", err)
	}

	// Track names of node groups which are added or removed (for logging).
	var newNodeGroupNames []string
	var droppedNodeGroupNames []string

	// Use maps for easier lookups of node group names.

	// Node group names as registered in the autoscaler.
	registeredNGs := make(map[string]*ixCloudNodeGroup)
	ixcp.nodeGroupsLock.Lock()
	for _, ng := range ixcp.nodeGroups {
		registeredNGs[ng.UUID] = ng
	}
	ixcp.nodeGroupsLock.Unlock()

	// Node group names that exist on the cloud side and should be autoscaled.
	autoscalingNGs := make(map[string]string)

	for _, nodeGroup := range nodeGroups {
		name := uniqueName(*nodeGroup)

		// Just need the name in the key.
		autoscalingNGs[nodeGroup.ID] = ""

		if ng, alreadyRegistered := registeredNGs[nodeGroup.ID]; alreadyRegistered {
			// Node group exists in autoscaler and in cloud, only need to check if min/max node count have changed.
			if ng.minSize != nodeGroup.MinNodeCount {
				ng.minSize = nodeGroup.MinNodeCount
				klog.V(2).Infof("Node group %s min node count changed to %d", nodeGroup.Name, ng.minSize)
			}
			// Node groups with unset max node count are not eligible for autoscaling, so this deference is safe.
			if ng.maxSize != *nodeGroup.MaxNodeCount {
				ng.maxSize = *nodeGroup.MaxNodeCount
				klog.V(2).Infof("Node group %s max node count changed to %d", nodeGroup.Name, ng.maxSize)
			}
			continue
		}

		// The node group is not known to the autoscaler, so create it.
		ng := &ixCloudNodeGroup{
			ixCloudManager:    ixcp.ixCloudManager,
			id:                name,
			UUID:              nodeGroup.ID,
			clusterUpdateLock: ixcp.clusterUpdateLock,
			minSize:           nodeGroup.MinNodeCount,
			maxSize:           *nodeGroup.MaxNodeCount,
			targetSize:        nodeGroup.CurrentSize,
			deletedNodes:      make(map[string]time.Time),
		}
		ixcp.AddNodeGroup(ng)
		newNodeGroupNames = append(newNodeGroupNames, name)
	}

	// Drop any node groups that should not be autoscaled either
	// because they were deleted or had their maximum node count unset.
	// Done by copying all node groups to a buffer, clearing the original
	// node groups and copying back only the ones that should still exist.
	ixcp.nodeGroupsLock.Lock()
	buffer := make([]*ixCloudNodeGroup, len(ixcp.nodeGroups))
	copy(buffer, ixcp.nodeGroups)

	ixcp.nodeGroups = nil

	for _, ng := range buffer {
		if _, ok := autoscalingNGs[ng.UUID]; ok {
			ixcp.nodeGroups = append(ixcp.nodeGroups, ng)
		} else {
			droppedNodeGroupNames = append(droppedNodeGroupNames, ng.id)
		}
	}

	ixcp.nodeGroupsLock.Unlock()

	// Log whatever actions were taken
	if len(newNodeGroupNames) == 0 && len(droppedNodeGroupNames) == 0 {
		klog.V(3).Info("No nodegroups added or removed")
		return nil
	}

	if len(newNodeGroupNames) > 0 {
		klog.V(2).Infof("Discovered %d new node groups for autoscaling: %s", len(newNodeGroupNames),
			strings.Join(newNodeGroupNames, ", "))
	}
	if len(droppedNodeGroupNames) > 0 {
		klog.V(2).Infof("Dropped %d node groups which should no longer be autoscaled: %s",
			len(droppedNodeGroupNames), strings.Join(droppedNodeGroupNames, ", "))
	}

	return nil
}

// BuildIxCloud is called by the autoscaler to build a ixCloud provider.
//
// The ixCloudManager is created here, and the initial node groups are created
// based on the static or auto discovery specs provided via the command line parameters.
func BuildIxCloud(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	var config io.ReadCloser

	// Should be loaded with --cloud-config /etc/kubernetes/kube_openstack_config from master node.
	if opts.CloudConfig != "" {
		var err error
		config, err = os.Open(opts.CloudConfig)
		if err != nil {
			klog.Fatalf("Couldn't open cloud provider configuration from %s: %#v", opts.CloudConfig, err)
		}
		defer config.Close()
	}

	// Check that one of static node group discovery or auto discovery are specified.
	// if !do.DiscoverySpecified() {
	// 	klog.Fatal("no node group discovery options specified")
	// }
	// if do.StaticDiscoverySpecified() && do.AutoDiscoverySpecified() {
	// 	klog.Fatal("can not use both static node group discovery and node group auto discovery")
	// }

	manager, err := createIxCloudManager(config, opts)
	if err != nil {
		klog.Fatalf("Failed to create ixcloud manager: %v", err)
	}

	provider, err := buildIxCloudProvider(manager, rl)
	if err != nil {
		klog.Fatalf("Failed to create ixcloud cloud provider: %v", err)
	}

	clusterUpdateLock := sync.Mutex{}
	provider.clusterUpdateLock = &clusterUpdateLock

	// Handle initial node group discovery.
	if do.StaticDiscoverySpecified() {
		for _, nodegroupSpec := range do.NodeGroupSpecs {
			// Parse a node group spec in the form min:max:name
			spec, err := dynamic.SpecFromString(nodegroupSpec, scaleToZeroSupported)
			if err != nil {
				klog.Fatalf("Could not parse node group spec %s: %v", nodegroupSpec, err)
			}

			ng := &ixCloudNodeGroup{
				ixCloudManager:    manager,
				id:                spec.Name,
				clusterUpdateLock: &clusterUpdateLock,
				minSize:           spec.MinSize,
				maxSize:           spec.MaxSize,
				targetSize:        1,
				deletedNodes:      make(map[string]time.Time),
			}

			// TODO: uniqueNameAndIDForNodeGroup 파라미터 수정
			name, uuid, err := ng.ixCloudManager.uniqueNameAndIDForNodeGroup(ng.id)
			if err != nil {
				klog.Fatalf("could not get unique name and UUID for node group %s: %v", spec.Name, err)
			}
			ng.id = name
			ng.UUID = uuid

			// Fetch the current size of this node group.
			ng.targetSize, err = ng.ixCloudManager.nodeGroupSize(ng.UUID)
			if err != nil {
				klog.Fatalf("Could not get current number of nodes in node group %s: %v", spec.Name, err)
			}

			provider.AddNodeGroup(ng)
		}
	} else if do.AutoDiscoverySpecified() {
		provider.usingAutoDiscovery = true
		cfgs, err := parseIxCloudAutoDiscoverySpecs(do)
		if err != nil {
			klog.Fatalf("Could not parse auto discovery specs: %v", err)
		}
		provider.autoDiscoveryConfigs = cfgs

		err = provider.refreshNodeGroups()
		if err != nil {
			klog.Fatalf("Initial node group discovery failed: %v", err)
		}
		provider.lastDiscoveryRefresh = time.Now()
	}

	return provider
}
