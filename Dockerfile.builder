FROM centos:7
LABEL maintainer="Devtools <devtools@redhat.com>"
LABEL maintainer="Devtools-test <devtools-test@redhat.com>"

ARG HOME
ARG GOPATH
ARG PROJECT_PATH
ARG GO_VERSION=go1.9.4

RUN yum install -y \
    gcc \
    git \
    wget \
    make \
    which

ENV LANG=en_US.utf8

RUN cd /tmp \
    && mkdir -p ${GOPATH}/{src,bin} \
    && mkdir -p ${PROJECT_PATH}

ENV PATH=${PATH}:${GOPATH}/bin:/usr/local/go/bin

ENV GIT_COMMITTER_NAME=alien-ike
ENV GIT_COMMITTER_EMAIL=arquillian-team@lists.jboss.org
RUN git config --global user.name "${GIT_COMMITTER_NAME}"
RUN git config --global user.email "${GIT_COMMITTER_EMAIL}"

ADD setup.sh /setup.sh
RUN /setup.sh --only-go-bins

RUN chmod a+rwx ${HOME}
RUN chmod -R a+rwx ${GOPATH}

ENTRYPOINT ["/bin/bash"]
