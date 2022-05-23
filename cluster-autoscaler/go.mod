module k8s.io/autoscaler/cluster-autoscaler

go 1.16

require (
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/go-resty/resty/v2 v2.7.0
	github.com/golang/mock v1.6.0
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/onsi/gomega v1.19.0 // indirect
	github.com/prometheus/client_golang v1.12.1
	github.com/satori/go.uuid v1.2.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.1
	google.golang.org/grpc v1.42.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/gcfg.v1 v1.2.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.25.0-alpha.0
	k8s.io/apimachinery v0.25.0-alpha.0
	k8s.io/apiserver v0.25.0-alpha.0
	k8s.io/client-go v0.25.0-alpha.0
	k8s.io/cloud-provider v0.25.0-alpha.0
	k8s.io/component-base v0.25.0-alpha.0
	k8s.io/component-helpers v0.25.0-alpha.0
	k8s.io/klog/v2 v2.60.1
	k8s.io/kubelet v0.24.0
	k8s.io/kubernetes v1.25.0-alpha.0
	k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace github.com/aws/aws-sdk-go/service/eks => github.com/aws/aws-sdk-go/service/eks v1.38.49

replace github.com/digitalocean/godo => github.com/digitalocean/godo v1.27.0

replace github.com/rancher/go-rancher => github.com/rancher/go-rancher v0.1.0

replace k8s.io/api => k8s.io/api v0.25.0-alpha.0

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.25.0-alpha.0

replace k8s.io/apimachinery => k8s.io/apimachinery v0.25.0-alpha.0

replace k8s.io/apiserver => k8s.io/apiserver v0.25.0-alpha.0

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.25.0-alpha.0

replace k8s.io/client-go => k8s.io/client-go v0.25.0-alpha.0

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.25.0-alpha.0

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.25.0-alpha.0

replace k8s.io/code-generator => k8s.io/code-generator v0.25.0-alpha.0

replace k8s.io/component-base => k8s.io/component-base v0.25.0-alpha.0

replace k8s.io/component-helpers => k8s.io/component-helpers v0.25.0-alpha.0

replace k8s.io/controller-manager => k8s.io/controller-manager v0.25.0-alpha.0

replace k8s.io/cri-api => k8s.io/cri-api v0.25.0-alpha.0

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.25.0-alpha.0

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.25.0-alpha.0

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.25.0-alpha.0

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.25.0-alpha.0

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.25.0-alpha.0

replace k8s.io/kubectl => k8s.io/kubectl v0.25.0-alpha.0

replace k8s.io/kubelet => k8s.io/kubelet v0.25.0-alpha.0

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.25.0-alpha.0

replace k8s.io/metrics => k8s.io/metrics v0.25.0-alpha.0

replace k8s.io/mount-utils => k8s.io/mount-utils v0.25.0-alpha.0

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.25.0-alpha.0

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.25.0-alpha.0

replace k8s.io/sample-controller => k8s.io/sample-controller v0.25.0-alpha.0

replace k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.25.0-alpha.0
