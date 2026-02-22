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
cp README_zh.md docs-site/guide/getting-started_zh.md
cp docs/quick-start.md docs-site/guide/quick-start.md
cp docs/architecture.md docs-site/guide/architecture.md
cp docs/architecture_zh.md docs-site/guide/architecture_zh.md
cp SECURITY.md docs-site/guide/security.md
cp docs/server/api.md docs-site/guide/websocket.md
cp docs/providers/opencode.md docs-site/guide/opencode-http.md
cp docs/providers/opencode_zh.md docs-site/guide/opencode-http_zh.md
cp docs/hooks-architecture.md docs-site/guide/hooks.md
cp docs/observability-guide.md docs-site/guide/observability.md
cp docs/docker-deployment.md docs-site/guide/docker.md
cp docs/production-guide.md docs-site/guide/deployment.md
cp docs/benchmark-report.md docs-site/guide/performance.md

# --- SDKs ---
cp docs/sdk-guide.md docs-site/sdks/go-sdk.md
cp docs/sdk-guide_zh.md docs-site/sdks/go-sdk_zh.md
cp sdks/python/README.md docs-site/sdks/python-sdk.md
cp sdks/typescript/README.md docs-site/sdks/typescript-sdk.md

# --- Reference ---
cp docs/server/api.md docs-site/reference/api.md
cp docs/server/api_zh.md docs-site/reference/api_zh.md

# --- Assets ---
if [ -d "docs/images" ]; then
    cp -r docs/images/* docs-site/public/images/
fi

if [ -d ".github/assets" ]; then
    mkdir -p docs-site/public/assets
    cp -r .github/assets/* docs-site/public/assets/
fi

# --- Path Fixes for VitePress ---
# Fix image paths
find docs-site -name "*.md" -type f -exec sed -i.bak 's|docs/images|/images|g' {} +
find docs-site -name "*.md" -type f -exec sed -i.bak 's|\./images|/images|g' {} +
find docs-site -name "*.md" -type f -exec sed -i.bak 's|\.github/assets|/assets|g' {} +

# Fix Bilingual Cross-links
# Go SDK Links
find docs-site -name "*.md" -type f -exec sed -i.bak 's|docs/sdk-guide\.md|/sdks/go-sdk.md|g' {} +
find docs-site -name "*.md" -type f -exec sed -i.bak 's|docs/sdk-guide_zh\.md|/sdks/go-sdk_zh.md|g' {} +
find docs-site -name "*.md" -type f -exec sed -i.bak 's|sdk-guide\.md|/sdks/go-sdk.md|g' {} +
find docs-site -name "*.md" -type f -exec sed -i.bak 's|sdk-guide_zh\.md|/sdks/go-sdk_zh.md|g' {} +

# Architecture Links
find docs-site -name "*.md" -type f -exec sed -i.bak 's|docs/architecture\.md|/guide/architecture.md|g' {} +
find docs-site -name "*.md" -type f -exec sed -i.bak 's|docs/architecture_zh\.md|/guide/architecture_zh.md|g' {} +
find docs-site -name "*.md" -type f -exec sed -i.bak 's|architecture\.md|/guide/architecture.md|g' {} +
find docs-site -name "*.md" -type f -exec sed -i.bak 's|architecture_zh\.md|/guide/architecture_zh.md|g' {} +

# OpenCode Links
find docs-site -name "*.md" -type f -exec sed -i.bak 's|docs/providers/opencode\.md|/guide/opencode-http.md|g' {} +
find docs-site -name "*.md" -type f -exec sed -i.bak 's|docs/providers/opencode_zh\.md|/guide/opencode-http_zh.md|g' {} +
find docs-site -name "*.md" -type f -exec sed -i.bak 's|providers/opencode\.md|/guide/opencode-http.md|g' {} +
find docs-site -name "*.md" -type f -exec sed -i.bak 's|providers/opencode_zh\.md|/guide/opencode-http_zh.md|g' {} +
find docs-site -name "*.md" -type f -exec sed -i.bak 's|opencode\.md|/guide/opencode-http.md|g' {} +
find docs-site -name "*.md" -type f -exec sed -i.bak 's|opencode_zh\.md|/guide/opencode-http_zh.md|g' {} +

# API Reference Links
find docs-site -name "*.md" -type f -exec sed -i.bak 's|docs/server/api\.md|/reference/api.md|g' {} +
find docs-site -name "*.md" -type f -exec sed -i.bak 's|docs/server/api_zh\.md|/reference/api_zh.md|g' {} +
find docs-site -name "*.md" -type f -exec sed -i.bak 's|server/api\.md|/reference/api.md|g' {} +
find docs-site -name "*.md" -type f -exec sed -i.bak 's|server/api_zh\.md|/reference/api_zh.md|g' {} +

# Getting Started / README Links
# Be careful not to replace external URLs containing README.md, but the pattern is specific enough
find docs-site -name "*.md" -type f -exec sed -i.bak 's|README\.md|/guide/getting-started.md|g' {} +
find docs-site -name "*.md" -type f -exec sed -i.bak 's|README_zh\.md|/guide/getting-started_zh.md|g' {} +

# Clean up sed backups
find docs-site -name "*.bak" -type f -delete

echo "✅ Documentation successfully synchronized."
