package iksclient

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

func GetCluster(client *IksApiClient, clusterId string) (Cluster, error) {
	resp, err := resty.New().R().
		SetHeader("Authorization", client.Token()).
		Get(client.ServiceURL("k8s", "clusters", clusterId))

	if err != nil {
		return Cluster{}, err
	}

	if resp.StatusCode() != http.StatusOK {
		return Cluster{}, fmt.Errorf("error : %v", string(resp.Body()))
	}

	response := Cluster{}
	err = json.Unmarshal(resp.Body(), &response)
	if err != nil {
		return Cluster{}, err
	}

	return response, nil
}
