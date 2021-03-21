# Copyright Contributors to the Open Cluster Management project

.PHONY: local-server

SERVER_IMG := mock-ocm-server
SERVER_URL ?= $(REGISTRY)/$(SERVER_IMG):$(VERSION)

## Build Mock OCM Server Binary
server/build:
	go build -o testserver/bin/local-server main.go

## Start Mock OCM Server Binary
server/run:
	cd testserver && go run main.go

## Build Mock OCM Server Image
server/docker-build:
	docker build . -f ./testserver/build/Dockerfile -t "${SERVER_URL}"

## Push Mock OCM Server Image
server/docker-push:
	docker push "${SERVER_URL}"

## Run Mock OCM Server Image Locally
server/docker-run:
	docker run -p 3000:3000 "${SERVER_URL}"

## Annotate DiscoveryConfig with Mock OCM Server URL
server/annotate:
	oc annotate discoveryconfig discoveryconfig ocmBaseURL=http://mock-ocm-server.open-cluster-management.svc.cluster.local:3000 --overwrite

DUMMY_ENCRYPTION = $(shell echo "ocmAPIToken: dummytoken" | base64)
server/secret:
	cat config/samples/ocm-api-secret.yaml | sed -e "s/ENCRYPTED_TOKEN/$(DUMMY_ENCRYPTION)/g" | kubectl apply -f - || true

## Build, push, and deploy server in cluster
server/deploy:
	@echo "Deploying with image ${SERVER_URL}"
	cd testserver/build/local-server && $(KUSTOMIZE) edit set image testserver="${SERVER_URL}"
	cd testserver/build/local-server && $(KUSTOMIZE) edit set namespace ${NAMESPACE}
	$(KUSTOMIZE) build testserver/build/local-server | kubectl apply -f -
	cd testserver/build/local-server && $(KUSTOMIZE) edit set image testserver="testserver:latest"
	cd testserver/build/local-server && $(KUSTOMIZE) edit set namespace open-cluster-management


############################################################
# e2e test section
############################################################

# Create deployment and configure it to never download image
kind-deploy-testserver:
	@echo Installing mock server
	cd testserver/build/local-server && $(KUSTOMIZE) edit set namespace $(NAMESPACE)
	$(KUSTOMIZE) build testserver/build/local-server | kubectl apply -f -
	cd testserver/build/local-server && $(KUSTOMIZE) edit set namespace open-cluster-management

	@echo "Patch mock server deployment image"
	kubectl patch deployment mock-ocm-server -n $(NAMESPACE) -p "{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"mock-ocm-server\",\"imagePullPolicy\":\"Never\"}]}}}}"
	kubectl patch deployment mock-ocm-server -n $(NAMESPACE) -p "{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"mock-ocm-server\",\"image\":\"$(SERVER_URL)\"}]}}}}"
	kubectl rollout status -n $(NAMESPACE) deployment mock-ocm-server --timeout=60s

kind-load-testserver-image:
	@echo Pushing image to KinD cluster
	kind load docker-image $(SERVER_URL) --name test-discovery