#!/bin/bash

set -e

echo "Building NEXUS-API binary..."

mkdir -p bin

echo "Building nexus (all-in-one)..."
go build -o bin/nexus ./cmd/nexus

echo ""
echo "✅ Build complete!"
echo ""
echo "Binary created:"
echo "  - bin/nexus  (CLI, TUI, AI, Mock, Collab - all in one)"
echo ""
echo "Quick start:"
echo "  ./bin/nexus run examples/collections/rest-api.yaml"
echo "  ./bin/nexus tui examples/collections/graphql.yaml"
echo "  ./bin/nexus mock 9999"
echo "  ./bin/nexus collab 8080"
echo "  ./bin/nexus ai generate-tests 'GET /users returns 200'"
echo ""
