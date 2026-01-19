#!/bin/bash
# Hydra Deployment Packager

DIST_DIR="hydra_dist"
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR/bin"

echo "📦 Packaging Hydra for deployment..."

# 1. Build fresh binaries
make build-all

# 2. Copy Binaries
cp bin/* "$DIST_DIR/bin/"

# 3. Copy Scripts
cp parallel.sh "$DIST_DIR/"
chmod +x "$DIST_DIR/parallel.sh"

# 4. Copy Config & Data
cp .env "$DIST_DIR/"
cp thecus.env "$DIST_DIR/" 2>/dev/null || true
cp users.txt "$DIST_DIR/"

cp passwords.txt "$DIST_DIR/"

# 5. Copy README and Setup
cp README.md "$DIST_DIR/"
cp setup.md "$DIST_DIR/"

echo "✅ Done! Deployment files are in: $DIST_DIR"
echo "You can now copy this folder to your other machine."
