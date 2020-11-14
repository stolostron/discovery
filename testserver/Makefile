.PHONY: local-server

DIR ?= $(shell basename `pwd`)
ifeq ($(DIR),testserver)
DIR := .
else
DIR := ./testserver
endif

## Build Mock OCM Server Binary
server/build:
	echo ${DIR}
	cd $(DIR) && go build -o ./bin/local-server main.go

## Start Mock OCM Server Binary
server/run:
	cd $(DIR) && go run main.go

## Build Mock OCM Server Image
server/docker-build:
	cd $(DIR) && docker build . -f ./build/Dockerfile -t quay.io/rhibmcollab/mock-ocm-server:latest

## Push Mock OCM Server Image
server/docker-push:
	docker push quay.io/rhibmcollab/mock-ocm-server:latest

## Run Mock OCM Server Image Locally
server/docker-run:
	docker run -p 3000:3000 quay.io/rhibmcollab/mock-ocm-server:latest

## Annotate DiscoveryConfig with Mock OCM Server URL
server/annotate:
	oc annotate discoveryconfig discoveryconfig OCM_URL=http://mock-ocm-server.open-cluster-management.svc.cluster.local:8080 --overwrite

## Build, push, and deploy server in cluster
server/deploy: server/docker-build server/docker-push
	oc apply -k $(DIR)/build/local-server