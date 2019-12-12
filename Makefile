
# Image URL to use all building/pushing image targets
IMG ?= controller:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

# Run tests
test: generate fmt lint vet manifests
	rm -rf cover.* cover
	mkdir -p cover

	TEST_USE_EXISTING_CLUSTER=false go test ./api/... ./controllers/... -coverprofile cover.out.tmp
	cat cover.out.tmp | grep -v "_generated.deepcopy.go" > cover.out
	gocov convert cover.out > cover.json
	gocov-xml < cover.json > cover.xml
	gocov-html < cover.json > cover/index.html

	rm -f cover.out cover.out.tmp cover.json

# Run tests with existing cluster
test-existing: generate fmt lint vet manifests
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
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

deploy-controller:
	kubectl create namespace azure-databricks-operator-system
	kubectl --namespace azure-databricks-operator-system \
		create secret generic dbrickssettings \
		--from-literal=DatabricksHost="${DATABRICKS_HOST}" \
		--from-literal=DatabricksToken="${DATABRICKS_TOKEN}"

	#create image and load it into cluster
	IMG="docker.io/controllertest:1" make docker-build
	kind load docker-image docker.io/controllertest:1 --loglevel "trace"
	make install
	make deploy
	sed -i'' -e 's@image: .*@image: '"IMAGE_URL"'@' ./config/default/manager_image_patch.yaml

timestamp := $(shell /bin/date "+%Y%m%d-%H%M%S")

update-deployed-controller:
	IMG="docker.io/controllertest:$(timestamp)" make ARGS="${ARGS}" docker-build
	kind load docker-image docker.io/controllertest:$(timestamp) --loglevel "trace"
	make install
	make deploy
	sed -i'' -e 's@image: .*@image: '"IMAGE_URL"'@' ./config/default/manager_image_patch.yaml

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
docker-build: test
	docker build . -t ${IMG} ${ARGS}
	@echo "updating kustomize image patch file for manager resource"
	sed -i'' -e 's@image: .*@image: '"${IMG}"'@' ./config/default/manager_image_patch.yaml

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
ifeq (,$(shell kind get clusters))
	@echo "no kind cluster"
else
	@echo "kind cluster is running, deleting the current cluster"
	kind delete cluster 
endif
	@echo "creating kind cluster"
	kind create cluster

set-kindcluster: install-kind
ifeq (${shell kind get kubeconfig-path --name="kind"},${KUBECONFIG})
	@echo "kubeconfig-path points to kind path"
else
	@echo "please run below command in your shell and then re-run make set-kindcluster"
	@echo  "\e[31mexport KUBECONFIG=$(shell kind get kubeconfig-path --name="kind")\e[0m"
	@exit 111
endif
	make create-kindcluster
	
	@echo "getting value of KUBECONFIG"
	@echo ${KUBECONFIG}
	@echo "getting value of kind kubeconfig-path"
	
	kubectl cluster-info

	@echo "deploying controller to cluster"
	make deploy-controller

install-kind:
ifeq (,$(shell which kind))
	@echo "installing kind"
	GO111MODULE="on" go get sigs.k8s.io/kind@v0.4.0
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

install-test-dependency:
	go get -u github.com/jstemmer/go-junit-report \
	&& go get github.com/axw/gocov/gocov \
	&& go get github.com/AlekSi/gocov-xml \
	&& go get github.com/onsi/ginkgo/ginkgo \
	&& go get golang.org/x/tools/cmd/cover \
	&& go get -u github.com/matm/gocov-html
