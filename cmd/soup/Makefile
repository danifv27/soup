ifeq ($(V),1)
  Q =
  PROGRESS = --progress plain
else
  Q = @
  PROGRESS = 
endif

# The binary to build (just the basename).
BIN ?= soup

# This repo's root import path (under GOPATH).
PKG := github.com/danifv27/soup

# Where to push the docker image.
# DOCKER_REGISTRY ?= registry.hub.docker.com
DOCKER_REGISTRY ?= registry.tools.3stripes.net

# Which architecture to build - see $(ALL_ARCH) for options.
# if the 'local' rule is being run, detect the ARCH from 'go env'
# if it wasn't specified by the caller.
local : ARCH ?= $(shell go env GOOS)-$(shell go env GOARCH)
ARCH ?= linux-amd64

REVISION := $(shell echo $$(git rev-parse --short HEAD) ||echo "Unknown Revision")

BUILDINFO_TAG ?= $(shell echo $$(git describe --long --all | tr '/' '-')$$(git diff-index --quiet HEAD -- || echo '-dirty-'$$(git diff-index -u HEAD | openssl sha1 | cut -c 10-17)))
VCS_TAG := $(shell git describe --tags)
ifeq ($(VCS_TAG),)
VCS_TAG := $(BUILDINFO_TAG)
endif

ifeq ($(VERSION),)
VERSION := $(VCS_TAG)
endif

TAG_LATEST ?= false

platform_temp = $(subst -, ,$(ARCH))
GOOS = $(word 1, $(platform_temp))
GOARCH = $(word 2, $(platform_temp))

BASE64_PASSWORD_DANIFV27 = ZHVtbXlWYWx1ZQ==

# timestamp value that is formatted according to the RFC3339 standard
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
BUILD_TARGET ?= 

# Set default base image dynamically for each arch
ifeq ($(GOARCH),amd64)
ifeq ($(DEBUG),true)
	DOCKERFILE ?= Dockerfile.$(BIN).debug
	VERSION_SUFFIX := -DBG
	BUILD_CACHE :=
else
	DOCKERFILE ?= Dockerfile.$(BIN)
	VERSION_SUFFIX := 
	BUILD_CACHE :=  --no-cache
endif
endif

GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_SHORT_COMMIT=$(shell git rev-parse --short HEAD)
IMAGE_NAME = $(DOCKER_REGISTRY)/ebtpc-commons/$(BIN)
IMAGE_NAME_LC = $(shell echo $(IMAGE_NAME) | tr A-Z a-z)

VCS_USER ?= danifv27
VCS_PASSWORD = $(shell echo $(BASE64_PASSWORD_DANIFV27) | base64 --decode)
VCS_PROTOCOL ?= https
VCS_URL ?= github.com/danifv27/soup.git
VCS_REF:= $(shell git rev-parse HEAD)
VCS_BRANCH := $(shell git symbolic-ref --short -q HEAD)

TOP_LEVEL = ../..

include $(TOP_LEVEL)/common.mk