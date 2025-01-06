# syntax=docker/dockerfile:1.4

# STEP 1: Build the frontend
FROM node:23-slim as fe-build

ENV NODE_ENV=production

WORKDIR /frontend

COPY ./backend/pkg/graph/schema.graphqls ../backend/pkg/graph/
COPY frontend/ .

# Install dependencies with package manager detection for SBOM
RUN --mount=type=cache,target=/root/.yarn \
    yarn install --frozen-lockfile --production=false

RUN yarn build

# STEP 2: Build the backend
FROM golang:1.22-alpine as be-build
ENV CGO_ENABLED=1
RUN apk add --no-cache gcc musl-dev

WORKDIR /backend

COPY backend/ .

# Download dependencies with module detection for SBOM
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

RUN go build -trimpath -o /pentagi ./cmd/pentagi

# STEP 3: Build the final image
FROM alpine:3.21

# Create non-root user and docker group with specific GID
RUN addgroup -g 998 docker && \
    addgroup -S pentagi && \
    adduser -S pentagi -G pentagi && \
    addgroup pentagi docker

# Install required packages
RUN apk --no-cache add ca-certificates openssl shadow && \
    rm -rf /var/cache/apk/*

ADD entrypoint.sh /opt/pentagi/bin/

RUN chmod +x /opt/pentagi/bin/entrypoint.sh

RUN mkdir -p \
    /opt/pentagi/bin \
    /opt/pentagi/ssl \
    /opt/pentagi/fe \
    /opt/pentagi/logs \
    /opt/pentagi/data

COPY --from=be-build /pentagi /opt/pentagi/bin/pentagi
COPY --from=fe-build /frontend/dist /opt/pentagi/fe

COPY LICENSE /opt/pentagi/LICENSE
COPY NOTICE /opt/pentagi/NOTICE
COPY EULA /opt/pentagi/EULA
COPY EULA /opt/pentagi/fe/EULA.md

RUN chown -R pentagi:pentagi /opt/pentagi

WORKDIR /opt/pentagi

USER pentagi

ENTRYPOINT ["/opt/pentagi/bin/entrypoint.sh", "/opt/pentagi/bin/pentagi"]

# Image Metadata
LABEL org.opencontainers.image.source="https://github.com/vxcontrol/pentagi"
LABEL org.opencontainers.image.description="Fully autonomous AI Agents system capable of performing complex penetration testing tasks"
LABEL org.opencontainers.image.authors="PentAGI Development Team"
LABEL org.opencontainers.image.licenses="MIT License"
