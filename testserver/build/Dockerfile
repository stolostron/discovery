# Copyright Contributors to the Open Cluster Management project

FROM golang:1.15 as builder

WORKDIR /workspace

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o local-server testserver/main.go

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

WORKDIR /
COPY --from=builder /workspace/local-server .
COPY testserver/data/ data/

ENTRYPOINT ["/local-server"]
