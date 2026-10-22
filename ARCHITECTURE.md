# Architecture: discovery

## Overview

The `discovery` operator (component name `discovery-operator`) is a
Kubernetes/OpenShift operator that discovers OpenShift/ROSA clusters
registered to a Red Hat account or organization in the **OpenShift Cluster
Manager (OCM)** — the `api.openshift.com` / `console.redhat.com` control
plane — and publishes them as `DiscoveredCluster` custom resources on the
ACM/MCE hub. It can optionally auto-import certain cluster types (ROSA and
MultiClusterEngineHCP/hosted-control-plane clusters) into ACM/MCE as
`ManagedCluster`s.

It is deployed as part of MultiCluster Engine (MCE), typically in the
`multicluster-engine` namespace, and is installed/managed by
`multiclusterhub-operator`/`backplane-operator`. It integrates with the OCM
(Open Cluster Management community project) registration/import pipeline via
`open-cluster-management.io/api` types (`ManagedCluster`, `Placement`,
`ManagedClusterSetBinding`, addon APIs).

> **Note:** the discovery mechanism is entirely HTTP calls against Red Hat's
> OCM REST API (accounts_mgmt and clusters_mgmt) — there is no direct
> integration with cloud-provider SDKs (AWS/Azure/GCP).

## Repository Structure

| Path | Purpose |
|------|---------|
| `main.go` | Entrypoint: manager setup, scheme registration, webhook bootstrap, dynamic Secret watch. |
| `api/v1/`, `api/v1alpha1/` | `DiscoveryConfig` and `DiscoveredCluster` CRD types (`v1` is the storage version; both are served, with no conversion webhook since the schemas are compatible). |
| `controllers/` | `discoveryconfig_controller.go` (polls OCM, syncs `DiscoveredCluster`s), `discoveredcluster_controller.go` (status + auto-import), `managedcluster_controller.go` (syncs managed-cluster state back onto `DiscoveredCluster`s). |
| `pkg/ocm/` | OCM API integration: `auth/` (OAuth2 token retrieval), `subscription/` (accounts_mgmt client), `cluster/` (clusters_mgmt client), `tls_util.go` (webhook TLS profile). |
| `util/` | Shared annotation/label constants, reconcile-interval defaults. |
| `config/` | Kustomize manifests: `crd/`, `default/`, `manager/`, `rbac/`, `samples/`, `scorecard/`. |
| `bundle/` | OLM bundle (CSV, CRDs, NetworkPolicies, scorecard tests). |
| `build/` | Alternate Dockerfiles for Prow/RHTAP/Konflux and test-image builds. |
| `testserver/` | Standalone mock OCM server used in KinD/e2e tests. |
| `test/e2e/`, `test/scale/` | Ginkgo e2e suite and a standalone scale-test program. |
| `.tekton/`, `.github/workflows/` | Konflux pipelines and GitHub Actions CI (KinD-based integration tests). |

## Core Components

### CRDs

- **`DiscoveryConfig`** — a singleton-per-namespace config resource (must be named `discovery`). `Spec.Credential` references a `Secret` holding OCM credentials; `Spec.Filters` narrows results by cluster type, infrastructure provider, last-active window, OpenShift version, and region.
- **`DiscoveredCluster`** — one per cluster found in OCM, named by its external cluster ID. Carries connection info (API URL, console URL), metadata (cloud provider, region, OpenShift version, support level), `ImportAsManagedCluster` (user-settable), `IsManagedCluster` (controller-managed), and `Status.Conditions` (`Available`, `Managed`).
- **Validating webhook**: restricts `ImportAsManagedCluster=true` to `ROSA`/`MultiClusterEngineHCP` cluster types and enforces a valid `DisplayName` pattern.

### Controllers

- **`DiscoveryConfigReconciler`**: reads the credential Secret, calls `ocm.DiscoverClusters()`, diffs the result against existing `DiscoveredCluster`s (create/update/delete), and cross-references live `ManagedCluster`s. Runs every 20 minutes, or immediately on credential Secret changes (via a dynamically registered watch). Fails closed: on unrecoverable auth errors, all `DiscoveredCluster`s in the namespace are deleted rather than left stale.
- **`DiscoveredClusterReconciler`**: recomputes status conditions every 5 minutes and, when `ImportAsManagedCluster=true`, orchestrates auto-import — creating the `ManagedCluster`, an appropriate auto-import `Secret`, and (for HCP clusters) `Placement`/`ManagedClusterSetBinding`/`AddOnDeploymentConfig` resources.
- **`ManagedClusterReconciler`**: watches `ManagedCluster` events to keep `DiscoveredCluster.Spec.IsManagedCluster` in sync, and clears `ImportAsManagedCluster` (marking a "previously auto-imported" annotation) when a `ManagedCluster` is deleted, to avoid re-import loops.

### OCM integration (`pkg/ocm/`)

Hand-rolled HTTP clients (no official OCM SDK):
- **`auth`**: acquires an access token from `sso.redhat.com` via either `refresh_token` (offline-token) or `client_credentials` (service-account) grant.
- **`subscription`**: queries the accounts_mgmt `subscriptions` API with server-side filters (status, activity, external cluster ID) plus additional client-side filtering.
- **`cluster`**: queries clusters_mgmt for the authoritative API URL of ROSA clusters, falling back to a URL heuristic if unavailable.

## Data / Control Flow

```
DiscoveryConfig (applied by user/operator, one per namespace, named "discovery")
        │
        ▼
DiscoveryConfigReconciler   (every 20 min, or immediately on Secret change)
        │  fetch credential Secret → build auth request
        ▼
ocm.DiscoverClusters()
        │  get OAuth2 token → list subscriptions (OCM) → resolve ROSA API URLs
        ▼
DiscoveryConfigReconciler   (diff vs. existing DiscoveredCluster CRs; create/update/delete)
        ▼
DiscoveredCluster CRs
        ├─► DiscoveredClusterReconciler → status conditions + auto-import (ROSA / HCP)
        └─► ManagedClusterReconciler   → sync IsManagedCluster / clear on deletion
```

Credentials never leave the cluster; fresh tokens are requested on every
reconcile rather than cached. `ocmBaseURL`/`authBaseURL` annotations on
`DiscoveryConfig` allow pointing at a mock server for testing.

## Build, Test & Release

- **Makefile**: `manifests`/`generate` (controller-gen), `test` (envtest, excludes e2e), `integration-tests` (Ginkgo e2e against a KinD cluster + mock OCM server), `bundle`/`bundle-build`/`catalog-build` (OLM), KinD helpers (`kind-create-cluster`, `kind-deploy-controller`).
- **Dockerfiles**: `Dockerfile` (community); `build/Dockerfile.rhtap` (Konflux/RHTAP production, post-quantum-crypto-enabled base image); Prow variants for CI.
- **Bundle**: CSV supports `OwnNamespace`/`SingleNamespace` install modes only; ships default-deny NetworkPolicies with explicit allow rules (zero-trust posture).
- **CI**: GitHub Actions `kind.yaml` runs `go mod verify`, format checks, unit tests, then a full KinD-based integration test against the mock OCM server; `.tekton/` Konflux PipelineRuns per MCE release line (e.g. `mce-51`, `mce-52`).
- **SonarQube** and **CodeRabbit** configured for code-quality and AI-assisted review.

## Dependencies & Integrations

- **controller-runtime**, **k8s.io** client libraries.
- **open-cluster-management.io/api** — `ManagedCluster`, `Placement`, `ManagedClusterSetBinding`, addon APIs — the primary integration point with the broader OCM ecosystem.
- **`stolostron/klusterlet-addon-controller`** — `KlusterletAddonConfig` for legacy addon bootstrapping on auto-imported clusters.
- **`openshift/api`** — `APIServer` CR, used to configure webhook server TLS to match cluster-wide crypto policy.
- Consumed/orchestrated by `multiclusterhub-operator` (ACM) and `backplane-operator` (MCE), which install this operator's CSV/manifests as part of their bundles.
- The ACM cluster-import controller (external) consumes the auto-import `Secret` this operator creates to complete the klusterlet registration handshake.

## Conventions & Patterns

- **Flat kubebuilder v4 layout**: `controllers/` at top level rather than `internal/controller/`.
- **Testability pattern**: package-level interface variables (lightweight DI) throughout `pkg/ocm/*` so tests can swap in fakes.
- **Fail-closed on credential loss**: any unrecoverable auth failure wipes all `DiscoveredCluster`s in the affected namespace.
- **Annotation-driven state machine**: a "previously-auto-imported" annotation prevents import/deletion thrash between controllers.
- **Versioning/branching**: `COMPONENT_VERSION` tracks the MCE release; long-lived `backplane-X.Y` release branches mirror MCE release lines.
- **Security posture**: default-deny NetworkPolicy bundle, non-root/no-privilege-escalation security context, dynamic TLS profile compliance for the admission webhook.
