FROM centos:7
LABEL maintainer="Devtools <devtools@redhat.com>"
LABEL maintainer="Devtools-test <devtools-test@redhat.com>"

RUN yum install -y git
RUN yum install -y ca-certificates

ENV LANG=en_US.utf8
ARG BINARY=0

COPY ./bin/${BINARY} /binary
ENTRYPOINT ["/binary"]