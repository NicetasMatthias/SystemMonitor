#!/usr/bin/env bash

set -euo pipefail

if [ "$#" -ne 3 ]; then
    echo "Usage: $0 <version> <commit> <date>"
    exit 1
fi

VERSION="$1"
COMMIT="$2"
DATE="$3"

APP_NAME="system-monitor"
MAIN_PACKAGE="./cmd/system-monitor"
INFO_PKG="github.com/NicetasMatthias/SystemMonitor/internal/info"

rm -rf dist
mkdir -p dist
BUILD_DIR="$(mktemp -d)"
trap 'rm -rf "$BUILD_DIR"' EXIT

LDFLAGS="
    -X ${INFO_PKG}.Version=${VERSION}
    -X ${INFO_PKG}.Commit=${COMMIT}
    -X ${INFO_PKG}.Date=${DATE}
"
build() {
  GOOS="$1"
  GOARCH="$2"

  local NAME="${APP_NAME}_v${VERSION}_${GOOS}_${GOARCH}"
  local TARGET_DIR="${BUILD_DIR}/${NAME}"
  mkdir -p "${TARGET_DIR}"

  echo "Building ${NAME}..."

  GOOS="${GOOS}" \
  GOARCH="${GOARCH}" \
  CGO_ENABLED=0 \
  go build \
      -ldflags "${LDFLAGS}" \
      -o "${TARGET_DIR}/${APP_NAME}" \
      "${MAIN_PACKAGE}"
  tar -C "${TARGET_DIR}" \
      -czf "dist/${NAME}.tar.gz" \
      "${APP_NAME}"
}

build linux amd64
build linux arm64

(
  cd dist
  sha256sum *.tar.gz > checksums.txt
)

echo
echo "Release ${VERSION} built successfully."
echo
echo "Artifacts:"
ls -lh dist/