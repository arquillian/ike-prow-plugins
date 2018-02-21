FROM gcr.io/k8s-prow/git:0.1
LABEL maintainer="Devtools <devtools@redhat.com>"
LABEL author="Bartosz Majsak <bartosz@redhat.com>"

ENV LANG=en_US.utf8
ARG PLUGIN_NAME=0

COPY ./bin/${PLUGIN_NAME} /plugin
ENTRYPOINT ["/plugin"]