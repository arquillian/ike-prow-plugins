PROJECT_NAME:=ike-prow-plugins
PACKAGE_NAME:=github.com/arquillian/ike-prow-plugins

PLUGINS?=test-keeper pr-sanitizer work-in-progress
BINARIES:=$(patsubst %,binaries-%, $(PLUGINS))

BINARY_DIR:=${PWD}/bin
CLUSTER_DIR?=${PWD}/cluster
PLUGIN_DEPLOYMENTS_DIR?=$(CLUSTER_DIR)/generated

REGISTRY?=docker.io
DOCKER_REPO?=ike-prow-plugins
BUILD_IMAGES:=$(patsubst %,build-%, $(PLUGINS))
PUSH_IMAGES:=$(patsubst %,push-%, $(PLUGINS))
CLEAN_IMAGES:=$(patsubst %,clean-%, $(PLUGINS))
OC_DEPLOYMENTS:=$(patsubst %,oc-%, $(PLUGINS))
OC_RESTART:=$(patsubst %,dc-%, $(PLUGINS))
PROW_VERSION:=$(shell grep "k8s.io/test-infra" ${PWD}/glide.yaml -A 1 | grep version | awk '{print $$2}')

in_docker_group:=$(filter docker,$(shell groups))
is_root:=$(filter 0,$(shell id -u))
DOCKER?=$(if $(or $(in_docker_group),$(is_root)),docker,sudo docker)

.DEFAULT_GOAL := all

.PHONY: all
all: clean generate install build build-images push-images oc-deploy-plugins ## (default) Performs clean build  and container packaging

help: ## Hey! That's me!
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-10s\033[0m %s\n", $$1, $$2}'

.PHONY: clean
clean: ## Removes binary, cache folder and docker images
	@rm -rf ${BINARY_DIR}
	@rm -rf $(PLUGIN_DEPLOYMENTS_DIR)

.PHONY: tools
tools: ## Installs required go tools
	@go get -u github.com/alecthomas/gometalinter && gometalinter --install
	@go get -u github.com/onsi/ginkgo/ginkgo
	@go get -u github.com/onsi/gomega
	@go get -u golang.org/x/tools/cmd/goimports
	@go get -u github.com/jteeuwen/go-bindata/...

.PHONY: install
install: generate ## Fetches all dependencies using Glide
	glide install -v

.PHONY: up
up: generate ## Updates all dependencies defined for glide
	glide up -v

.PHONY: up-skip-test-quick
up-skip-test-quick:
	glide update --skip-test --quick ## Updates only things in glide.yaml and skips tests

.PHONY: compile
compile: up compile-only ## Compiles all plugins and puts them in the bin/ folder (calls up target)

.PHONY: compile-only
compile-only: generate $(BINARIES)

.PHONY: generate ## Generates any necessary assets using go-bindata library
generate:
	go generate ./pkg/assets/...

.PHONY: test
test:
	ginkgo -r

.PHONY: build
build: compile test check

.PHONY: format ## Removes unneeded imports and formats source code
format:
	@goimports -l -w pkg

# Build configuration
BUILD_TIME=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
COMMIT:=$(shell git rev-parse --short HEAD)
TIMESTAMP:=$(shell date +%s)
TAG:=$(COMMIT)-$(TIMESTAMP)
GITUNTRACKEDCHANGES:=$(shell git status --porcelain --untracked-files=no)
ifneq ($(GITUNTRACKEDCHANGES),)
  COMMIT := $(COMMIT)-dirty
endif
LDFLAGS="-X main.Commit=${COMMIT} -X main.BuildTime=${BUILD_TIME}"

$(BINARIES): binaries-%: %
	@echo "Building $< binary"
	@cd ./pkg/plugin/$</cmd && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags ${LDFLAGS} -o ${BINARY_DIR}/$<

.PHONY: check
check: ## Concurrently runs a whole bunch of static analysis tools
	gometalinter --enable=misspell --enable=gosimple --enable-gc --vendor --deadline 300s ./...

.PHONY: oc-generate-deployments
oc-generate-deployments: $(OC_DEPLOYMENTS) ## Creates openshift deployments for ike-prow plugins

define populate_secret ## params: secret filename, secret name, secret key
	@touch config/$(1)
	@oc create secret generic $(2) --from-file=$(3)=config/$(1) || true
	@oc create secret generic $(2) --from-file=$(3)=config/$(1) --dry-run -o yaml | oc replace secret generic $(2) -f -
endef

define populate_configmap ## params: configmap name, configmap file
	@oc create configmap $(1) || true
	@oc create configmap $(1) --from-file=$(1)=$(2) --dry-run -o yaml | oc replace configmap $(1) -f -
endef

define restart_dc ## params: deploymentconfig name
	@oc patch deploymentconfig $(1) --type json -p='[{"op": "replace", "path": "/spec/template/metadata/labels/timestamp", "value":'"'$(TIMESTAMP)'"'}]' || true
endef

OC_PROJECT_NAME?=ike-prow-plugins
PLUGINS_CONFIG?=plugins.yaml
.PHONY: oc-init-project
oc-init-project: ## Initializes new project with config maps and secrets
	@echo "Setting up project '$(OC_PROJECT_NAME)' in the cluster (ignoring potential errors if entries already exist)"
	@oc new-project $(OC_PROJECT_NAME) || true

	$(call populate_configmap,plugins,$(PLUGINS_CONFIG))
	$(call populate_configmap,config,config.yaml)
	$(call populate_secret,oauth.token,oauth-token,oauth)
	$(call populate_secret,hmac.token,hmac-token,hmac)
	$(call populate_secret,sentry.dsn,sentry-dsn,sentry)

.PHONY: oc-deploy-hook
oc-deploy-hook: build-hook build-hook-image push-hook-image ## Deploys hook service only
	@echo "Deploys hook service ${PROW_VERSION}"
	@oc process -f $(CLUSTER_DIR)/hook-template.yaml \
		-p REGISTRY=$(REGISTRY) \
		-p DOCKER_REPO=$(DOCKER_REPO) \
		-p VERSION=$(PROW_VERSION) \
		-o yaml | oc apply -f -

.PHONY: build-hook
build-hook: up-skip-test-quick
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/hook ./vendor/k8s.io/test-infra/prow/cmd/hook

.PHONY: build-hook-image
build-hook-image:
	$(DOCKER) build --build-arg BINARY=hook -t $(REGISTRY)/$(DOCKER_REPO)/hook:$(PROW_VERSION) -f Dockerfile.builder .

.PHONY: push-hook-image
push-hook-image: build-hook-image
	$(DOCKER) push $(REGISTRY)/$(DOCKER_REPO)/hook:$(PROW_VERSION)

.PHONY: clean-hook-image
clean-hook-image:
	$(DOCKER) rmi -f $(REGISTRY)/$(DOCKER_REPO)/hook:$(PROW_VERSION)

.PHONY: oc-deploy-plugins ## Builds plugin images, updates configuration and deploys new version of ike-plugins
oc-deploy-plugins: oc-init-project build-plugin-images push-plugin-images oc-generate-deployments
	@echo "Updating cluster configuration for '$(OC_PROJECT_NAME)'..."

$(OC_DEPLOYMENTS): oc-%: %
	@echo "Creating new deploymet for $<"
	@mkdir -p $(PLUGIN_DEPLOYMENTS_DIR)
	@oc process -f $(CLUSTER_DIR)/ike-prow-template.yaml \
		-p REGISTRY=$(REGISTRY) \
		-p DOCKER_REPO=$(DOCKER_REPO) \
		-p PLUGIN_NAME=$< \
		-p VERSION=$(TAG) \
		-o yaml > $(PLUGIN_DEPLOYMENTS_DIR)/$<.yaml

	@oc apply -f $(PLUGIN_DEPLOYMENTS_DIR)/$<.yaml

.PHONY: oc-restart-all ## Restarts all deploymentconfigs by patching it
oc-restart-all: oc-restart-hook oc-restart-plugins

.PHONY: oc-restart-hook ## Restarts hook deploymentconfig by patching it
oc-restart-hook:
	$(call restart_dc,hook)

.PHONY: oc-restart-plugins ## Restarts plugin deploymentconfigs by patching it
oc-restart-plugins: $(OC_RESTART)
$(OC_RESTART): dc-%: %
	$(call restart_dc,$<)

.PHONY: build-plugin-images $(PLUGINS)
build-plugin-images: compile $(BUILD_IMAGES)
$(BUILD_IMAGES): build-%: %
	$(DOCKER) build --build-arg BINARY=$< -t $(REGISTRY)/$(DOCKER_REPO)/$<:$(TAG) -f Dockerfile.builder .

.PHONY: clean-plugin-images
clean-plugin-images: $(CLEAN_IMAGES)
$(CLEAN_IMAGES): clean-%: %
	$(DOCKER) rmi -f $(REGISTRY)/$(DOCKER_REPO)/$<:$(TAG)

.PHONY: push-plugin-images
push-plugin-images: build-plugin-images $(PUSH_IMAGES)
$(PUSH_IMAGES): push-%: %
	$(DOCKER) push $(REGISTRY)/$(DOCKER_REPO)/$<:$(TAG)
