# Copyright Contributors to the Open Cluster Management project

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
# stolostron/discovery-bundle:$VERSION and stolostron/discovery-catalog:$VERSION.
IMAGE_TAG_BASE ?= stolostron/discovery

# BUNDLE_IMG defines the image:tag used for the bundle.
# You can use it as an arg. (E.g make bundle-build BUNDLE_IMG=<some-registry>/<project-name-bundle>:<tag>)
BUNDLE_IMG ?= $(IMAGE_TAG_BASE)-bundle:v$(VERSION)

# Image URL to use all building/pushing image targets
REGISTRY ?= quay.io/stolostron
IMG ?= discovery-operator
URL ?= $(REGISTRY)/$(IMG):$(VERSION)
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:crdVersions=v1"
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.25

# Namespace to deploy resources into
NAMESPACE ?= multicluster-engine

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

ENCRYPTED = $(shell echo "${OCM_API_TOKEN}" | base64)
secret: ## Generate secret for OCM access
	cat config/samples/ocm-api-secret.yaml | sed -e "s/ENCRYPTED_TOKEN/$(ENCRYPTED)/g" |  sed -e "s/NAMESPACE/$(NAMESPACE)/g" | kubectl apply -f - || true

.PHONY: config
config: ## Create custom resources
	cd config/samples && $(KUSTOMIZE) edit set namespace $(NAMESPACE)
	$(KUSTOMIZE) build config/samples | kubectl apply -f -
	cd config/samples && $(KUSTOMIZE) edit set namespace multicluster-engine

logs: ## Print operator logs
	@kubectl logs -f $(shell kubectl get pod -l app=discovery-operator -o jsonpath="{.items[0].metadata.name}")

annotate: ## Annotate DiscoveryConfig with Mock OCM Server URL.
	kubectl annotate discoveryconfig discovery ocmBaseURL=http://mock-ocm-server.$(NAMESPACE).svc.cluster.local:3000 authBaseURL=http://mock-ocm-server.$(NAMESPACE).svc.cluster.local:3000 -n $(NAMESPACE) --overwrite

unannotate: ## Remove mock server annotation
	kubectl annotate discoveryconfig discovery -n $(NAMESPACE) ocmBaseURL- authBaseURL-

set-copyright:
	@bash ./cicd-scripts/set-copyright.sh

verify: test deploy-and-test manifests

##@ Build

build: generate fmt vet ## Build manager binary.
	go build -o bin/manager main.go

run: manifests generate fmt ## Run a controller from your host.
	go run ./main.go

docker-build: ## Build docker image with the manager.
	docker build -t "${URL}" .

docker-push: ## Push docker image with the manager.
	docker push "${URL}"

podman-build: ## Build podman image with the manager.
	podman build -t "${URL}" .

podman-push: ## Push podman image with the manager.
	podman push "${URL}"

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

.PHONY: bundle
bundle: manifests kustomize ## Generate bundle manifests and metadata, then validate generated files.
	operator-sdk generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	operator-sdk bundle validate ./bundle
	cd config/manager && $(KUSTOMIZE) edit set image controller="discovery-operator:latest"

.PHONY: bundle-build
bundle-build: ## Build the bundle image.
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

.PHONY: bundle-push
bundle-push: ## Push the bundle image.
	$(MAKE) docker-push IMG=$(BUNDLE_IMG)

.PHONY: podman-bundle-build
podman-bundle-build: ## Build the bundle image.
	podman build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

.PHONY: podman-bundle-push
podman-bundle-push: ## Push the bundle image.
	$(MAKE) podman-push IMG=$(BUNDLE_IMG)

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

# Build a catalog image by adding bundle images to an empty catalog using the operator package manager tool, 'opm'.
# This recipe invokes 'opm' in 'semver' bundle add mode. For more information on add modes, see:
# https://github.com/operator-framework/community-operators/blob/7f1438c/docs/packaging-operator.md#updating-your-existing-operator
.PHONY: podman-catalog-build
podman-catalog-build: opm ## Build a catalog image.
	$(OPM) index add --container-tool podman --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS) $(FROM_INDEX_OPT)

# Push the catalog image.
.PHONY: podman-catalog-push
podman-catalog-push: ## Push a catalog image.
	$(MAKE) podman-push IMG=$(CATALOG_IMG)

##@ Testing

.PHONY: test
test: envtest ## Run unit tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test `go list ./... | grep -v e2e` -coverprofile cover.out

integration-tests: ## Run functional/integration tests
	kubectl apply -f testserver/build/clusters.open-cluster-management.io_managedclusters.yaml
	go test -v ./test/e2e -coverprofile cover.out -args -ginkgo.v -ginkgo.trace -namespace $(NAMESPACE)

integration-tests-local: ## Run functional tests with binaries running locally
	kubectl create namespace $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -f testserver/build/clusters.open-cluster-management.io_managedclusters.yaml
	go test -v ./test/e2e -coverprofile cover.out -args -ginkgo.v -ginkgo.trace -namespace $(NAMESPACE) -baseURL http://localhost:3000

deploy-and-test: install deploy server-deploy ## Install CRDs, deploy components, and run tests
	kubectl apply -f testserver/build/clusters.open-cluster-management.io_managedclusters.yaml
	kubectl wait --for=condition=available --timeout=60s deployment/discovery-operator -n $(NAMESPACE)
	kubectl wait --for=condition=available --timeout=60s deployment/mock-ocm-server -n $(NAMESPACE)
	sleep 10
	go test -v ./test/e2e -coverprofile cover.out -args -ginkgo.v -ginkgo.trace -namespace $(NAMESPACE)

docker-build-tests: ## Build the functional test image
	@echo "Building $(REGISTRY)/$(IMG)-tests:$(VERSION)"
	docker build . -f test/e2e/build/Dockerfile -t $(REGISTRY)/$(IMG)-tests:$(VERSION)

docker-run-tests: ## Run the containerized functional tests
	docker run --network host \
		--volume ~/.kube/config:/opt/.kube/config \
		--volume $(shell pwd)/test/e2e/results:/results \
		$(REGISTRY)/$(IMG)-tests:$(VERSION)

podman-build-tests: ## Build the functional test image
	@echo "Building $(REGISTRY)/$(IMG)-tests:$(VERSION)"
	podman build . -f test/e2e/build/Dockerfile -t $(REGISTRY)/$(IMG)-tests:$(VERSION)

podman-run-tests: ## Run the containerized functional tests
	podman run --network host \
		--volume ~/.kube/config:/opt/.kube/config \
		--volume $(shell pwd)/test/e2e/results:/results \
		$(REGISTRY)/$(IMG)-tests:$(VERSION)

scale-test: ## Run scalability test
	go run test/scale/scale.go

-include testserver/Makefile

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest

## Tool Versions
KUSTOMIZE_VERSION ?= v4.5.7
CONTROLLER_TOOLS_VERSION ?= v0.15.0

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(LOCALBIN)/kustomize || { curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@release-0.17

############################################################
# KinD CI section
############################################################

# Create deployment and configure it to never download image
kind-deploy-controller: kustomize
	@echo Installing discovery controller
	kubectl create namespace $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -

	@echo "Creating dummy webhook secret"
	kubectl create secret generic discovery-operator-webhook-service \
		--from-literal=tls.crt="" \
		--from-literal=tls.key="" \
		-n $(NAMESPACE) \
		--dry-run=client -o yaml | kubectl apply -f -

	cd config/default && $(KUSTOMIZE) edit set namespace $(NAMESPACE)
	$(KUSTOMIZE) build config/default | kubectl apply -f -
	cd config/default && $(KUSTOMIZE) edit set namespace open-cluster-management

	@echo "Patch deployment image"
	kubectl patch deployment discovery-operator -n $(NAMESPACE) -p "{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"discovery-operator\",\"imagePullPolicy\":\"Never\"}]}}}}"
	kubectl patch deployment discovery-operator -n $(NAMESPACE) -p "{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"discovery-operator\",\"image\":\"$(URL)\"}]}}}}"
	kubectl patch deployment discovery-operator -n $(NAMESPACE) -p "{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"discovery-operator\",\"env\":[{\"name\":\"ENABLE_WEBHOOKS\",\"value\":\"false\"}]}]}}}}"
	kubectl delete validatingwebhookconfiguration discovery.open-cluster-management.io --ignore-not-found=true
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

kind-debug:
	kubectl get all -n $(NAMESPACE)
	kubectl describe pods -n $(NAMESPACE)
	kubectl logs deployment/mock-ocm-server -n $(NAMESPACE)
	kubectl logs deployment/discovery-operator -n $(NAMESPACE)
