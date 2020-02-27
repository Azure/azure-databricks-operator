# Timestamp for image tags
timestamp := $(shell /bin/date "+%Y%m%d-%H%M%S")

# Image URL to use all building/pushing image targets
IMG ?= controller:${timestamp}

# MockAPI image URL to use all building/pushing image targets
MOCKAPI_IMG ?= mockapi:${timestamp}

# MockAPI image URL to use all building/pushing image targets
LOCUST_IMG ?= locust:${timestamp}

# Default namespace for the installation
LOCUST_FILE ?= "behaviours/scenario1_run_submit_delete.py"

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

create-namespace:
	@echo "$(shell tput setaf 10)$(shell tput bold)Creating ${OPERATOR_NAMESPACE} namespace if doesn't exist $(shell tput sgr0)"
	-kubectl create namespace ${OPERATOR_NAMESPACE}

	# Verify namespace was successfully created
	kubectl get namespace azure-databricks-operator-system


create-dbrickssettings-secret:
	@echo "$(shell tput setaf 10)$(shell tput bold)Creating dbrickssettings secret if doesn't exist $(shell tput sgr0)"
	-kubectl --namespace ${OPERATOR_NAMESPACE} \
		create secret generic dbrickssettings \
		--from-literal=DatabricksHost="${DATABRICKS_HOST}" \
		--from-literal=DatabricksToken="${DATABRICKS_TOKEN}"

	# Verify secret was created
	kubectl get secret dbrickssettings -n azure-databricks-operator-system

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: create-namespace create-dbrickssettings-secret manifests
	@echo "$(shell tput setaf 10)$(shell tput bold)Deploying the operator $(shell tput sgr0)" 
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -
	kustomize build config/default > operatorsetup.yaml

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	find . -name '*.go' | grep -v -E 'vendor|.gocache' | xargs gofmt -s -w

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
	@echo "$(shell tput setaf 10)$(shell tput bold)Building docker image for the operator $(shell tput sgr0)" 
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
	@echo "$(shell tput setaf 1)$(shell tput bold)Deleting kind cluster if running $(shell tput sgr0)"
	-kind delete cluster --name ${KIND_CLUSTER_NAME}

	@echo "$(shell tput setaf 10)$(shell tput bold)Creating kind cluster $(shell tput sgr0)"
	kind create cluster --name ${KIND_CLUSTER_NAME} --config ./kind-cluster.yaml

set-kindcluster: install-kind
	make create-kindcluster
	kubectl cluster-info
	make deploy-kindcluster
	make install
	make install-prometheus

deploy-image-to-kind:
	@echo "$(shell tput setaf 10)$(shell tput bold)Load operator image into kind $(shell tput sgr0)" 
	kind load docker-image $(IMG) --loglevel "debug" --name ${KIND_CLUSTER_NAME}

# Deploy controller
deploy-kindcluster: docker-build deploy-image-to-kind deploy

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
	@echo "$(shell tput setaf 10)$(shell tput bold)Installing Prometheus in cluster $(shell tput sgr0)" 
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

build-mock-api:
	go build -o bin/mock-databricks-api ./mockapi

run-mock-api:
	go run ./mockapi

test-mock-api: lint
	go test ./mockapi/...

docker-build-mock-api: 
	@echo "$(shell tput setaf 10)$(shell tput bold)Building mockapi docker image $(shell tput sgr0)" 

	docker build -t ${MOCKAPI_IMG} -f mockapi/Dockerfile .

docker-push-mock-api: docker-build
	docker push ${IMG}

apply-manifests-mock-api:
	@echo "$(shell tput setaf 10)$(shell tput bold)Deploying mockapi $(shell tput sgr0)" 

	cat ./mockapi/manifests/deployment.yaml | sed "s|mockapi:latest|${MOCKAPI_IMG}|" | kubectl apply -f -
	kubectl apply -f ./mockapi/manifests/service.yaml

kind-load-image-mock-api: docker-build-mock-api 
	@echo "$(shell tput setaf 10)$(shell tput bold)Loading mockapi image into kind $(shell tput sgr0)" 

	kind load docker-image ${MOCKAPI_IMG} --name ${KIND_CLUSTER_NAME} -v 1

deploy-mock-api:kind-load-image-mock-api apply-manifests-mock-api

kind-deploy-mock-api: create-kindcluster install-prometheus deploy-mock-api

# Args passed to locust must be in CSV format as passed in "command" section of yaml doc
LOCUST_ARGS?=,'--no-web', '-c', '25', '-r', '0.08'
deploy-locust:
	@echo "$(shell tput setaf 10)$(shell tput bold)Deploying Locust $(shell tput sgr0)" 

	# Delete locust pod if already running
	-kubectl delete job locust-loadtest

	docker build -t ${LOCUST_IMG} -f locust/Dockerfile .
	kind load docker-image ${LOCUST_IMG} --name ${KIND_CLUSTER_NAME} -v 1

	# do some magic
	cat ./locust/manifests/deployment.yaml | sed "s|locust:latest|${LOCUST_IMG}|" | sed "s|behaviours/scenario1_run_submit_delete.py'|${LOCUST_FILE}' ${LOCUST_ARGS}|" | kubectl apply -f -

kind-deploy-locust: create-kindcluster install-prometheus deploy-locust	

format-locust:
	black .

test-locust: 
	pip install -e ./locust -q
	pytest

port-forward:
	./portforwards.sh

create-db-mock-secret: create-namespace
	@echo "$(shell tput setaf 10)$(shell tput bold)Creating mock api databricks secret $(shell tput sgr0)" 

	kubectl --namespace ${OPERATOR_NAMESPACE} \
			create secret generic dbrickssettings \
			--from-literal=DatabricksHost="http://databricks-mock-api.databricks-mock-api:8080" \
			--from-literal=DatabricksToken="dummy"

deploy-cluster-for-load-testing: create-kindcluster install-prometheus create-db-mock-secret deploy-kindcluster deploy-mock-api 
	@echo "$(shell tput setaf 10)$(shell tput bold)Deploying grafana dashboards $(shell tput sgr0)" 

	# deploy service monitor
	cat ./config/prometheus/monitor.yaml | sed "s/namespace: system/namespace: ${OPERATOR_NAMESPACE}/g" | kubectl apply -f -

	# deploy graphs
	kubectl apply -f ./config/prometheus/grafana-dashboard-configmap.yaml
	kubectl apply -f ./config/prometheus/grafana-dashboard-load-test-configmap.yaml
	kubectl apply -f ./config/prometheus/grafana-dashboard-mockapi-configmap.yaml

run-load-testing: deploy-cluster-for-load-testing deploy-locust port-forward

	##
	# python3 ./test.py

	sleep 45
	curl localhost:9090 > promstats-locust.txt


	go get -u github.com/ryotarai/prometheus-query
	
	prometheus-query -server http://localhost:9091 -query locust_user_count -start "now" -end "now" | jq .[0].values[0].value	
	
	# LOCUST_ARGS?="'--noweb', '-c', '25', '-r', '0.03'"

	# python3 ./test.py ./promstats-lucust.txt
	
	# Check stats
	# while "curl localhost:9090" prom metics "locust_user_count" < 25... wait

	# get prommetrics and extract stats
	# check against thresholds 

	# pass or fail!

	# prometheus-query -server http://localhost:9091 -query locust_user_count -start "now" -end "now" | jq .[0].values[0].value
	# # LOCUST_ARGS?=,'--no-web', '-c', '25', '-r', '0.08


