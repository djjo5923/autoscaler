package iksclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"k8s.io/autoscaler/cluster-autoscaler/config"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud"
)

// Config is used to read and store information from the cloud configuration file
//
// Taken from kubernetes/pkg/cloudprovider/providers/openstack/openstack.go
// LoadBalancer, BlockStorage, Route, Metadata are not needed for the autoscaler,
// but are kept so that if a cloud-config file with those sections is provided
// then the parsing will not fail.
type Config struct {
	Global struct {
		ApiURL          string `gcfg:"auth-url"`
		UserID          string `gcfg:"user-id"`
		Password        string
		Region          string
		SecretName      string `gcfg:"secret-name"`
		SecretNamespace string `gcfg:"secret-namespace"`
	}
}

type IksApiClient struct {
	// Endpoint is the base URL of the service's API, acquired from a service catalog.
	// It MUST end with a /.
	Endpoint string

	// ResourceBase is the base URL shared by the resources within a service's API. It should include
	// the API version and, like Endpoint, MUST end with a / if set. If not set, the Endpoint is used
	// as-is, instead.
	ResourceBase string

	Context context.Context

	TokenID string

	HTTPClient http.Client

	// mut is a mutex for the client. It protects read and write access to client attributes such as getting
	// and setting the TokenID.
	mut *sync.RWMutex
}

func CreateIksApiClient(cfg *Config, opts config.AutoscalingOptions) (*IksApiClient, error) {
	if opts.ClusterName == "" {
		return nil, errors.New("the cluster-name parameter must be set")
	}

	// TODO: Endpoint
	iksApiClient := IksApiClient{
		Endpoint: cfg.Global.ApiURL,
	}

	return &iksApiClient, nil
}

func (client *IksApiClient) Get(url string, JSONResponse interface{}, opts *gophercloud.RequestOpts) (*http.Response, error) {
	if opts == nil {
		opts = new(gophercloud.RequestOpts)
	}
	client.initReqOpts(url, nil, JSONResponse, opts)
	return client.Request("GET", url, opts)
}

func (client *IksApiClient) Post(url string, JSONBody interface{}, JSONResponse interface{}, opts *gophercloud.RequestOpts) (*http.Response, error) {
	if opts == nil {
		opts = new(gophercloud.RequestOpts)
	}
	client.initReqOpts(url, JSONBody, JSONResponse, opts)
	return client.Request("POST", url, opts)
}

// ResourceBaseURL returns the base URL of any resources used by this service. It MUST end with a /.
func (client *IksApiClient) ResourceBaseURL() string {
	if client.ResourceBase != "" {
		return client.ResourceBase
	}
	return client.Endpoint
}

// ServiceURL constructs a URL for a resource belonging to this provider.
func (client *IksApiClient) ServiceURL(parts ...string) string {
	return client.ResourceBaseURL() + strings.Join(parts, "/")
}

func (client *IksApiClient) initReqOpts(url string, JSONBody interface{}, JSONResponse interface{}, opts *gophercloud.RequestOpts) {
	if v, ok := (JSONBody).(io.Reader); ok {
		opts.RawBody = v
	} else if JSONBody != nil {
		opts.JSONBody = JSONBody
	}

	if JSONResponse != nil {
		opts.JSONResponse = JSONResponse
	}

	if opts.MoreHeaders == nil {
		opts.MoreHeaders = make(map[string]string)
	}
}

func (client *IksApiClient) Request(method, url string, options *gophercloud.RequestOpts) (*http.Response, error) {
	var body io.Reader
	var contentType *string
	var applicationJSON = "application/json"

	// Derive the content body by either encoding an arbitrary object as JSON, or by taking a provided
	// io.ReadSeeker as-is. Default the content-type to application/json.
	if options.JSONBody != nil {
		if options.RawBody != nil {
			return nil, errors.New("please provide only one of JSONBody or RawBody to gophercloud.Request()")
		}

		rendered, err := json.Marshal(options.JSONBody)
		if err != nil {
			return nil, err
		}

		body = bytes.NewReader(rendered)
		contentType = &applicationJSON
	}

	if options.RawBody != nil {
		body = options.RawBody
	}

	// Construct the http.Request.
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if client.Context != nil {
		req = req.WithContext(client.Context)
	}
	// Populate the request headers. Apply options.MoreHeaders last, to give the caller the chance to
	// modify or omit any header.
	if contentType != nil {
		req.Header.Set("Content-Type", *contentType)
	}
	req.Header.Set("Accept", applicationJSON)

	// Set token
	req.Header.Set("Authorization", client.Token())

	if options.MoreHeaders != nil {
		for k, v := range options.MoreHeaders {
			if v != "" {
				req.Header.Set(k, v)
			} else {
				req.Header.Del(k)
			}
		}
	}

	// Set connection parameter to close the connection immediately when we've got the response
	req.Close = true

	// Issue the request.
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Allow default OkCodes if none explicitly set
	okc := options.OkCodes
	if okc == nil {
		okc = defaultOkCodes(method)
	}

	// Validate the HTTP response status.
	var ok bool
	for _, code := range okc {
		if resp.StatusCode == code {
			ok = true
			break
		}
	}

	if !ok {
		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		respErr := gophercloud.ErrUnexpectedResponseCode{
			URL:      url,
			Method:   method,
			Expected: options.OkCodes,
			Actual:   resp.StatusCode,
			Body:     body,
		}

		errType := options.ErrorContext
		switch resp.StatusCode {
		case http.StatusBadRequest:
			err = gophercloud.ErrDefault400{ErrUnexpectedResponseCode: respErr}
			if error400er, ok := errType.(gophercloud.Err400er); ok {
				err = error400er.Error400(respErr)
			}
		case http.StatusUnauthorized:
			//TODO implement reauth function
			err = gophercloud.ErrDefault401{ErrUnexpectedResponseCode: respErr}
			if error401er, ok := errType.(gophercloud.Err401er); ok {
				err = error401er.Error401(respErr)
			}
		case http.StatusForbidden:
			err = gophercloud.ErrDefault403{ErrUnexpectedResponseCode: respErr}
			if error403er, ok := errType.(gophercloud.Err403er); ok {
				err = error403er.Error403(respErr)
			}
		case http.StatusNotFound:
			err = gophercloud.ErrDefault404{ErrUnexpectedResponseCode: respErr}
			if error404er, ok := errType.(gophercloud.Err404er); ok {
				err = error404er.Error404(respErr)
			}
		case http.StatusMethodNotAllowed:
			err = gophercloud.ErrDefault405{ErrUnexpectedResponseCode: respErr}
			if error405er, ok := errType.(gophercloud.Err405er); ok {
				err = error405er.Error405(respErr)
			}
		case http.StatusRequestTimeout:
			err = gophercloud.ErrDefault408{ErrUnexpectedResponseCode: respErr}
			if error408er, ok := errType.(gophercloud.Err408er); ok {
				err = error408er.Error408(respErr)
			}
		case http.StatusConflict:
			err = gophercloud.ErrDefault409{ErrUnexpectedResponseCode: respErr}
			if error409er, ok := errType.(gophercloud.Err409er); ok {
				err = error409er.Error409(respErr)
			}
		case 429:
			err = gophercloud.ErrDefault429{ErrUnexpectedResponseCode: respErr}
			if error429er, ok := errType.(gophercloud.Err429er); ok {
				err = error429er.Error429(respErr)
			}
		case http.StatusInternalServerError:
			err = gophercloud.ErrDefault500{ErrUnexpectedResponseCode: respErr}
			if error500er, ok := errType.(gophercloud.Err500er); ok {
				err = error500er.Error500(respErr)
			}
		case http.StatusServiceUnavailable:
			err = gophercloud.ErrDefault503{ErrUnexpectedResponseCode: respErr}
			if error503er, ok := errType.(gophercloud.Err503er); ok {
				err = error503er.Error503(respErr)
			}
		}

		if err == nil {
			err = respErr
		}

		return resp, err
	}

	// Parse the response body as JSON, if requested to do so.
	if options.JSONResponse != nil {
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(options.JSONResponse); err != nil {
			return nil, err
		}
	}

	return resp, nil
}

func (client *IksApiClient) Token() string {
	if client.mut != nil {
		client.mut.RLock()
		defer client.mut.RUnlock()
	}
	return client.TokenID
}

func defaultOkCodes(method string) []int {
	switch {
	case method == "GET":
		return []int{200}
	case method == "POST":
		return []int{201, 202}
	case method == "PUT":
		return []int{201, 202}
	case method == "PATCH":
		return []int{200, 202, 204}
	case method == "DELETE":
		return []int{202, 204}
	}

	return []int{}
}
