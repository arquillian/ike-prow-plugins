#!/bin/sh

LIGHT_GREEN='\033[1;32m'
GREEN='\033[0;32m'
CLEAR='\033[0m'
echo -e "${CLEAR}${LIGHT_GREEN}Install prerequisites${CLEAR}"

# Fedora
if [ -n "$(command -v dnf)" ]; then
    sudo dnf -y install curl git mercurial make binutils bison gcc glibc-devel
fi

# Ubuntu
if [ -n "$(command -v apt-get)" ]; then
    sudo apt-get -y install curl git mercurial make binutils bison gcc build-essential
fi

GO_VERSION=go1.9.4

echo -e "${CLEAR}${LIGHT_GREEN}Installing GVM${CLEAR}"
curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer | sh
[[ -s "$GVM_ROOT/scripts/gvm" ]] && source $GVM_ROOT/scripts/gvm || source ~/scripts/gvm # otherwise gvm use will not be recognized

echo -e "${CLEAR}${LIGHT_GREEN}Installing Go 1.4 first - see ${CLEAR}https://github.com/moovweb/gvm/issues/124${CLEAR}"
gvm install go1.4 --binary
gvm use go1.4

echo -e "${CLEAR}${LIGHT_GREEN}Installing target Go version ${GO_VERSION}${CLEAR}"
gvm install ${GO_VERSION}
gvm use ${GO_VERSION}

echo -e "${CLEAR}${LIGHT_GREEN}Installing required go packages${CLEAR}"
make tools

echo -e "${CLEAR}${LIGHT_GREEN}Installing glide and project dependencies${CLEAR}"
curl https://glide.sh/get | sh
make install

echo -e "${CLEAR}${GREEN}Don't forget to extend your ${CLEAR}\$GOPATH${GREEN} to contain your workspace, e.g.: \n${CLEAR}${LIGHT_GREEN}export GOPATH=\$GOPATH:~/code/golang${CLEAR}"

