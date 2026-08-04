# CLAUDE.md

@AGENTS.md

## Build commands

```bash
make build         # Build controller binary
make docker-build  # Build controller image
make bundle        # Generate operator bundle
```

## Test commands

```bash
make test          # Run unit tests
make test-e2e      # Run end-to-end tests
make lint          # Run linters
```

## Local development

### Run controller locally

```bash
# Install CRDs
make install

# Run controller outside cluster
make run
```

### Deploy to cluster

```bash
# Deploy controller
make deploy

# Create discovery config
oc apply -f config/samples/discovery_v1_discoveryconfig_aws.yaml
```

### Test discovery with credentials

```bash
# Create secret with cloud credentials
oc create secret generic aws-creds \
  --from-literal=aws_access_key_id=$AWS_ACCESS_KEY_ID \
  --from-literal=aws_secret_access_key=$AWS_SECRET_ACCESS_KEY \
  -n open-cluster-management

# Watch discovered clusters
watch oc get discoveredclusters -A
```
