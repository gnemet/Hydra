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
cp users.txt "$DIST_DIR/"
cp passwords.txt "$DIST_DIR/"

# 5. Create README
cat <<EOF > "$DIST_DIR/README.md"
# 🐉 Hydra Deployment

## Quick Start
1. Edit \`.env\` with your target NAS IP and settings.
2. Ensure you have static IP configured if using direct UTP connection.
3. Run:
   ./parallel.sh

## Requirements
- Linux OS (x86_64)
- curl, grep, awk, split (Standard on most Linux distros)
EOF

echo "✅ Done! Deployment files are in: $DIST_DIR"
echo "You can now copy this folder to your other machine."
