# Copyright Contributors to the Open Cluster Management project

.PHONY: local-server

SERVER_IMG := mock-ocm-server
SERVER_URL ?= $(REGISTRY)/$(SERVER_IMG):$(VERSION)

##@ Mock Server

server-build: ## Build Mock OCM Server binary.
	go build -o testserver/bin/local-server main.go

server-run: ## Start Mock OCM Server binary.
	cd testserver && go run main.go

server-docker-build: ## Build Mock OCM Server Image.
	docker build . -f ./testserver/build/Dockerfile -t "${SERVER_URL}"

server-docker-push: ## Push Mock OCM Server Image.
	docker push "${SERVER_URL}"

server-docker-run: ## Run Mock OCM Server Image Locally.
	docker run -p 3000:3000 "${SERVER_URL}"

server-annotate: ## Annotate DiscoveryConfig with Mock OCM Server URL.
	oc annotate discoveryconfig discovery ocmBaseURL=http://mock-ocm-server.$(NAMESPACE).svc.cluster.local:3000 -n $(NAMESPACE) --overwrite

DUMMY_ENCRYPTION = $(shell echo "dummytoken" | base64)
server-secret:
	cat config/samples/ocm-api-secret.yaml | sed -e "s/ENCRYPTED_TOKEN/$(DUMMY_ENCRYPTION)/g" |  sed -e "s/NAMESPACE/$(NAMESPACE)/g" | kubectl apply -f - || true

## Build, push, and deploy server in cluster
server-deploy:
	@echo "Deploying with image ${SERVER_URL}"
	cd testserver/build/local-server && $(KUSTOMIZE) edit set image testserver="${SERVER_URL}"
	cd testserver/build/local-server && $(KUSTOMIZE) edit set namespace ${NAMESPACE}
	$(KUSTOMIZE) build testserver/build/local-server | kubectl apply -f -
	cd testserver/build/local-server && $(KUSTOMIZE) edit set image testserver="mock-ocm-server:latest"
	cd testserver/build/local-server && $(KUSTOMIZE) edit set namespace open-cluster-management

server-update-dates:
	python3 testserver/data/date-update-scripts/update-dates.py

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
