module github.com/kube-queue/et-operator-extension

require (
	github.com/AliyunContainerService/et-operator v0.0.0-20210106061249-14999850bf84
	github.com/kube-queue/api v0.0.0-20210623033849-bffe1acb5aa9
	github.com/kube-queue/tf-operator-extension v0.0.0-20211115073552-deb5090c14d8 // indirect
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	go.uber.org/zap v1.9.1 // indirect
	golang.org/x/crypto v0.0.0-20200220183623-bac4c82f6975 // indirect
	k8s.io/api v0.21.1
	k8s.io/apiextensions-apiserver v0.0.0 // indirect
	k8s.io/apimachinery v0.21.1
	k8s.io/cli-runtime v0.16.8 // indirect
	k8s.io/client-go v0.21.1
	k8s.io/code-generator v0.21.2
	k8s.io/klog v1.0.0 // indirect
	k8s.io/klog/v2 v2.8.0
	k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
	k8s.io/kubernetes v1.16.8 // indirect
	k8s.io/sample-controller v0.21.1
	k8s.io/utils v0.0.0-20190801114015-581e00157fb1 // indirect
	sigs.k8s.io/controller-runtime v0.4.0
)

replace (
	k8s.io/api => k8s.io/api v0.16.8
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.16.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.8
	k8s.io/apiserver => k8s.io/apiserver v0.16.8
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.16.8
	k8s.io/client-go => k8s.io/client-go v0.16.8
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.16.8
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.16.8
	k8s.io/code-generator => k8s.io/code-generator v0.18.5
	k8s.io/component-base => k8s.io/component-base v0.16.8
	k8s.io/cri-api => k8s.io/cri-api v0.16.8
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.16.8
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.16.8
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.16.8
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.16.8
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.16.8
	k8s.io/kubectl => k8s.io/kubectl v0.16.8
	k8s.io/kubelet => k8s.io/kubelet v0.16.8
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.16.8
	k8s.io/metrics => k8s.io/metrics v0.16.8
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.16.8
)

go 1.16
