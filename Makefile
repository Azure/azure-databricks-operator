
# Image URL to use all building/pushing image targets
IMG ?= controller:latest

all: test build

# Build manager binary
build: fmt vet
	go build -o bin/manager microsoft/azure-databricks-operator/cmd/manager

test:
	go test ./pkg/... ./cmd/... -coverprofile cover.out

# Run tests with code generation
test-gen: generate fmt vet manifests test
	go test ./pkg/... ./cmd/... -coverprofile cover.out

# Build manager binary with code generation
build-gen: generate fmt vet build

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run ./cmd/manager/main.go

# Install CRDs into a cluster
install:
	kubectl apply -f config/crds

install-gen: manifests install

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	kubectl apply -f config/crds
	kustomize build config | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests:
	go run ../../../pkg/mod/sigs.k8s.io/controller-tools@v0.1.10/cmd/controller-gen/main.go all

# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

# Generate code
generate:
ifndef GOPATH
	$(error GOPATH not defined, please define GOPATH. Run "go help gopath" to learn more about GOPATH)
endif
	go generate ./pkg/... ./cmd/...

# Build the docker image
docker-build: test
	docker build . -t ${IMG}
	@echo "updating kustomize image patch file for manager resource"
	sed -i'' -e 's@image: .*@image: '"${IMG}"'@' ./config/default/azure_databricks_operator_image_patch.yaml

# Push the docker image
docker-push:
	docker push ${IMG}
