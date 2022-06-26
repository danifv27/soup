.PHONY: all
all: package tag push

PHONY: init 
init: $(TOP_LEVEL)/go.mod ## Initialize the module

$(TOP_LEVEL)/go.mod:
	$(Q)go mod init $(PKG)
	$(Q)go mod tidy

PHONY: build-dirs
build-dirs:
	@mkdir -p output/$(GOOS)/$(GOARCH)/bin
	@mkdir -p .go/src/$(PKG) .go/pkg .go/bin .go/std/$(GOOS)/$(GOARCH) .go/go-build

CLEAN_PHONY_TARGETS = clean-linux-amd64 clean-darwin-amd64 clean-windows-amd64
.PHONY: $(CLEAN_PHONY_TARGETS)
$(CLEAN_PHONY_TARGETS): clean-%:
	$(Q)$(MAKE) --no-print-directory ARCH=$* clean

.PHONY: clean
clean-%:
	$(Q)$(MAKE) --no-print-directory ARCH=$* clean

clean: ## Clean out all generated items
	$(Q)echo "Cleaning binary: ./output/$(GOOS)/$(GOARCH)/bin/${BIN}"
	$(Q)test -e ./output/$(GOOS)/$(GOARCH)/bin/${BIN} && rm -r ./output/$(GOOS)/$(GOARCH)

.PHONY: coverage
coverage: ## Generates the total code coverage of the project
	$(Q)$(eval COVERAGE_DIR=$(shell mktemp -d))
	$(Q)mkdir -p $(COVERAGE_DIR)/tmp
	$(Q)for j in $$(go list ./... | grep -v '/vendor/' | grep -v '/ext/'); do go test -covermode=count -coverprofile=$(COVERAGE_DIR)/$$(basename $$j).out $$j > /dev/null 2>&1; done
	$(Q)echo 'mode: count' > $(COVERAGE_DIR)/tmp/full.out
	$(Q)tail -q -n +2 $(COVERAGE_DIR)/*.out >> $(COVERAGE_DIR)/tmp/full.out
	$(Q)@go tool cover -func=$(COVERAGE_DIR)/tmp/full.out | tail -n 1 | sed -e 's/^.*statements)[[:space:]]*//' -e 's/%//'

.PHONY: package
package: ## Create a docker image of the project
	@echo "Packaging image: $(VERSION) [$(GIT_COMMIT)]"
	$(Q)docker build $(BUILD_CACHE) \
		--label 'org.label-schema.vcs-url=$(VCS_PROTOCOL)://$(VCS_USER):$(VCS_PASSWORD)@$(VCS_URL)' \
		--label 'org.label-schema.vcs-ref=$(VCS_REF)' \
		--label 'org.label-schema.version=$(VERSION)' \
		--label 'org.label-schema.name=$(BIN)', \
		--label 'org.label-schema.build-date=$(BUILD_DATE)' \
		--label 'org.label-schema.vendor=adidas' \
		$(PROGRESS) $(DOCKER_TARGET) \
		-t $(IMAGE_NAME_LC):local -f $(DOCKERFILE) $(TOP_LEVEL)

.PHONY: tag
tag: ## Tag image created by package with latest, git commit and version
	@echo "Tagging image: ${VERSION}${VERSION_SUFFIX} $(GIT_COMMIT)"
	$(Q)docker tag $(IMAGE_NAME_LC):local $(IMAGE_NAME_LC):$(GIT_SHORT_COMMIT)
	$(Q)docker tag $(IMAGE_NAME_LC):local $(IMAGE_NAME_LC):${VERSION}${VERSION_SUFFIX}

.PHONY: push
push: tag ## Push tagged images to docker registry
	@echo "Pushing docker image to registry: ${VERSION}${VERSION_SUFFIX} $(GIT_SHORT_COMMIT)"
#	$(Q)(echo $(BASE64_PASSWORD) | base64 --decode | docker login -u danifv27 --password-stdin $(DOCKER_REGISTRY))
	$(Q)docker push $(IMAGE_NAME_LC):$(GIT_SHORT_COMMIT)
	$(Q)docker push $(IMAGE_NAME_LC):${VERSION}${VERSION_SUFFIX}
#	$(Q)docker logout $(DOCKER_REGISTRY)

.PHONY: help
help: ## Show This Help
	@for line in $$(cat Makefile | grep "##" | grep -v "grep" | sed  "s/:.*##/:/g" | sed "s/\ /!/g"); do verb=$$(echo $$line | cut -d ":" -f 1); desc=$$(echo $$line | cut -d ":" -f 2 | sed "s/!/\ /g"); printf "%-30s--%s\n" "$$verb" "$$desc"; done

.PHONY: test
test: unit_test ## Run all available tests

.PHONY: unit_test
unit_test: ## Run all available unit tests
	go test -v $(shell go list ./... | grep -v /vendor/)

.PHONY: ktunnel
ktunnel:
	ktunnel -v expose sre-soup-ktunnel-main 8081:8081 -v -r -n origo-devops-dev
# GNU make required targets declared as .PHONY to be explicit
BUILD_PHONY_TARGETS = build-linux-amd64 build-darwin-amd64 build-windows-amd64 build-linux-arm
.PHONY: $(BUILD_PHONY_TARGETS)
$(BUILD_PHONY_TARGETS): build-%:
	$(Q)$(MAKE) --no-print-directory ARCH=$* build

PHONY: build
build: output/$(GOOS)/$(GOARCH)/bin/$(BIN)

output/$(GOOS)/$(GOARCH)/bin/$(BIN): build-dirs
	@echo "Building binary: $@ $(VERSION)"
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	VERSION=$(VERSION) \
	REVISION=$(REVISION) \
	BRANCH=$(VCS_BRANCH) \
	BUILDUSER=$(VCS_USER) \
	PKG=$(PKG) \
	BIN=$(BIN) \
	DEBUG=$(DEBUG) \
	OUTPUT_DIR=./output/$(GOOS)/$(GOARCH)/bin \
	$(TOP_LEVEL)/scripts/build.sh

PHONY: local
local: build-dirs ## Build application for the local arch
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	VERSION=$(VERSION) \
	REVISION=$(REVISION) \
	BRANCH=$(VCS_BRANCH) \
	BUILDUSER=$(VCS_USER) \
	PKG=$(PKG) \
	BIN=$(BIN) \
	OUTPUT_DIR=./output/$(GOOS)/$(GOARCH)/bin \
	DEBUG=$(DEBUG) \
	$(TOP_LEVEL)/scripts/build.sh