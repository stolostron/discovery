# WORK IN PROGRESS

We are in the process of enabling this repo for community contribution. See wiki [here](https://open-cluster-management.io/concepts/architecture/).

# Discovery

Operator for managing discovered clusters from OpenShift Cluster Manager

## Prerequisites

- Go v1.23+
- kubectl 1.21+
- Operator-sdk v1.22.2
- Docker
- Connection to an existing Kubernetes cluster

## Installation

Before deploying, the Discovery CRDs need to be installed onto the cluster.

```sh
make install
```

### Outside the Cluster

The operator can be run locally against the configured Kubernetes cluster in ~/.kube/config with the following command:

```sh
make run
```

### Inside the Cluster

The operator can also run inside the cluster as a Deployment. To do that first build the container image and push to an accessible image registry:

1. Build the image:

```sh
make docker-build URL=<registry>/<imagename>:<tag>
```

2. Push the image:

```sh
make docker-push URL=<registry>/<imagename>:<tag>
```

3. Deploy the Operator:

```sh
make deploy URL=<registry>/<imagename>:<tag>
```

## Usage

The discovery operator generates `DiscoveredClusters` based on a `DiscoveryConfig` resource. The config takes in a secret containing your OCM api token. To create this secret run the following:

```sh
make secret OCM_API_TOKEN=<OpenShift Cluster Manager API Token>
```

The OpenShift Cluster Manager API Token can be retrieved from [here](https://cloud.redhat.com/openshift/token). This will create a secret named `ocm-api-token` in the current namespace. With the secret created you can then create the `DiscoveryConfig` resource using the following commands:

```sh
make secret
make connfig
```

This will create a `DiscoveryConfig` like the example below:

```yaml
apiVersion: discovery.open-cluster-management.io/v1
kind: DiscoveryConfig
metadata:
  name: discovery
spec:
  credential: ocm-api-token
  filters:
    lastActive: 7
```

Rebuild Image: Wed Jan  8 12:52:04 EST 2025
