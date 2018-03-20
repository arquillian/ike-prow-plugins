PROJECT_NAME:=ike-prow-plugins
PACKAGE_NAME:=github.com/arquillian/ike-prow-plugins

PLUGINS:=test-keeper pr-sanitizer work-in-progress
BINARIES:=$(patsubst %,binaries-%, $(PLUGINS))

BINARY_DIR=${PWD}/bin
CLUSTER_DIR?=${PWD}/cluster
PLUGIN_DEPLOYMENTS_DIR?=$(CLUSTER_DIR)/generated

REGISTRY?=docker.io
DOCKER_REPO?=bartoszmajsak
BUILD_IMAGES:=$(patsubst %,build-%, $(PLUGINS))
PUSH_IMAGES:=$(patsubst %,push-%, $(PLUGINS))
CLEAN_IMAGES:=$(patsubst %,clean-%, $(PLUGINS))
OC_DEPLOYMENTS:=$(patsubst %,oc-%, $(PLUGINS))

in_docker_group:=$(filter docker,$(shell groups))
is_root:=$(filter 0,$(shell id -u))
DOCKER?=$(if $(or $(in_docker_group),$(is_root)),docker,sudo docker)

.DEFAULT_GOAL := all

.PHONY: all
all: clean install build build-images push-images oc-apply ## (default) Performs clean build  and container packaging

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

.PHONY: install
install: ## Fetches all dependencies using Glide
	glide install -v

.PHONY: up
up: ## Updates all dependencies defined for glide
	glide up -v

.PHONY: compile
compile: up compile-only ## Compiles all plugins and puts them in the bin/ folder (calls up target)

.PHONY: compile-only
compile-only: $(BINARIES)

.PHONY: test
test:
	ginkgo -r

.PHONY: build
build: compile test check

.PHONY: imports ## Removes unneeded imports and formats source code
imports:
	@goimports -l -w pkg

# Build configuration
BUILD_TIME=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
COMMIT:=$(shell git rev-parse --short HEAD)
TAG:=$(COMMIT)-$(shell date +%s)
GITUNTRACKEDCHANGES:=$(shell git status --porcelain --untracked-files=no)
ifneq ($(GITUNTRACKEDCHANGES),)
  COMMIT := $(COMMIT)-dirty
endif
LDFLAGS="-X main.Commit=${COMMIT} -X main.BuildTime=${BUILD_TIME}"

$(BINARIES): binaries-%: %
	@echo "Building $< binary"
	@cd ./pkg/plugin/$< && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags ${LDFLAGS} -o ${BINARY_DIR}/$<

.PHONY: check
check: ## Concurrently runs a whole bunch of static analysis tools
	gometalinter --vendor --deadline 300s ./...

.PHONY: oc-generate-deployments
oc-generate-deployments: $(OC_DEPLOYMENTS) ## Creates openshift deployments for ike-prow plugins


OC_PROJECT_NAME?=ike-prow-plugins
.PHONY: oc-init-project
oc-init-project: ## Initiliazes new project with config maps and secrets
	@echo "Setting cluster project"
	@oc new-project $(OC_PROJECT_NAME)
	@oc create configmap plugins && oc create configmap config
	@oc create secret generic hmac-token --from-file=hmac=config/hmac.token
	@oc create secret generic oauth-token --from-file=oauth=config/oauth.token
	@oc create secret generic sentry-dsn --from-file=sentry=config/sentry.dsn
	@oc create configmap plugins --from-file=plugins=plugins.yaml --dry-run -o yaml | oc replace configmap plugins -f -
	@oc create configmap config --from-file=config=config.yaml --dry-run -o yaml | oc replace configmap config -f -
	@oc create secret generic hmac-token --from-file=hmac=config/hmac.token --dry-run -o yaml | oc replace secret generic hmac-token  -f -
	@oc create secret generic oauth-token --from-file=oauth=config/oauth.token --dry-run -o yaml | oc replace secret generic oauth-token  -f -
	@oc create secret generic sentry-dsn --from-file=sentry=config/sentry.dsn --dry-run -o yaml | oc replace secret generic sentry-dsn  -f -

.PHONY: oc-deploy-starter
oc-deploy-starter: ## Deploys basic prow infrastructure
	@echo "Deploying prow infrastructure"
	@oc apply -f cluster/starter.yaml

HOOK_VERSION?=v20180316-93ade3390
.PHONY: oc-deploy-hook
oc-deploy-hook: ## Deploys hook service only
	@echo "Deploys hook service ${HOOK_VERSION}"
	@oc process -f $(CLUSTER_DIR)/hook-template.yaml \
    		-p VERSION=$(HOOK_VERSION) \
    		-o yaml | oc apply -f -

.PHONY: oc-apply ## Builds plugin images, updates configuration and deploys new version of ike-plugins
oc-apply: build-images push-images oc-generate-deployments
	@echo "Updating cluster configuration..."
	@oc create configmap plugins --from-file=plugins=plugins.yaml --dry-run -o yaml | oc replace configmap plugins -f -
	@oc create configmap config --from-file=config=config.yaml --dry-run -o yaml | oc replace configmap config -f -
	@oc create secret generic hmac-token --from-file=hmac=config/hmac.token --dry-run -o yaml | oc replace secret generic hmac-token  -f -
	@oc create secret generic oauth-token --from-file=oauth=config/oauth.token --dry-run -o yaml | oc replace secret generic oauth-token  -f -
	@oc create secret generic sentry-dsn --from-file=sentry=config/sentry.dsn --dry-run -o yaml | oc replace secret generic sentry-dsn  -f -

$(OC_DEPLOYMENTS): oc-%: %
	@mkdir -p $(PLUGIN_DEPLOYMENTS_DIR)
	@oc process -f $(CLUSTER_DIR)/ike-prow-template.yaml \
		-p REGISTRY=$(REGISTRY) \
		-p DOCKER_REPO=$(DOCKER_REPO) \
		-p PLUGIN_NAME=$< \
		-p VERSION=$(TAG) \
		-o yaml > $(PLUGIN_DEPLOYMENTS_DIR)/$<.yaml

	@oc apply -f $(PLUGIN_DEPLOYMENTS_DIR)/$<.yaml

.PHONY: build-images $(PLUGINS)
build-images: compile $(BUILD_IMAGES)
$(BUILD_IMAGES): build-%: %
	$(DOCKER) build --build-arg PLUGIN_NAME=$< -t $(REGISTRY)/$(DOCKER_REPO)/$<:$(TAG) -f Dockerfile.builder .

.PHONY: clean-images
clean-images: $(CLEAN_IMAGES)
$(CLEAN_IMAGES): clean-%: %
	$(DOCKER) rmi -f $(REGISTRY)/$(DOCKER_REPO)/$<:$(TAG)

.PHONY: push-images
push-images: $(PUSH_IMAGES)
$(PUSH_IMAGES): push-%: %
	$(DOCKER) push $(REGISTRY)/$(DOCKER_REPO)/$<:$(TAG)
