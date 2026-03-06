# syntax=docker/dockerfile:1.4

# STEP 1: Build the frontend
FROM node:23-slim as fe-build

ENV NODE_ENV=production
ENV VITE_BUILD_MEMORY_LIMIT=4096
ENV NODE_OPTIONS="--max-old-space-size=4096"

WORKDIR /frontend

# Install build essentials
RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    gcc \
    g++ \
    make \
    git

COPY ./backend/pkg/graph/schema.graphqls ../backend/pkg/graph/
COPY frontend/ .

# Install dependencies with package manager detection for SBOM
RUN --mount=type=cache,target=/root/.npm \
    npm ci --include=dev

# Build frontend with optimizations and parallel processing
RUN npm run build -- \
    --mode production \
    --minify esbuild \
    --outDir dist \
    --emptyOutDir \
    --sourcemap false \
    --target es2020

# STEP 2: Build the backend
FROM golang:1.24-bookworm as be-build

# Build arguments for version information
ARG PACKAGE_VER=develop
ARG PACKAGE_REV=

ENV CGO_ENABLED=0
ENV GO111MODULE=on

# Install build essentials
RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    gcc \
    g++ \
    make \
    git \
    musl-dev

WORKDIR /backend

COPY backend/ .

# Download dependencies with module detection for SBOM
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Build backend with version information
RUN go build -trimpath \
    -ldflags "\
        -X pentagi/pkg/version.PackageName=pentagi \
        -X pentagi/pkg/version.PackageVer=${PACKAGE_VER} \
        -X pentagi/pkg/version.PackageRev=${PACKAGE_REV}" \
    -o /pentagi ./cmd/pentagi

# Build ctester utility
RUN go build -trimpath \
    -ldflags "\
        -X pentagi/pkg/version.PackageName=ctester \
        -X pentagi/pkg/version.PackageVer=${PACKAGE_VER} \
        -X pentagi/pkg/version.PackageRev=${PACKAGE_REV}" \
    -o /ctester ./cmd/ctester

# Build ftester utility
RUN go build -trimpath \
    -ldflags "\
        -X pentagi/pkg/version.PackageName=ftester \
        -X pentagi/pkg/version.PackageVer=${PACKAGE_VER} \
        -X pentagi/pkg/version.PackageRev=${PACKAGE_REV}" \
    -o /ftester ./cmd/ftester

# Build etester utility
RUN go build -trimpath \
    -ldflags "\
        -X pentagi/pkg/version.PackageName=etester \
        -X pentagi/pkg/version.PackageVer=${PACKAGE_VER} \
        -X pentagi/pkg/version.PackageRev=${PACKAGE_REV}" \
    -o /etester ./cmd/etester

# STEP 3: Build the final image
FROM alpine:3.23.3

# Create non-root user and docker group with specific GID
RUN addgroup -g 998 docker && \
    addgroup -S pentagi && \
    adduser -S pentagi -G pentagi && \
    addgroup pentagi docker

# Install required packages
RUN apk --no-cache add ca-certificates openssl shadow

ADD scripts/entrypoint.sh /opt/pentagi/bin/

RUN chmod +x /opt/pentagi/bin/entrypoint.sh

RUN mkdir -p \
    /opt/pentagi/bin \
    /opt/pentagi/ssl \
    /opt/pentagi/fe \
    /opt/pentagi/logs \
    /opt/pentagi/data \
    /opt/pentagi/conf

COPY --from=be-build /pentagi /opt/pentagi/bin/pentagi
COPY --from=be-build /ctester /opt/pentagi/bin/ctester
COPY --from=be-build /ftester /opt/pentagi/bin/ftester
COPY --from=be-build /etester /opt/pentagi/bin/etester
COPY --from=fe-build /frontend/dist /opt/pentagi/fe

# Copy provider configuration files
COPY examples/configs/custom-openai.provider.yml /opt/pentagi/conf/
COPY examples/configs/deepinfra.provider.yml /opt/pentagi/conf/
COPY examples/configs/deepseek.provider.yml /opt/pentagi/conf/
COPY examples/configs/moonshot.provider.yml /opt/pentagi/conf/
COPY examples/configs/ollama-llama318b-instruct.provider.yml /opt/pentagi/conf/
COPY examples/configs/ollama-llama318b.provider.yml /opt/pentagi/conf/
COPY examples/configs/ollama-qwen332b-fp16-tc.provider.yml /opt/pentagi/conf/
COPY examples/configs/ollama-qwq32b-fp16-tc.provider.yml /opt/pentagi/conf/
COPY examples/configs/openrouter.provider.yml /opt/pentagi/conf/
COPY examples/configs/vllm-qwen332b-fp16.provider.yml /opt/pentagi/conf/

COPY LICENSE /opt/pentagi/LICENSE
COPY NOTICE /opt/pentagi/NOTICE
COPY EULA.md /opt/pentagi/EULA
COPY EULA.md /opt/pentagi/fe/EULA.md

RUN chown -R pentagi:pentagi /opt/pentagi

WORKDIR /opt/pentagi

USER pentagi

ENTRYPOINT ["/opt/pentagi/bin/entrypoint.sh", "/opt/pentagi/bin/pentagi"]

# Image Metadata
LABEL org.opencontainers.image.source="https://github.com/vxcontrol/pentagi"
LABEL org.opencontainers.image.description="Fully autonomous AI Agents system capable of performing complex penetration testing tasks"
LABEL org.opencontainers.image.authors="PentAGI Development Team"
LABEL org.opencontainers.image.licenses="MIT License"
