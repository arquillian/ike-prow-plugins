#!/bin/sh

LIGHT_GREEN='\033[1;32m'
GREEN='\033[0;32m'
CLEAR='\033[0m'
if [ -n "$(which sudo 2>/dev/null)" ]; then
    SUDO=sudo
fi

function install_packages() {
  echo -e "${CLEAR}${LIGHT_GREEN}Install prerequisites${CLEAR}"

  # Fedora
  if [ -n "$(command -v dnf)" ]; then
    ${SUDO} dnf -y install gcc git wget make which
  fi

  # Ubuntu
  if [ -n "$(command -v apt-get)" ]; then
    ${SUDO} apt-get -y install gcc git wget make which
  fi
}

function install_go_bins() {
  GO_VERSION=go1.10.2
  GO_LOCATION=/usr/local

  if [ -z $(which go 2>/dev/null) ]; then
    echo -e "${CLEAR}${LIGHT_GREEN}Installing target Go version ${GO_VERSION}${CLEAR} to ${GO_LOCATION}"
    ${SUDO} wget -P /tmp --no-verbose https://dl.google.com/go/${GO_VERSION}.linux-amd64.tar.gz \
      && echo "15b0937615809f87321a457bb1265f946f9f6e736c563d6c5e0bd2c22e44f779  /tmp/${GO_VERSION}.linux-amd64.tar.gz" > /tmp/go-bin-checksum \
      && sha256sum -c /tmp/go-bin-checksum \
      && tar -C ${GO_LOCATION} -xzf /tmp/${GO_VERSION}.linux-amd64.tar.gz \
      && rm -f /tmp/${GO_VERSION}.linux-amd64.tar.gz

      if [ -z ${GOPATH} ]; then
        export GOPATH=${HOME}/go-workspace
      fi
      mkdir -p ${GOPATH}

      GO_BIN_LOCATION=${GO_LOCATION}/go/bin
      export PATH=$PATH:${GO_BIN_LOCATION}:${GOPATH}/bin

      echo -e "${CLEAR}${GREEN}Go has been installed to ${GO_LOCATION} and ${CLEAR}\$GOPATH${GREEN} variable set to ${GOPATH}." \
      "Don't forget to add the Go binary directory along with ${GOPATH}/bin to your ${CLEAR}\$PATH${GREEN}: \n${CLEAR}" \
      "${LIGHT_GREEN}export PATH=\$PATH:${GO_BIN_LOCATION}:${GOPATH}/bin" &&
      echo -e "${CLEAR}${GREEN}You can also extend your ${CLEAR}\$GOPATH${GREEN} to contain your workspace, e.g.: \n${CLEAR}" \
      "${LIGHT_GREEN}export GOPATH=\$GOPATH:~/code/golang${CLEAR}"
  fi

  if [ -z $(which glide 2>/dev/null) ]; then
    echo -e "${CLEAR}${LIGHT_GREEN}Installing glide${CLEAR}"
    curl https://glide.sh/get | sh
  fi
}

if [[ $1 != "--only-go-bins" ]]; then
  install_packages
fi
install_go_bins

if [[ $1 != "--only-go-bins" ]]; then
  echo -e "${CLEAR}${LIGHT_GREEN}Installing required go packages${CLEAR}"
  make tools
  echo -e "${CLEAR}${LIGHT_GREEN}Installing project dependencies${CLEAR}"
  make install
fi