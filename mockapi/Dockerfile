#-------------------------------------------------------------------------------------------------------------
# Copyright (c) Microsoft Corporation. All rights reserved.
# Licensed under the MIT License. See https://go.microsoft.com/fwlink/?linkid=2090316 for license information.
#-------------------------------------------------------------------------------------------------------------

FROM golang:1.12-alpine3.9 as builder

# Install certs, git, and mercurial
RUN apk add --no-cache ca-certificates git mercurial

WORKDIR /workspace

# Copy go.mod etc and download dependencies (leverage docker layer caching)
COPY go.mod go.mod
COPY go.sum go.sum
ENV GO111MODULE=on
RUN go mod download

# Copy source code over
COPY mockapi/ mockapi/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o mock-databricks-api ./mockapi/

# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot 
WORKDIR /
COPY --from=builder /workspace/mock-databricks-api .
USER nonroot:nonroot

ENTRYPOINT [ "/mock-databricks-api" ]
