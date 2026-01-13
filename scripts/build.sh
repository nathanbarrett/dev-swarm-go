#!/bin/bash

set -e

VERSION=${1:-"dev"}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS="-s -w"
LDFLAGS="$LDFLAGS -X github.com/nathanbarrett/dev-swarm-go/pkg/version.Version=$VERSION"
LDFLAGS="$LDFLAGS -X github.com/nathanbarrett/dev-swarm-go/pkg/version.Commit=$COMMIT"
LDFLAGS="$LDFLAGS -X github.com/nathanbarrett/dev-swarm-go/pkg/version.Date=$DATE"

echo "Building dev-swarm $VERSION..."

mkdir -p bin

go build -ldflags "$LDFLAGS" -o bin/dev-swarm ./cmd/dev-swarm

echo "Done: bin/dev-swarm"
