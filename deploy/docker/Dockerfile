# ==============================================================================
# Stage 1: Build Statically Linked Go Binaries
# ==============================================================================
FROM golang:alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /workspace

# Cache Go modules by copying workspace definitions first
COPY go.work go.work.sum ./
COPY pkg/go.mod pkg/go.sum* ./pkg/
COPY proto/go.mod proto/go.sum* ./proto/
COPY bff/go.mod bff/go.sum* ./bff/
COPY services/portfolio-api/go.mod services/portfolio-api/go.sum* ./services/portfolio-api/
COPY services/user-api/go.mod services/user-api/go.sum* ./services/user-api/

RUN go work sync

# Copy all source trees
COPY pkg/ ./pkg/
COPY proto/ ./proto/
COPY bff/ ./bff/
COPY services/portfolio-api/ ./services/portfolio-api/
COPY services/user-api/ ./services/user-api/

# Build optimized static binaries (stripped symbols for minimal image size)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/bff ./bff/cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/portfolio-api ./services/portfolio-api/cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/portfolio-worker ./services/portfolio-api/cmd/worker
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/user-api ./services/user-api/cmd/server

# ==============================================================================
# Stage 2: Minimal Secure Runtime
# ==============================================================================
FROM alpine:3.21 AS runtime

RUN apk add --no-cache ca-certificates tzdata curl bash netcat-openbsd

# Create unprivileged service user
RUN addgroup -S graphfolio && adduser -S -G graphfolio -u 1001 graphfolio

WORKDIR /app

# Copy compiled binaries from builder stage
COPY --from=builder /bin/bff /app/bff
COPY --from=builder /bin/portfolio-api /app/portfolio-api
COPY --from=builder /bin/portfolio-worker /app/portfolio-worker
COPY --from=builder /bin/user-api /app/user-api

# Copy entrypoint supervisor script
COPY deploy/docker/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh /app/bff /app/portfolio-api /app/portfolio-worker /app/user-api && \
    chown -R graphfolio:graphfolio /app

USER graphfolio:graphfolio

EXPOSE 8080

ENTRYPOINT ["/app/entrypoint.sh"]
