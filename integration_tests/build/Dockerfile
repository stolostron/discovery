FROM golang:1.15 as builder

WORKDIR /workspace

COPY api/ api/
COPY integration_tests/ integration_tests/
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
# RUN go mod download

RUN go get github.com/onsi/ginkgo/ginkgo
# RUN go get github.com/onsi/gomega/...
RUN ginkgo build integration_tests/controller_tests


FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

ENV KUBECONFIG "/opt/.kube/config"

USER root

WORKDIR /
# COPY controller_tests/ controller_tests/
# COPY start-tests.sh start-tests.sh

COPY --from=builder /workspace/integration_tests/controller_tests/controller_tests.test controller_tests.test


ENTRYPOINT ["/controller_tests.test", "-ginkgo.v"]
