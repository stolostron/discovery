# Copyright Contributors to the Open Cluster Management project

.PHONY: local-server

IMG := mock-ocm-server
URL ?= $(REGISTRY)/$(IMG):$(VERSION)

## Build Mock OCM Server Binary
server/build:
	go build -o testserver/bin/local-server main.go

## Start Mock OCM Server Binary
server/run:
	cd testserver && go run main.go

## Build Mock OCM Server Image
server/docker-build:
	docker build . -f ./testserver/build/Dockerfile -t "${URL}"

## Push Mock OCM Server Image
server/docker-push:
	docker push "${URL}"

## Run Mock OCM Server Image Locally
server/docker-run:
	docker run -p 3000:3000 "${URL}"

## Annotate DiscoveryConfig with Mock OCM Server URL
server/annotate:
	oc annotate discoveryconfig discoveryconfig ocmBaseURL=http://mock-ocm-server.open-cluster-management.svc.cluster.local:3000 --overwrite

DUMMY_ENCRYPTION = $(shell echo "ocmAPIToken: dummytoken" | base64)
server/secret:
	cat config/samples/ocm-api-secret.yaml | sed -e "s/ENCRYPTED_TOKEN/$(DUMMY_ENCRYPTION)/g" | kubectl apply -f - || true

## Build, push, and deploy server in cluster
server/deploy:
	oc apply -k testserver/build/local-server
