#!/usr/bin/env bash
set -euo pipefail

VERSION="0.7"
BUILD_DATE=$(date -u +%Y-%m-%d)
OUTPUT_DIR="build"

mkdir -p "$OUTPUT_DIR"

echo "Building mget_temp_exporter v${VERSION} (${BUILD_DATE})"

LDFLAGS="-s -w -X main.version=${VERSION} -X main.buildDate=${BUILD_DATE}"

echo "  -> linux/amd64"
GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "${OUTPUT_DIR}/mget_exporter_linux_amd64" .

echo "  -> linux/arm64"
GOOS=linux GOARCH=arm64 go build -ldflags "$LDFLAGS" -o "${OUTPUT_DIR}/mget_exporter_linux_arm64" .

echo "  -> windows/amd64"
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "${OUTPUT_DIR}/mget_exporter_windows_amd64.exe" .

echo ""
echo "Done. Binaries in ${OUTPUT_DIR}/:"
ls -lh "$OUTPUT_DIR"/
