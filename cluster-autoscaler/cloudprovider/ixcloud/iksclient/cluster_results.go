package iksclient

import (
	"time"
)

type Cluster struct {
	ID                      string      `json:"id"`
	AccountID               string      `json:"account_id"`
	Name                    string      `json:"name"`
	Description             string      `json:"description"`
	ZoneName                string      `json:"zone_name"`
	ProjectID               string      `json:"project_id"`
	NodeGroupIds            []string    `json:"node_group_ids"`
	NodeGroupCount          int         `json:"node_group_count"`
	WorkerCount             int         `json:"worker_count"`
	SubnetID                string      `json:"subnet_id"`
	Keypair                 string      `json:"keypair"`
	KubernetesVersion       string      `json:"kubernetes_version"`
	PodNetworkCidr          string      `json:"pod_network_cidr"`
	ServiceNetworkCidr      string      `json:"service_network_cidr"`
	CniType                 string      `json:"cni_type"`
	PrometheusEnabled       bool        `json:"prometheus_enabled"`
	PrometheusVolumeType    string      `json:"prometheus_volume_type"`
	PrometheusVolumeSize    int         `json:"prometheus_volume_size"`
	PrometheusURL           string      `json:"prometheus_url"`
	K8SDashboardEnabled     bool        `json:"k8s_dashboard_enabled"`
	K8SDashboardURL         string      `json:"k8s_dashboard_url"`
	K8SDashboardAccessToken string      `json:"k8s_dashboard_access_token"`
	ConnectorIP             string      `json:"connector_ip"`
	Status                  string      `json:"status"`
	CreatedAt               time.Time   `json:"created_at"`
	UpdatedAt               time.Time   `json:"updated_at"`
	DeletedAt               interface{} `json:"deleted_at"`
}
