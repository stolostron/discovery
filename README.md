# WORK IN PROGRESS

We are in the process of enabling this repo for community contribution. See wiki [here](https://open-cluster-management.io/concepts/architecture/).
# Discovery

This project manages discovered clusters

## Prerequisites

- go version v1.13+
- operator-sdk version v1.3.0
- yq v4.4.x
- docker
- quay credentials for https://quay.io/organization/rhibmcollab and https://quay.io/organization/open-cluster-management
- Connection to an existing Kubernetes cluster

## Installing

### Required Variables

To run, export your Docker/Quay credentials and your OpenShift Cluster Manager API Token

```bash
$ export DOCKER_USER=<DOCKER_USER>
$ export DOCKER_PASS=<DOCKER_PASS>
$ export OCM_API_TOKEN=<OpenShift Cluster Manager API Token>
```
The OpenShift Cluster Manager API Token can be retrieved from [here](https://cloud.redhat.com/openshift/token).

It is also recommended to set a unique version label when building the image

```bash
$ export VERSION=<A_UNIQUE_VERSION>
```

### Building Image
The image can be built and pushed with
```bash
$ make docker-build
$ make docker-push
```

### Installing
Be sure you are logged in to a Kubernetes cluster, then run

```bash
$ make install
$ make deploy
```
This will create the CRDs, RBAC, and deployment

_To run the image locally instead, run_
```bash
$ make run
```

Once the deployment is running a default DiscoveryConfig and DiscoveredClusterRefresh can be created with
```bash
$ make samples
```
