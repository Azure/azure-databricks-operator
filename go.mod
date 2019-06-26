module github.com/microsoft/azure-databricks-operator

go 1.12

require (
	github.com/go-logr/logr v0.1.0
	github.com/go-playground/errors v3.3.0+incompatible // indirect
	github.com/go-playground/log v6.3.0+incompatible
	github.com/onsi/ginkgo v1.6.0
	github.com/onsi/gomega v1.4.2
	github.com/xinsnake/databricks-sdk-golang v0.0.0-20190625012826-3a272d083f0b
	golang.org/x/net v0.0.0-20180906233101-161cd47e91fd
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.2.0-beta.2
)
