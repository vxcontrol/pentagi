#!/bin/bash
#
# Generate license reports for PentAGI dependencies
#

set -e

cd "$(dirname "$0")/.."
LICENSES_DIR="./licenses"

mkdir -p "$LICENSES_DIR"

echo "Generating license reports..."

# Backend (Go)
echo "→ Backend..."
cd backend

# Generate module list
go list -m all > "../$LICENSES_DIR/backend-dependencies.txt"

# Generate detailed license report using go-licenses
if command -v go-licenses &> /dev/null; then
    echo "  Generating detailed license report with go-licenses..."
    GOROOT=$(go env GOROOT) GOTOOLCHAIN=auto go-licenses csv ./cmd/pentagi > "../$LICENSES_DIR/backend-licenses.csv" 2>/dev/null || {
        echo "  go-licenses failed, install it with: go install github.com/google/go-licenses@latest"
    }
else
    echo "  go-licenses not found, install it with: go install github.com/google/go-licenses@latest"
fi

cd ..

# Frontend (npm)
echo "→ Frontend..."
cd frontend
if [ -d "node_modules" ]; then
    npm ls --production --json > "../$LICENSES_DIR/frontend-dependencies.json" 2>/dev/null || true
    
    if command -v license-checker &> /dev/null; then
        license-checker --production --json > "../$LICENSES_DIR/frontend-licenses.json" 2>/dev/null || true
        license-checker --production --csv > "../$LICENSES_DIR/frontend-licenses.csv" 2>/dev/null || true
    fi
else
    echo "  Run 'npm ci' in frontend/ for detailed reports"
fi
cd ..

echo "Done! Reports saved in: $LICENSES_DIR/"
ls -1 "$LICENSES_DIR/" | grep -v -E "(README|gitignore)" | sed 's/^/   - /'
