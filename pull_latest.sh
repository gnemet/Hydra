#!/bin/bash
# Hydra Auto-Pull and Setup
# This script pulls the latest version from GitHub and initializes it.

REPO="gnemet/Hydra"
ZIP_NAME="hydra_v1.1.5_deploy.zip" # Fallback or pattern

echo "🔍 Checking for the latest Hydra release on GitHub..."

# 1. Get the latest release metadata from GitHub API
LATEST_JSON=$(curl -s "https://api.github.com/repos/$REPO/releases/latest")

# 2. Extract the download URL for the deploy zip
DOWNLOAD_URL=$(echo "$LATEST_JSON" | grep "browser_download_url" | grep "_deploy.zip" | head -n 1 | cut -d '"' -f 4)

if [ -z "$DOWNLOAD_URL" ]; then
    echo "❌ Error: Could not find a deployment zip in the latest release."
    # If API fails or grep fails, try a direct download of the known last version
    echo "📂 Attempting fallback download of v1.1.5..."
    DOWNLOAD_URL="https://github.com/gnemet/Hydra/releases/download/v1.1.5/hydra_v1.1.5_deploy.zip"
fi

FILE_NAME=$(basename "$DOWNLOAD_URL")

echo "⬇️  Downloading $FILE_NAME..."
curl -L -o "$FILE_NAME" "$DOWNLOAD_URL"

# 3. Check for setup_remote.sh internally or externally
# We can try to extract just setup_remote.sh first if needed, 
# but usually we just run it after unzipping.
# Since we just downloaded the zip, we unzip and then find the script.

if ! command -v unzip &> /dev/null; then
    echo "📦 Installing unzip..."
    sudo apt update && sudo apt install -y unzip
fi

echo "📦 Unpacking $FILE_NAME..."
unzip -o "$FILE_NAME"

# 4. Find and run the remote setup script
SETUP_SCRIPT=$(find . -name "setup_remote.sh" | head -n 1)

if [ -f "$SETUP_SCRIPT" ]; then
    echo "🚀 Running internal setup script ($SETUP_SCRIPT)..."
    bash "$SETUP_SCRIPT"
else
    echo "❌ Error: setup_remote.sh not found inside the package."
    exit 1
fi
