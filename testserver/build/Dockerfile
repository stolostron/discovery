# Copyright Contributors to the Open Cluster Management project

FROM golang:1.15 as builder

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum
COPY main.go main.go
COPY api/ api/

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o local-server main.go

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

WORKDIR /
COPY --from=builder /workspace/local-server .
COPY data/ data/

ENTRYPOINT ["/local-server"]
