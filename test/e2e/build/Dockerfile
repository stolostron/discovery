# Copyright Contributors to the Open Cluster Management project

FROM golang:1.18 as builder

WORKDIR /workspace

COPY api/ api/
COPY test/e2e/ test/e2e/
COPY go.mod go.mod
COPY go.sum go.sum

RUN go get github.com/onsi/ginkgo/v2/ginkgo
RUN ginkgo build test/e2e

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

ENV KUBECONFIG "/opt/.kube/config"

USER root

WORKDIR /

COPY --from=builder /workspace/test/e2e/e2e.test e2e.test

ENTRYPOINT ["/e2e.test", "-ginkgo.v"]
