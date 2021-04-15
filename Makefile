# Copyright Contributors to the Open Cluster Management project

USE_VENDORIZED_BUILD_HARNESS ?=

ifndef USE_VENDORIZED_BUILD_HARNESS
	ifeq ($(CI),true)
	-include $(shell curl -H 'Accept: application/vnd.github.v4.raw' -L https://api.github.com/repos/open-cluster-management/build-harness-extensions/contents/templates/Makefile.build-harness-bootstrap -o .build-harness-bootstrap; echo .build-harness-bootstrap)
	endif
else
-include vbh/.build-harness-vendorized
endif

# Allow operator-sdk version/binary to be used to be specified externally via
# an environment variable.
#
# Default to currrent dev approach of using a version specific alias or
# symbolic link called "osdk".

OPERATOR_SDK ?= osdk

# Current Operator version
VERSION ?= 0.0.1
# Default bundle image tag
BUNDLE_IMG ?= controller-bundle:$(VERSION)
# Options for 'bundle-build'
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

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

-include testserver/Makefile

all: manager

# Run tests
test: generate fmt vet manifests
	go test `go list ./... | grep -v e2e` -coverprofile cover.out

# Run fast tests that don't require extra binary
unit-tests:
	go test ./... -short -v

# Run tests
integration-tests: install deploy server/deploy
	kubectl apply -f testserver/build/clusters.open-cluster-management.io_managedclusters.yaml
	kubectl wait --for=condition=available --timeout=60s deployment/discovery-operator -n $(NAMESPACE)
	kubectl wait --for=condition=available --timeout=60s deployment/mock-ocm-server -n $(NAMESPACE)
	go test -v ./test/e2e -coverprofile cover.out -args -ginkgo.v -ginkgo.trace -namespace $(NAMESPACE)

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster 
install: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests kustomize
	@echo "Deploying with image ${URL}"
	cd config/manager && $(KUSTOMIZE) edit set image controller="${URL}"
	cd config/default && $(KUSTOMIZE) edit set namespace $(NAMESPACE)
	$(KUSTOMIZE) build config/default | kubectl apply -f -
	# Reset values
	cd config/manager && $(KUSTOMIZE) edit set image controller="discovery-operator:latest"
	cd config/default && $(KUSTOMIZE) edit set namespace open-cluster-management

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=discovery-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet `go list ./... | grep -v test`

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build:
	docker build . -t "${URL}"

# Push the docker image
docker-push:
	docker push "${URL}"

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
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

kustomize:
ifeq (, $(shell which kustomize))
	@{ \
	set -e ;\
	KUSTOMIZE_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$KUSTOMIZE_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/kustomize/kustomize/v3@v3.5.4 ;\
	rm -rf $$KUSTOMIZE_GEN_TMP_DIR ;\
	}
KUSTOMIZE=$(GOBIN)/kustomize
else
KUSTOMIZE=$(shell which kustomize)
endif

# Generate bundle manifests and metadata, then validate generated files.
#
# Note: Generated bundle material must  be committed to be picked up and included
# as part of the ACM composite bundle. The merge tooling assumes this stuff is
# found in operator-sdk (V1's) default output directory (bundle).

.PHONY: bundle
bundle: manifests kustomize
	$(OPERATOR_SDK) generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	$(OPERATOR_SDK) bundle validate ./bundle
	cd config/manager && $(KUSTOMIZE) edit set image controller="discovery-operator:latest"

# Build the bundle image.
.PHONY: bundle-build
bundle-build:
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

# Generate secrets for image pull and OCM access
.PHONY: secrets
ENCRYPTED = $(shell echo "ocmAPIToken: ${OCM_API_TOKEN}" | base64)
secrets:
	cat config/samples/ocm-api-secret.yaml | sed -e "s/ENCRYPTED_TOKEN/$(ENCRYPTED)/g" | kubectl apply -f - || true
	@kubectl create secret docker-registry discovery-operator-pull-secret --docker-server=quay.io --docker-username=$(DOCKER_USER) --docker-password=$(DOCKER_PASS) || true

# Generate secret for OCM access
.PHONY: secret
secret:
	cat config/samples/ocm-api-secret.yaml | sed -e "s/ENCRYPTED_TOKEN/$(ENCRYPTED)/g" | kubectl apply -f - || true

# Create custom resources
.PHONY: samples
samples:
	$(KUSTOMIZE) build config/samples | kubectl apply -f -

logs:
	@kubectl logs -f $(shell kubectl get pod -l app=discovery-operator -o jsonpath="{.items[0].metadata.name}")

# Annotate discoveryconfig to target mock server
annotate:
	kubectl annotate discoveryconfig discoveryconfig ocmBaseURL=http://mock-ocm-server.$(NAMESPACE).svc.cluster.local:3000 --overwrite

# Remove mock server annotation
unannotate:
	kubectl annotate discoveryconfig discoveryconfig ocmBaseURL-

set-copyright:
	@bash ./cicd-scripts/set-copyright.sh

verify: test integration-tests manifests

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

## Build the functional test image
tests/docker-build:
	@echo "Building $(REGISTRY)/$(IMG)-tests:$(VERSION)"
	docker build . -f test/e2e/build/Dockerfile -t $(REGISTRY)/$(IMG)-tests:$(VERSION)

## Run the downstream functional tests
tests/docker-run:
	docker run --network host \
		--volume ~/.kube/config:/opt/.kube/config \
		$(REGISTRY)/$(IMG)-tests:$(VERSION)
