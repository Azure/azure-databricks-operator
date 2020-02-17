# Image URL to use all building/pushing image targets
IMG ?= controller:latest

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Prometheus helm installation name
PROMETHEUS_NAME ?= "prom-azure-databricks-operator"

# Default kind cluster name
KIND_CLUSTER_NAME ?= "azure-databricks-operator"

# Default namespace for the installation
OPERATOR_NAMESPACE ?= "azure-databricks-operator-system"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

timestamp := $(shell /bin/date "+%Y%m%d-%H%M%S")
all: manager

# Run tests
test: generate fmt vet manifests lint
	rm -rf cover.* cover
	mkdir -p cover

	TEST_USE_EXISTING_CLUSTER=false go test ./api/... ./controllers/... -coverprofile cover.out.tmp
	cat cover.out.tmp | grep -v "_generated.deepcopy.go" > cover.out
	gocov convert cover.out > cover.json
	gocov-xml < cover.json > cover.xml
	gocov-html < cover.json > cover/index.html

	rm -f cover.out cover.out.tmp cover.json

# Run tests with existing cluster
test-existing: generate fmt vet manifests lint
	rm -rf cover.* cover
	mkdir -p cover

	TEST_USE_EXISTING_CLUSTER=true go test ./api/... ./controllers/... -coverprofile cover.out.tmp
	cat cover.out.tmp | grep -v "_generated.deepcopy.go" > cover.out
	gocov convert cover.out > cover.json
	gocov-xml < cover.json > cover.xml
	gocov-html < cover.json > cover/index.html

	rm -f cover.out cover.out.tmp cover.json

# Build manager binary
manager: generate fmt lint vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt lint vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
ifeq (0, $(shell kubectl get namespaces 2>&1 | grep ${OPERATOR_NAMESPACE} | wc -l))
	@echo "creating ${OPERATOR_NAMESPACE} namespace"
	kubectl create namespace ${OPERATOR_NAMESPACE}
	make create-dbrickssettings-secret
else
	@echo "${OPERATOR_NAMESPACE} namespace exists"
ifeq (0, $(shell kubectl get secrets --namespace ${OPERATOR_NAMESPACE} | grep dbrickssettings | wc -l))
	@echo "creating dbrickssettings secret"
	create-dbrickssettings-secret
else
	@echo "dbrickssettings secret exists"
endif
endif

	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -
	kustomize build config/default > operatorsetup.yaml
	




create-dbrickssettings-secret:
	kubectl --namespace ${OPERATOR_NAMESPACE} \
		create secret generic dbrickssettings \
		--from-literal=DatabricksHost="${DATABRICKS_HOST}" \
		--from-literal=DatabricksToken="${DATABRICKS_TOKEN}"

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	find . -name '*.go' | grep -v vendor | xargs gofmt -s -w

# Run go vet against code
vet:
	go vet ./...
	
# Run linting
lint:
	GO111MODULE=on golangci-lint run
	
# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

# Build the docker image
docker-build: 
	docker build . -t ${IMG} ${ARGS}
	@echo "updating kustomize image patch file for manager resource"
	cd config/manager && kustomize edit set image controller=${IMG}
# Push the docker image
docker-push:
	docker push ${IMG}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.4 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

create-kindcluster:
ifeq (0, $(shell kind get clusters | grep ${KIND_CLUSTER_NAME} | wc -l))
	@echo "no kind cluster"
else
	@echo "kind cluster is running, deleting the current cluster"
	kind delete cluster --name ${KIND_CLUSTER_NAME}
endif
	@echo "creating kind cluster"
	kind create cluster --name ${KIND_CLUSTER_NAME}

set-kindcluster: install-kind
	make create-kindcluster
	kubectl cluster-info
	@echo "deploying controller to cluster"
	make deploy-kindcluster
	make install
	make install-prometheus

# Deploy controller
deploy-kindcluster:
	#create image and load it into cluster
	$(eval newimage := "docker.io/controllertest:$(timestamp)")
	IMG=$(newimage) make docker-build
	kind load docker-image $(newimage) --loglevel "debug" --name ${KIND_CLUSTER_NAME}

	#deploy operator
	IMG=$(newimage) make deploy
	#change image name back to orignal image name
	cd config/manager && kustomize edit set image controller="IMAGE_URL"

install-kind:
ifeq (,$(shell which kind))
	@echo "installing kind"
	curl -Lo ./kind "https://github.com/kubernetes-sigs/kind/releases/download/v0.7.0/kind-$(shell uname)-amd64" 
	chmod +x ./kind 
	mv ./kind /usr/local/bin/kind
else
	@echo "kind has been installed"
endif
install-kubebuilder:
ifeq (,$(shell which kubebuilder))
	@echo "installing kubebuilder"
	# download kubebuilder and extract it to tmp
	curl -sL https://go.kubebuilder.io/dl/2.2.0/$(shell go env GOOS)/$(shell go env GOARCH) | tar -xz -C /tmp/
	# move to a long-term location and put it on your path
	# (you'll need to set the KUBEBUILDER_ASSETS env var if you put it somewhere else)
	mv /tmp/kubebuilder_2.2.0_$(shell go env GOOS)_$(shell go env GOARCH) /usr/local/kubebuilder
	export PATH=$PATH:/usr/local/kubebuilder/bin
else
	@echo "kubebuilder has been installed"
endif

install-kustomize:
ifeq (,$(shell which kustomize))
	@echo "installing kustomize"
	# download kustomize
	curl -sL https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv3.4.0/kustomize_v3.4.0_$(shell go env GOOS)_$(shell go env GOARCH).tar.gz | tar -xz -C /tmp/
	mv /tmp/kustomize /usr/local/kubebuilder/bin/kustomize
	# set permission
	chmod a+x /usr/local/kubebuilder/bin/kustomize
	$(shell which kustomize)
else
	@echo "kustomize has been installed"
endif

install-prometheus:
	@echo "installing prometheus"
	# install prometheus (and set to monitor all namespaces in our kind cluster)
	helm install ${PROMETHEUS_NAME} stable/prometheus-operator --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false 
	@echo "prometheus has been installed"

install-test-dependency:
	go get -u github.com/jstemmer/go-junit-report \
	&& go get github.com/axw/gocov/gocov \
	&& go get github.com/AlekSi/gocov-xml \
	&& go get github.com/onsi/ginkgo/ginkgo \
	&& go get golang.org/x/tools/cmd/cover \
	&& go get -u github.com/matm/gocov-html
