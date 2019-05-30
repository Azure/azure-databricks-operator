# Build the manager binary
FROM golang:1 as builder

# Copy in the go src
WORKDIR /go/src/microsoft/azure-databricks-operator/databricks-operator
COPY pkg/    pkg/
COPY cmd/    cmd/
COPY vendor/ vendor/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager microsoft/azure-databricks-operator/databricks-operator/cmd/manager

# Copy the controller-manager into a thin image
FROM alpine:latest
ENV DATABRICKS_HOST ""
ENV DATABRICKS_TOKEN ""
WORKDIR /
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY --from=builder /go/src/microsoft/azure-databricks-operator/databricks-operator/manager .
ENTRYPOINT ["/manager"]