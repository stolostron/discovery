# Copyright Contributors to the Open Cluster Management project

FROM golang:1.20 as builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY . .

# Build
RUN go build -a -o local-server testserver/main.go

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

WORKDIR /
COPY --from=builder /workspace/local-server .
COPY testserver/data/ data/

ENTRYPOINT ["/local-server"]
