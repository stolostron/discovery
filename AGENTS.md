# discovery — Agent Instructions

This repository contains the controller for discovering and importing clusters into RHACM.

## What this controller does

The discovery controller automates cluster discovery and import by:

- Discovering OpenShift and Kubernetes clusters in cloud providers (AWS, Azure, GCP)
- Using cloud provider APIs to enumerate clusters
- Creating DiscoveredCluster resources for found clusters
- Facilitating cluster import into RHACM
- Managing discovery credentials and configurations

## Repository layout

- `api/` - CRD definitions for DiscoveryConfig and DiscoveredCluster
- `controllers/` - Reconciliation logic for discovery and import
- `pkg/` - Cloud provider clients (AWS, Azure, GCP, OCM)
- `test/` - Unit and integration tests
- `config/` - Deployment manifests and samples

## Development workflow

### Building locally

```bash
make build        # Build controller binary
make docker-build # Build controller image
```

### Running locally

```bash
# Run controller outside cluster (for development)
make run

# Deploy controller to cluster
make deploy
```

### Testing

```bash
make test         # Run unit tests
make test-e2e     # Run end-to-end tests
```

## Dependencies

- **OpenShift 4.x / Kubernetes 1.19+** - Target platform
- **Cloud provider SDKs** - AWS SDK, Azure SDK, GCP SDK
- **OCM APIs** - Open Cluster Management
- **Cloud credentials** - AWS access keys, Azure service principals, GCP service accounts

## Documentation

- [Discovery API Reference](https://access.redhat.com/documentation/en-us/red_hat_advanced_cluster_management_for_kubernetes/)
- [Supported Cloud Providers](docs/)
- [Discovery Configuration Guide](config/samples/)

## Common tasks

### Create a discovery configuration

```bash
# Create a DiscoveryConfig for AWS
oc apply -f config/samples/discovery_v1_discoveryconfig_aws.yaml

# Watch discovered clusters
oc get discoveredclusters -A
```

### Test discovery locally

```bash
# Set cloud credentials
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...

# Run controller
make run
```

### Debug controller logs

```bash
oc logs -n multicluster-engine deployment/discovery-operator -f
```
