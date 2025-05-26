#!/bin/bash

TARGET_OS=$1
TARGET_ARCH=$2
OUTPUT_DIR=$3

echo "[*] Building for $TARGET_OS/$TARGET_ARCH..."

mkdir -p "$OUTPUT_DIR"
cd src || exit 1

EXT=""
LDFLAGS=""
if [[ "$TARGET_OS" == "windows" ]]; then
  EXT=".exe"
  LDFLAGS='-ldflags -H=windowsgui'
fi

OUTPUT_FILE="$OUTPUT_DIR/ring-$TARGET_OS-$TARGET_ARCH$EXT"

GOOS=$TARGET_OS GOARCH=$TARGET_ARCH go build $LDFLAGS -o "$OUTPUT_FILE"

if [[ $? -ne 0 ]]; then
  echo "[!] Build failed!"
  exit 1
fi

cd .. || exit 1
echo "[*] Build succeeded"
rm -f ./src/add.go

echo "[+] Done! Binary saved to $OUTPUT_FILE"
