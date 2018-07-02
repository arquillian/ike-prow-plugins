#!/bin/bash

# Output command before executing
set -x

# Exit on error
set -e

# Source environment variables of the jenkins slave
# that might interest this worker.
function load_jenkins_vars() {
  if [ -e "jenkins-env" ]; then
    cat jenkins-env \
      | grep -E "(IMG_REPO_TAG_LEN|IMG_REPO_USERNAME|IMG_REPO_PASSWORD|JENKINS_URL|GIT_BRANCH|GIT_COMMIT|BUILD_NUMBER|ghprbSourceBranch|ghprbActualCommit|BUILD_URL|ghprbPullId)=" \
      | sed 's/^/export /g' \
      > ~/.jenkins-env
    source ~/.jenkins-env
  fi
}

function install_deps() {
  # We need to disable selinux for now, XXX
  /usr/sbin/setenforce 0 || :

  yum -y install git docker make

  service docker start

  echo 'CICO: Dependencies installed'
}

function run_build() {
  trap cleanup_env EXIT;
  make docker-build
  echo "CICO: ran build"
}

function prepare() {
  make docker-start
  make docker-tools
  make docker-install
}

function cleanup_env {
  EXIT_CODE=$?
  echo "CICO: Cleanup environment"
  make docker-rm
  echo "CICO: Exiting with $EXIT_CODE"
}

function deploy() {
  export REGISTRY="quay.io"
  export PLUGINS='work-in-progress test-keeper pr-sanitizer'

  if [ "${TARGET}" = "rhel" ]; then
    export DEPLOY_DOCKERFILE='Dockerfile.deploy.rhel'
    export DOCKER_REPO="openshiftio/rhel-ike-prow-plugins"
  fi

  # Login first
  if [ -n "${IMG_REPO_USERNAME}" -a -n "${IMG_REPO_PASSWORD}" ]; then
    docker login -u ${IMG_REPO_USERNAME} -p ${IMG_REPO_PASSWORD} ${REGISTRY}
  else
    echo "Could not login, missing credentials for the registry"
  fi

  # compile, build and deploy the hook
  export TAG=$(echo ${GIT_COMMIT} | cut -c1-${DEVSHIFT_TAG_LEN})

  make docker-build-hook
  make deploy-hook

  # compile, build and deploy plugins
  make deploy-plugins

  echo 'CICO: Image pushed, ready to update deployed app'
}

function cico_setup() {
  load_jenkins_vars;
  install_deps;
  prepare;
}
