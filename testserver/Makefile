# Copyright Contributors to the Open Cluster Management project

.PHONY: local-server

## Build Mock OCM Server Binary
server/build:
	go build -o testserver/bin/local-server main.go

## Start Mock OCM Server Binary
server/run:
	go run testserver/main.go

## Build Mock OCM Server Image
server/docker-build:
	docker build . -f ./testserver/build/Dockerfile -t quay.io/jakobgray/mock-ocm-server:latest

## Push Mock OCM Server Image
server/docker-push:
	docker push quay.io/jakobgray/mock-ocm-server:latest

## Run Mock OCM Server Image Locally
server/docker-run:
	docker run -p 3000:3000 quay.io/jakobgray/mock-ocm-server:latest

## Annotate DiscoveryConfig with Mock OCM Server URL
server/annotate:
	oc annotate discoveryconfig discoveryconfig ocmBaseURL=http://mock-ocm-server.open-cluster-management.svc.cluster.local:3000 --overwrite

## Build, push, and deploy server in cluster
server/deploy:
	oc apply -k testserver/build/local-server
