#!/bin/sh
set -e

# ---------------------------------------------------------------------------
# Configurable environment variables (with defaults)
# ---------------------------------------------------------------------------
NGINX_PORT=${NGINX_PORT:-3000}
NEO4J_SERVER_PORT=${PORT:-4000}

# ---------------------------------------------------------------------------
# Patch nginx config with actual port values
# ---------------------------------------------------------------------------
sed -i "s/listen 3000;/listen ${NGINX_PORT};/" /etc/nginx/conf.d/default.conf
sed -i "s/127.0.0.1:4000;/127.0.0.1:${NEO4J_SERVER_PORT};/" /etc/nginx/conf.d/default.conf

# ---------------------------------------------------------------------------
# Start services
# ---------------------------------------------------------------------------
echo "Starting neo4j-server on port ${NEO4J_SERVER_PORT}..."
cd /neo4j-server
node index.js &
NEO4J_PID=$!

echo "Starting Nginx on port ${NGINX_PORT}..."
nginx -g "daemon off;" &
NGINX_PID=$!

echo "Both services started. Nginx: PID=$NGINX_PID (port ${NGINX_PORT}), Neo4j-server: PID=$NEO4J_PID (port ${NEO4J_SERVER_PORT})"

# Graceful shutdown: forward signals to both processes (POSIX signal names)
cleanup() {
    echo "Shutting down..."
    kill "$NEO4J_PID" "$NGINX_PID" 2>/dev/null || true
    wait "$NEO4J_PID" 2>/dev/null || true
    wait "$NGINX_PID" 2>/dev/null || true
}
trap cleanup TERM INT

# Poll until either process exits
while kill -0 "$NEO4J_PID" 2>/dev/null && kill -0 "$NGINX_PID" 2>/dev/null; do
    sleep 1
done

echo "A process exited, shutting down..."
cleanup
