# Copyright Contributors to the Open Cluster Management project

FROM golang:1.23 AS builder

WORKDIR /workspace

COPY api/ api/
COPY test/e2e/ test/e2e/
COPY go.mod go.mod
COPY go.sum go.sum

RUN go install github.com/onsi/ginkgo/v2/ginkgo@latest
RUN ginkgo build test/e2e

FROM registry.access.redhat.com/ubi9/ubi-minimal:latest

ENV KUBECONFIG "/opt/.kube/config"

USER root

WORKDIR /

COPY --from=builder /workspace/test/e2e/e2e.test e2e.test

ENTRYPOINT ["/e2e.test", "--ginkgo.v", "--ginkgo.junit-report=./results/e2e-results.xml"]
