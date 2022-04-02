#!/bin/sh
set -x
set -e
set -o errexit
set -o nounset
# # bashism, wsl uses dash
# set -o pipefail

if [ -z "${PKG}" ]; then
    echo "PKG must be set"
    exit 1
fi
if [ -z "${BIN}" ]; then
    echo "BIN must be set"
    exit 1
fi
if [ -z "${GOOS}" ]; then
    echo "GOOS must be set"
    exit 1
fi
if [ -z "${GOARCH}" ]; then
    echo "GOARCH must be set"
    exit 1
fi
if [ -z "${VERSION}" ]; then
    echo "VERSION must be set"
    exit 1
fi
if [ -z "${REVISION}" ]; then
    echo "REVISION must be set"
    exit 1
fi

export CGO_ENABLED=0

GIT_SHA=$(git rev-parse HEAD)
GIT_DIRTY=$(git status --porcelain 2> /dev/null)
if [ -z "${GIT_DIRTY}" ]; then
  GIT_TREE_STATE=clean
else
  GIT_TREE_STATE=dirty
fi

BUILD_DATE=$(date '+%Y-%m-%d-%H:%M:%S')
# To optimize the build for alpine linux
# LDFLAGS="${LDFLAGS} -w -linkmode external -extldflags \"-static\""

VERSION_JSON=./internal/infrastructure/storage/embed/version.json
cat <<- EOF > ${VERSION_JSON}
{
  "version":"${VERSION}",
  "revision":"${REVISION}",
  "git": {
    "branch":"${BRANCH}",
    "commit":"${GIT_SHA}"
  },
  "build": {
    "user":"${BUILDUSER}",
    "date": "${BUILD_DATE}"
  }
}
EOF

if [ -z "${OUTPUT_DIR:-}" ]; then
  OUTPUT_DIR=.
fi
OUTPUT=${OUTPUT_DIR}/${BIN}
if [ "${GOOS}" = "windows" ]; then
  OUTPUT="${OUTPUT}.exe"
fi

if [ "${GOARCH}" = "arm" ]; then
  # Build for Raspberry Pi 2 and 3 boards
  GOARM="7"
  # Build for other versions of the Pi – A, A+, B, B+ or Zero
  # GOARM="6"
fi

go build \
    -o ${OUTPUT} \
    -installsuffix "static" \
    ./cmd/**/*.go

rm -f ${VERSION_JSON}; touch ${VERSION_JSON}
