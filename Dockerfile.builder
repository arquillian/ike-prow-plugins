FROM centos:7
LABEL maintainer="Devtools <devtools@redhat.com>"
LABEL maintainer="Devtools-test <devtools-test@redhat.com>"

RUN yum install -y \
    gcc \
    git \
    wget \
    make \
    which

ENV LANG=en_US.utf8

ARG HOME
ARG GOPATH
ARG PROJECT_PATH

ENV HOME=${HOME}
ENV GOPATH=${GOPATH}
ENV GO_VERSION=go1.9.4

RUN cd /tmp \
    && mkdir -p ${GOPATH}/src \
    && mkdir -p ${GOPATH}/bin \
    && echo ${GOPATH} \
    && echo ${PROJECT_PATH} \
    && mkdir -p ${PROJECT_PATH}

ENV PATH=${PATH}:${GOPATH}/bin:/usr/local/go/bin

RUN git config --global user.email "arquillian-team@lists.jboss.org"
RUN git config --global user.name "alien-ike"
ENV GIT_COMMITTER_NAME=alien-ike
ENV GIT_COMMITTER_EMAIL=arquillian-team@lists.jboss.org

ADD setup.sh /setup.sh
RUN /setup.sh --only-go-bins

RUN chmod a+rwx ${HOME}
RUN chmod -R a+rwx ${GOPATH}

ENTRYPOINT ["/bin/bash"]
