#!/bin/sh
set -e

# Always run from the repository root
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
cd "$ROOT_DIR"

rm -f mistatic_deploy.tar

tar -cvf mistatic_deploy.tar \
    --exclude='mistatic_deploy.tar' \
    --exclude='node_modules' \
    --exclude='.git' \
    --exclude='backend/pb_data' \
    --exclude='frontend/build' \
    --exclude='frontend/.svelte-kit' \
    --exclude='backend/mistatic' \
    --exclude='./sites' \
    --exclude='./sites/*' \
    --exclude='mistatic-feature-forms-embedded.md' \
    .
