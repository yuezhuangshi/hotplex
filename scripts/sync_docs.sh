#!/usr/bin/env bash

# HotPlex Documentation SSOT Sync Script
# This script serves as the Single Source of Truth (SSOT) builder.
# It copies markdown files from their core repository locations into the VitePress docs-site structure.

set -e

echo "🔄 Synchronizing documentation sources to docs-site..."

# Ensure target directories exist
mkdir -p docs-site/guide
mkdir -p docs-site/sdks
mkdir -p docs-site/reference
mkdir -p docs-site/public/images

# --- Guides ---
cp README.md docs-site/guide/getting-started.md
cp docs/quick-start.md docs-site/guide/quick-start.md
cp docs/architecture.md docs-site/guide/architecture.md
cp SECURITY.md docs-site/guide/security.md
cp docs/server/api.md docs-site/guide/websocket.md
cp docs/providers/opencode.md docs-site/guide/opencode-http.md
cp docs/hooks-architecture.md docs-site/guide/hooks.md
cp docs/observability-guide.md docs-site/guide/observability.md
cp docs/docker-deployment.md docs-site/guide/docker.md
cp docs/production-guide.md docs-site/guide/deployment.md
cp docs/benchmark-report.md docs-site/guide/performance.md

# --- SDKs ---
cp docs/sdk-guide.md docs-site/sdks/go-sdk.md
cp sdks/python/README.md docs-site/sdks/python-sdk.md
cp sdks/typescript/README.md docs-site/sdks/typescript-sdk.md

# --- Reference ---
cp docs/server/api.md docs-site/reference/api.md

# --- Assets ---
if [ -d "docs/images" ]; then
    cp -r docs/images/* docs-site/public/images/
fi

if [ -d ".github/assets" ]; then
    mkdir -p docs-site/public/assets
    cp -r .github/assets/* docs-site/public/assets/
fi

# --- Path Fixes for VitePress ---
# In VitePress, images in public/ are accessed with absolute route /images/ (or relatively without docs/).
# Here we'll patch markdown links like 'docs/images/' to '/hotplex/images/' or just '/images/' relative to root.
# Actually, '/images/' is best if base is handled by vitepress.
find docs-site -name "*.md" -type f -exec sed -i.bak 's|docs/images|/images|g' {} +
find docs-site -name "*.md" -type f -exec sed -i.bak 's|\./images|/images|g' {} +
find docs-site -name "*.md" -type f -exec sed -i.bak 's|\.github/assets|/assets|g' {} +
find docs-site -name "*.bak" -type f -delete
# Also fix any root relative links to other markdown files that broke during copy
# e.g., link to [API](docs/server/api.md) could be broken, but let's stick to ignoreDeadLinks for now

echo "✅ Documentation successfully synchronized."
