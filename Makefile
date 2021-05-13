# Copyright Contributors to the Open Cluster Management project

# Allow operator-sdk version/binary to be used to be specified externally via
# an environment variable.
#
# Default to currrent dev approach of using a version specific alias or
# symbolic link called "osdk".

OPERATOR_SDK ?= osdk

# Current Operator version
VERSION ?= 0.0.1

# Options for 'bundle-build'
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# IMAGE_TAG_BASE defines the docker.io namespace and part of the image name for remote images.
# This variable is used to construct full image tags for bundle and catalog images.
#
# For example, running 'make bundle-build bundle-push catalog-build catalog-push' will build and push both
# open-cluster-management.io/discovery-bundle:$VERSION and open-cluster-management.io/discovery-catalog:$VERSION.
IMAGE_TAG_BASE ?= open-cluster-management/discovery

# BUNDLE_IMG defines the image:tag used for the bundle.
# You can use it as an arg. (E.g make bundle-build BUNDLE_IMG=<some-registry>/<project-name-bundle>:<tag>)
BUNDLE_IMG ?= $(IMAGE_TAG_BASE)-bundle:v$(VERSION)

# Image URL to use all building/pushing image targets
REGISTRY ?= quay.io/rhibmcollab
IMG ?= discovery-operator
URL ?= $(REGISTRY)/$(IMG):$(VERSION)
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:crdVersions=v1"

# Namespace to deploy resources into
NAMESPACE ?= open-cluster-management

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

-include testserver/Makefile

all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=discovery-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet `go list ./... | grep -v test`

test: manifests generate fmt vet ## Run tests.
	go test `go list ./... | grep -v e2e` -coverprofile cover.out

integration-tests: install deploy server/deploy ## Run integration tests
	kubectl apply -f testserver/build/clusters.open-cluster-management.io_managedclusters.yaml
	kubectl wait --for=condition=available --timeout=60s deployment/discovery-operator -n $(NAMESPACE)
	kubectl wait --for=condition=available --timeout=60s deployment/mock-ocm-server -n $(NAMESPACE)
	go test -v ./test/e2e -coverprofile cover.out -args -ginkgo.v -ginkgo.trace -namespace $(NAMESPACE)

ENCRYPTED = $(shell echo "ocmAPIToken: ${OCM_API_TOKEN}" | base64)
secret: ## Generate secret for OCM access
	cat config/samples/ocm-api-secret.yaml | sed -e "s/ENCRYPTED_TOKEN/$(ENCRYPTED)/g" | kubectl apply -f - || true

samples: ## Create custom resources
	$(KUSTOMIZE) build config/samples | kubectl apply -f -

logs: ## Print operator logs
	@kubectl logs -f $(shell kubectl get pod -l app=discovery-operator -o jsonpath="{.items[0].metadata.name}")

annotate: ## Annotate discoveryconfig to target mock server
	kubectl annotate discoveryconfig discovery ocmBaseURL=http://mock-ocm-server.$(NAMESPACE).svc.cluster.local:3000 --overwrite

unannotate: ## Remove mock server annotation
	kubectl annotate discoveryconfig discovery ocmBaseURL-

set-copyright:
	@bash ./cicd-scripts/set-copyright.sh

verify: test integration-tests manifests

##@ Build

build: generate fmt vet ## Build manager binary.
	go build -o bin/manager main.go

run: manifests generate fmt vet ## Run a controller from your host.
	go run ./main.go

docker-build: ## Build docker image with the manager.
	docker build -t "${URL}" .

docker-push: ## Push docker image with the manager.
	docker push "${URL}"

##@ Deployment

install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller="${URL}"
	cd config/default && $(KUSTOMIZE) edit set namespace $(NAMESPACE)
	$(KUSTOMIZE) build config/default | kubectl apply -f -
	# Reset values
	cd config/manager && $(KUSTOMIZE) edit set image controller="discovery-operator:latest"
	cd config/default && $(KUSTOMIZE) edit set namespace open-cluster-management

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl delete -f -

controller-gen: ## Download controller-gen locally if necessary.
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

kustomize: ## Download kustomize locally if necessary.
ifeq (, $(shell which kustomize))
	@{ \
	set -e ;\
	KUSTOMIZE_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$KUSTOMIZE_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/kustomize/kustomize/v3@v3.8.7 ;\
	rm -rf $$KUSTOMIZE_GEN_TMP_DIR ;\
	}
KUSTOMIZE=$(GOBIN)/kustomize
else
KUSTOMIZE=$(shell which kustomize)
endif

.PHONY: bundle
bundle: manifests kustomize ## Generate bundle manifests and metadata, then validate generated files.
	$(OPERATOR_SDK) generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	$(OPERATOR_SDK) bundle validate ./bundle
	cd config/manager && $(KUSTOMIZE) edit set image controller="discovery-operator:latest"

.PHONY: bundle-build
bundle-build: ## Build the bundle image.
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

.PHONY: bundle-push
bundle-push: ## Push the bundle image.
	$(MAKE) docker-push IMG=$(BUNDLE_IMG)

.PHONY: opm
OPM = ./bin/opm
opm: ## Download opm locally if necessary.
ifeq (,$(wildcard $(OPM)))
ifeq (,$(shell which opm 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.15.1/$${OS}-$${ARCH}-opm ;\
	chmod +x $(OPM) ;\
	}
else
OPM = $(shell which opm)
endif
endif

# A comma-separated list of bundle images (e.g. make catalog-build BUNDLE_IMGS=example.com/operator-bundle:v0.1.0,example.com/operator-bundle:v0.2.0).
# These images MUST exist in a registry and be pull-able.
BUNDLE_IMGS ?= $(BUNDLE_IMG)

# The image tag given to the resulting catalog image (e.g. make catalog-build CATALOG_IMG=example.com/operator-catalog:v0.2.0).
CATALOG_IMG ?= $(IMAGE_TAG_BASE)-catalog:v$(VERSION)

# Set CATALOG_BASE_IMG to an existing catalog image tag to add $BUNDLE_IMGS to that image.
ifneq ($(origin CATALOG_BASE_IMG), undefined)
FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG)
endif

# Build a catalog image by adding bundle images to an empty catalog using the operator package manager tool, 'opm'.
# This recipe invokes 'opm' in 'semver' bundle add mode. For more information on add modes, see:
# https://github.com/operator-framework/community-operators/blob/7f1438c/docs/packaging-operator.md#updating-your-existing-operator
.PHONY: catalog-build
catalog-build: opm ## Build a catalog image.
	$(OPM) index add --container-tool docker --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS) $(FROM_INDEX_OPT)

# Push the catalog image.
.PHONY: catalog-push
catalog-push: ## Push a catalog image.
	$(MAKE) docker-push IMG=$(CATALOG_IMG)

############################################################
# e2e test section
############################################################
.PHONY: kind-bootstrap-cluster
# Full setup of KinD cluster
kind-bootstrap-cluster: kind-create-cluster kind-load-image kind-load-testserver-image kind-deploy-controller kind-deploy-testserver

# Create deployment and configure it to never download image
kind-deploy-controller:
	@echo Installing discovery controller
	kubectl create namespace $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -
	
	cd config/default && $(KUSTOMIZE) edit set namespace $(NAMESPACE)
	$(KUSTOMIZE) build config/default | kubectl apply -f -
	cd config/default && $(KUSTOMIZE) edit set namespace open-cluster-management
	
	@echo "Patch deployment image"
	kubectl patch deployment discovery-operator -n $(NAMESPACE) -p "{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"discovery-operator\",\"imagePullPolicy\":\"Never\"}]}}}}"
	kubectl patch deployment discovery-operator -n $(NAMESPACE) -p "{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"discovery-operator\",\"image\":\"$(URL)\"}]}}}}"
	kubectl rollout status -n $(NAMESPACE) deployment discovery-operator --timeout=180s

kind-load-image:
	@echo Pushing image to KinD cluster
	kind load docker-image $(URL) --name test-discovery

kind-create-cluster:
	@echo "creating cluster"
	kind create cluster --name test-discovery
	# kind get kubeconfig --name test-discovery > $(PWD)/kubeconfig_managed
	kubectl cluster-info --context kind-test-discovery

kind-delete-cluster:
	kind delete cluster --name test-discovery

kind-e2e-tests:
	kubectl apply -f testserver/build/clusters.open-cluster-management.io_managedclusters.yaml
	go test -v ./test/e2e -coverprofile cover.out -args -ginkgo.v -ginkgo.trace -namespace $(NAMESPACE)

test-local:
	kubectl apply -f testserver/build/clusters.open-cluster-management.io_managedclusters.yaml
	go test -v ./test/e2e -coverprofile cover.out -args -ginkgo.v -ginkgo.trace -namespace $(NAMESPACE) -baseURL http://localhost:3000

## Build the functional test image
tests/docker-build:
	@echo "Building $(REGISTRY)/$(IMG)-tests:$(VERSION)"
	docker build . -f test/e2e/build/Dockerfile -t $(REGISTRY)/$(IMG)-tests:$(VERSION)

## Run the downstream functional tests
tests/docker-run:
	docker run --network host \
		--volume ~/.kube/config:/opt/.kube/config \
		$(REGISTRY)/$(IMG)-tests:$(VERSION)
