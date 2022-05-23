package iksclient

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

func ListNodeGroups(client *IksApiClient, clusterID string) ([]NodeGroup, error) {
	return []NodeGroup{}, nil
}

func GetNodeGroup(client *IksApiClient, nodeGroupID string) (NodeGroup, error) {
	resp, err := resty.New().R().
		SetHeader("Authorization", client.Token()).
		Get(client.ServiceURL("k8s", "node_groups", nodeGroupID))

	if err != nil {
		return NodeGroup{}, err
	}

	if resp.StatusCode() != http.StatusOK {
		return NodeGroup{}, fmt.Errorf("error : %v", string(resp.Body()))
	}

	response := NodeGroup{}
	err = json.Unmarshal(resp.Body(), &response)
	if err != nil {
		return NodeGroup{}, err
	}

	return response, nil
}

// ResizeOpts params
// TODO: options 우리 API에 맞게 수정
type ResizeOpts struct {
	NodeCount     *int     `json:"node_count" required:"true"`
	NodesToRemove []string `json:"nodes_to_remove,omitempty"`
	NodeGroup     string   `json:"nodegroup,omitempty"`
}

func Resize(iksApiClient *IksApiClient, nodeGroupID string, opts ResizeOpts) (NodeGroup, error) {
	return NodeGroup{}, nil
}
