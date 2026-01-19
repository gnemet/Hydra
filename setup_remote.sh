#!/bin/bash
# Hydra Remote Setup Script
# Use this on the target machine to initialize the environment.

ZIP_FILE="hydra_v1.1.5_deploy.zip"
TARGET_DIR="hydra_dist"

echo "🔧 Initializing Hydra Deployment Environment..."

# 1. Check for unzip
if ! command -v unzip &> /dev/null; then
    echo "📦 Installing unzip..."
    sudo apt update && sudo apt install -y unzip
fi

# 2. Unpack
if [ ! -f "$ZIP_FILE" ]; then
    echo "❌ Error: $ZIP_FILE not found. Please upload it to this directory first."
    exit 1
fi

echo "📦 Unpacking $ZIP_FILE..."
unzip -o "$ZIP_FILE"

# 3. Enter directory
if [ -d "$TARGET_DIR" ]; then
    cd "$TARGET_DIR"
else
    echo "❌ Error: Expected directory $TARGET_DIR not found."
    exit 1
fi

# 4. Set Permissions and Ownership
echo "🔐 Setting permissions and ownership..."
chmod +x bin/* parallel.sh
# Ensure the current user owns the deployment files
sudo chown -R $(whoami):$(whoami) .

# 5. Install runtime dependencies
if ! command -v bc &> /dev/null; then
    echo "📊 Installing dependency: bc (for metrics)..."
    sudo apt update && sudo apt install -y bc
fi

echo "----------------------------------------------------"
echo "✅ Setup Complete!"
echo "----------------------------------------------------"
echo "🚀 Next Steps:"
echo "   1. Update .env (URL, success/error markers)"
echo "   2. Launch search: ./parallel.sh"
echo "----------------------------------------------------"
