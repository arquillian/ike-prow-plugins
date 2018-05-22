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

COMMIT:=$(shell git rev-parse --short HEAD)
TIMESTAMP:=$(shell date +%s)
TAG?=$(COMMIT)-$(TIMESTAMP)

in_docker_group:=$(filter docker,$(shell groups))
is_root:=$(filter 0,$(shell id -u))
DOCKER?=$(if $(or $(in_docker_group),$(is_root)),docker,sudo docker)

.DEFAULT_GOAL := all

CUR_DIR = $(shell pwd)

include ./.make/Makefile.build
include ./.make/Makefile.openshift
include ./.make/Makefile.deploy.prow
include ./.make/Makefile.docker.build

.PHONY: all
all: clean generate install build oc-deploy-hook oc-deploy-plugins ## (default) Performs clean build  and container packaging

help: ## Hey! That's me!
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-10s\033[0m %s\n", $$1, $$2}'
  