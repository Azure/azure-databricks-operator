module github.com/microsoft/azure-databricks-operator

go 1.12

require (
	github.com/go-logr/logr v0.1.0
	github.com/go-playground/errors v3.3.0+incompatible // indirect
	github.com/go-playground/log v6.3.0+incompatible
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/xinsnake/databricks-sdk-golang v0.0.0-20190710053313-b92a552450b1
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4 // indirect
	golang.org/x/net v0.0.0-20190628185345-da137c7871d7
	golang.org/x/sys v0.0.0-20190626221950-04f50cda93cb // indirect
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.2.0-beta.2
	sigs.k8s.io/controller-tools v0.2.0-beta.2 // indirect
)
