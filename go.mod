module github.com/ns1/cert-manager-webhook-ns1

go 1.16

require (
	github.com/jetstack/cert-manager v0.13.0
	gopkg.in/ns1/ns1-go.v2 v2.6.2
	k8s.io/apiextensions-apiserver v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v0.17.0
)

replace github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.4
