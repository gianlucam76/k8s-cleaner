# Stage 1: Build the web UI
FROM node:22-alpine AS web-builder
WORKDIR /web
# Copy dependency manifests first for layer caching
COPY web/package.json web/package-lock.json* ./
RUN if [ -f package.json ]; then npm ci; else mkdir -p dist; fi
# Then copy source and build (only re-runs when source changes, not on dep changes)
COPY web/ .
RUN if [ -f package.json ]; then npm run build; else mkdir -p dist; fi

# Stage 2: Build the manager binary
FROM golang:1.26.1 AS builder

ARG BUILDOS
ARG TARGETARCH

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/ cmd/
COPY api/ api/
COPY internal/ internal/
COPY pkg/ pkg/
COPY web/ web/

# Override web/dist/ with built assets from node stage
COPY --from=web-builder /web/dist/ web/dist/

RUN CGO_ENABLED=0 GOOS=$BUILDOS GOARCH=$TARGETARCH go build -a -o manager cmd/main.go

# Stage 3: Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
