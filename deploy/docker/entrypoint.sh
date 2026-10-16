#!/usr/bin/env bash
set -eo pipefail

echo "[graphfolio-init] Starting GraphFolio Unified Backend Services..."

# PID tracking
USER_PID=""
PORTFOLIO_PID=""
BFF_PID=""

# Graceful termination handler
cleanup() {
    echo "[graphfolio-init] Termination signal received. Shutting down child processes..."
    if [ -n "$BFF_PID" ] && kill -0 "$BFF_PID" 2>/dev/null; then
        kill -TERM "$BFF_PID" 2>/dev/null || true
    fi
    if [ -n "$PORTFOLIO_PID" ] && kill -0 "$PORTFOLIO_PID" 2>/dev/null; then
        kill -TERM "$PORTFOLIO_PID" 2>/dev/null || true
    fi
    if [ -n "$USER_PID" ] && kill -0 "$USER_PID" 2>/dev/null; then
        kill -TERM "$USER_PID" 2>/dev/null || true
    fi
    wait 2>/dev/null || true
    echo "[graphfolio-init] All backend processes terminated cleanly."
    exit 0
}
trap cleanup SIGINT SIGTERM

# Enforce internal loopback addresses for microservices
export PORTFOLIO_API_URL="127.0.0.1:50051"
export USER_API_URL="127.0.0.1:50052"
export PORT="8080"

# 1. Start User API (gRPC :50052)
echo "[graphfolio-init] Launching User API on port 50052..."
PORT=50052 /app/user-api &
USER_PID=$!

# 2. Start Portfolio API (gRPC :50051)
echo "[graphfolio-init] Launching Portfolio API on port 50051..."
PORT=50051 /app/portfolio-api &
PORTFOLIO_PID=$!

# 3. Wait for internal gRPC listeners to accept connections
echo "[graphfolio-init] Awaiting gRPC services loopback readiness..."
MAX_ATTEMPTS=50
READY=0
for i in $(seq 1 $MAX_ATTEMPTS); do
    if nc -z 127.0.0.1 50051 2>/dev/null && nc -z 127.0.0.1 50052 2>/dev/null; then
        READY=1
        echo "[graphfolio-init] Internal gRPC microservices are healthy and accepting connections."
        break
    fi
    sleep 0.1
done

if [ "$READY" -ne 1 ]; then
    echo "[graphfolio-init] Error: Timed out waiting for internal gRPC microservices to initialize."
    cleanup
    exit 1
fi

# 4. Start GraphQL BFF (HTTP :8080)
echo "[graphfolio-init] Launching GraphQL BFF on port 8080..."
/app/bff &
BFF_PID=$!

# Wait for BFF or any service failure
wait -n "$BFF_PID" "$PORTFOLIO_PID" "$USER_PID"
EXIT_CODE=$?
echo "[graphfolio-init] One of the backend processes exited with status $EXIT_CODE. Initiating container shutdown..."
cleanup
