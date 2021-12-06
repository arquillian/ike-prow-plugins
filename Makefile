PROJECT_NAME:=ike-prow-plugins
PACKAGE_NAME:=github.com/arquillian/ike-prow-plugins

PLUGINS?=test-keeper pr-sanitizer work-in-progress
BINARIES:=$(patsubst %,binaries-%, $(PLUGINS))

BINARY_DIR:=${PWD}/dist
CLUSTER_DIR?=${PWD}/cluster
PLUGIN_DEPLOYMENTS_DIR?=$(CLUSTER_DIR)/generated

PROJECT_DIR:=$(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
REGISTRY?=quay.io
DOCKER_REPO?=ike-prow-plugins
BUILD_IMAGES:=$(patsubst %,build-%, $(PLUGINS))
PUSH_IMAGES:=$(patsubst %,push-%, $(PLUGINS))
CLEAN_IMAGES:=$(patsubst %,clean-%, $(PLUGINS))
OC_DEPLOYMENTS:=$(patsubst %,oc-%, $(PLUGINS))
OC_RESTART:=$(patsubst %,dc-%, $(PLUGINS))

COMMIT:=$(shell git rev-parse --short HEAD)
TIMESTAMP:=$(shell date +%s)
TAG?=$(COMMIT)-$(TIMESTAMP)

in_docker_group:=$(filter docker,$(shell groups))
is_root:=$(filter 0,$(shell id -u))
DOCKER?=$(if $(or $(in_docker_group),$(is_root)),docker,sudo docker)

CUR_DIR = $(shell pwd)
GOPATH_1:=$(shell echo ${GOPATH} | cut -d':' -f 1)
GOBIN=$(GOPATH_1)/bin
PATH:=${GOBIN}/bin:$(PROJECT_DIR)/bin:$(PATH)

# Call this function with $(call header,"Your message") to see underscored green text
define header =
@echo -e "\n\e[92m\e[4m\e[1m$(1)\e[0m\n"
endef

include ./.make/Makefile.build
include ./.make/Makefile.openshift
include ./.make/Makefile.deploy.prow
include ./.make/Makefile.docker.build

.DEFAULT_GOAL := all
.PHONY: all
all: clean tools generate deps build ## (default) Performs clean build

help: ## Hey! That's me!
	 @echo -e "$$(grep -hE '^\S+:.*##' $(MAKEFILE_LIST) | sort | sed -e 's/:.*##\s*/:/' -e 's/^\(.\+\):\(.*\)/\\x1b[36m\1\\x1b[m:\2/' | column -c2 -t -s :)"
