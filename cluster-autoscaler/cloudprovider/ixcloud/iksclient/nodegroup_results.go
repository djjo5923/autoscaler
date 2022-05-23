package iksclient

import (
	"time"
)

// NodeGroup is the API representation of a IKS node group.
//TODO NodeGroup에 autoscale, min_node_count, max_node_count 추가
type NodeGroup struct {
	ID           string    `json:"id"`
	AccountID    string    `json:"account_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	ClusterID    string    `json:"cluster_id"`
	ZoneName     string    `json:"zone_name"`
	ProjectID    string    `json:"project_id"`
	ImageID      string    `json:"image_id"`
	FlavorID     string    `json:"flavor_id"`
	CurrentSize  int       `json:"current_size"`
	GpuEnabled   bool      `json:"gpu_enabled"`
	FipEnabled   bool      `json:"fip_enabled"`
	Nodes        []Node    `json:"nodes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	DeletedAt    time.Time `json:"deleted_at"`
	MinNodeCount int       `json:"min_node_count"`
	MaxNodeCount *int      `json:"max_node_count"`
}

//Node is the API representation of a IKS node.
type Node struct {
	ID        string    `json:"id"`
	PrivateIP string    `json:"private_ip"`
	PublicIP  string    `json:"public_ip"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`
}
