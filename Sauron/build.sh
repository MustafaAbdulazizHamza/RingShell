#!/bin/bash

TARGET_OS=$1
TARGET_ARCH=$2
OUTPUT_DIR=$3

echo "[*] Building for $TARGET_OS/$TARGET_ARCH..."

mkdir -p "$OUTPUT_DIR"
cd src
GOOS=$TARGET_OS GOARCH=$TARGET_ARCH go build -o "$OUTPUT_DIR/ring-$TARGET_OS-$TARGET_ARCH"

if [[ $? -ne 0 ]]; then
  echo "[!] Build failed!"
  exit 1
fi
cd ..
echo "[*] Build succeeded"
rm -f ./src/add.go

echo "[+] Done! Binary saved to $OUTPUT_DIR/ring-$TARGET_OS-$TARGET_ARCH"
