# The binary to build (just the basename).
BIN ?= soup
# This repo's root import path (under GOPATH).
PKG := github.com/danifv27/soup
BASE64_PASSWORD_DANIFV27 = ZHVtbXlWYWx1ZQ==
IMAGE_NAME = $(DOCKER_REGISTRY)/ebtpc-commons/$(BIN)
VCS_USER ?= danifv27
VCS_PASSWORD = $(shell echo $(BASE64_PASSWORD_DANIFV27) | base64 --decode)
VCS_PROTOCOL ?= https
VCS_URL ?= github.com/danifv27/soup.git


TOP_LEVEL = ../..

include $(TOP_LEVEL)/common.mk

.PHONY: help
help: ## Show This Help
	@for line in $$(cat ./Makefile $(TOP_LEVEL)/common.mk | grep "##" | grep -v "grep" | sed  "s/:.*##/:/g" | sed "s/\ /!/g"); do verb=$$(echo $$line | cut -d ":" -f 1); desc=$$(echo $$line | cut -d ":" -f 2 | sed "s/!/\ /g"); printf "%-30s--%s\n" "$$verb" "$$desc"; done
