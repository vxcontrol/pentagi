# syntax=docker/dockerfile:1.4

# ========================================
# Stage 1: Frontend Application Build
# ========================================
FROM node:23-slim AS frontend-compiler

# Production build configuration
ENV NODE_ENV=production
ENV VITE_BUILD_MEMORY_LIMIT=4096
ENV NODE_OPTIONS="--max-old-space-size=4096"

WORKDIR /app/ui

# Install build essentials
RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    gcc \
    g++ \
    make \
    git

# GraphQL schema for code generation
COPY ./backend/pkg/graph/schema.graphqls ../backend/pkg/graph/

# Application source code
COPY frontend/ .

# Install dependencies with package manager detection for SBOM
RUN --mount=type=cache,target=/root/.npm \
    npm ci --include=dev

# Generate license report for frontend dependencies
RUN npm install -g license-checker && \
    mkdir -p /licenses/frontend && \
    license-checker --production --json > /licenses/frontend/licenses.json && \
    license-checker --production --csv > /licenses/frontend/licenses.csv

# Build frontend with optimizations and parallel processing
RUN npm run build -- \
    --mode production \
    --minify esbuild \
    --outDir dist \
    --emptyOutDir \
    --sourcemap false \
    --target es2020

# ========================================
# Stage 2: Backend Services Compilation
# ========================================
FROM golang:1.24-bookworm AS api-builder

# Version injection arguments
ARG PACKAGE_VER=develop
ARG PACKAGE_REV=

# Static binary compilation settings
ENV CGO_ENABLED=0
ENV GO111MODULE=on

# Install compilation toolchain and dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    gcc \
    g++ \
    make \
    git \
    musl-dev

WORKDIR /app/backend

COPY backend/ .

# Fetch Go module dependencies (cached for faster rebuilds)
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify

# Install go-licenses tool for license extraction
RUN --mount=type=cache,target=/go/pkg/mod \
    go install github.com/google/go-licenses@latest

# Generate license reports for backend dependencies
RUN mkdir -p /licenses/backend && \
    go list -m all > /licenses/backend/dependencies.txt && \
    GOROOT=$(go env GOROOT) GOTOOLCHAIN=auto go-licenses csv ./cmd/pentagi > /licenses/backend/licenses.csv 2>/dev/null || true

# Compile main application binary with embedded version metadata
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

# ========================================
# Stage 3: Production Runtime Environment
# ========================================
FROM alpine:3.23.3

# Establish non-privileged execution context with docker socket access
RUN addgroup -g 998 docker && \
    addgroup -S pentagi && \
    adduser -S pentagi -G pentagi && \
    addgroup pentagi docker

# Install required packages
RUN apk --no-cache add ca-certificates openssl openssh-keygen shadow

ADD scripts/entrypoint.sh /opt/pentagi/bin/

RUN sed -i 's/\r//' /opt/pentagi/bin/entrypoint.sh && \
    chmod +x /opt/pentagi/bin/entrypoint.sh

RUN mkdir -p \
    /root/.ollama \
    /opt/pentagi/bin \
    /opt/pentagi/ssl \
    /opt/pentagi/fe \
    /opt/pentagi/logs \
    /opt/pentagi/data \
    /opt/pentagi/conf && \
    chmod 777 /root/.ollama

COPY --from=api-builder /pentagi /opt/pentagi/bin/pentagi
COPY --from=api-builder /ctester /opt/pentagi/bin/ctester
COPY --from=api-builder /ftester /opt/pentagi/bin/ftester
COPY --from=api-builder /etester /opt/pentagi/bin/etester
COPY --from=frontend-compiler /app/ui/dist /opt/pentagi/fe
COPY --from=api-builder /licenses/backend /opt/pentagi/licenses/backend
COPY --from=frontend-compiler /licenses/frontend /opt/pentagi/licenses/frontend

# Copy provider configuration files
COPY examples/configs/custom-openai.provider.yml /opt/pentagi/conf/
COPY examples/configs/deepinfra.provider.yml /opt/pentagi/conf/
COPY examples/configs/deepseek.provider.yml /opt/pentagi/conf/
COPY examples/configs/moonshot.provider.yml /opt/pentagi/conf/
COPY examples/configs/ollama-cloud.provider.yml /opt/pentagi/conf/
COPY examples/configs/ollama-llama318b-instruct.provider.yml /opt/pentagi/conf/
COPY examples/configs/ollama-llama318b.provider.yml /opt/pentagi/conf/
COPY examples/configs/ollama-qwen332b-fp16-tc.provider.yml /opt/pentagi/conf/
COPY examples/configs/ollama-qwq32b-fp16-tc.provider.yml /opt/pentagi/conf/
COPY examples/configs/openrouter.provider.yml /opt/pentagi/conf/
COPY examples/configs/novita.provider.yml /opt/pentagi/conf/
COPY examples/configs/vllm-qwen3.5-27b-fp8.provider.yml /opt/pentagi/conf/
COPY examples/configs/vllm-qwen3.5-27b-fp8-no-think.provider.yml /opt/pentagi/conf/
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
