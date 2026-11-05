FROM --platform=$BUILDPLATFORM alpine:latest@sha256:ce9e1e6d23f04a0b7593445a062a9fc73ed52f8260b9396f77a958295ad18ee7 AS builder
WORKDIR /tmp
RUN apk add curl git tini-static
RUN curl https://zyedidia.github.io/eget.sh | sh

# https://github.com/carvel-dev/ytt/releases
FROM builder AS ytt
ARG YTT_VERSION=0.49.0
RUN ./eget carvel-dev/ytt -t v${YTT_VERSION}

# https://github.com/carvel-dev/kapp/releases
FROM builder AS kapp
ARG KAPP_VERSION=0.62.0
RUN ./eget carvel-dev/kapp -t v${KAPP_VERSION}

FROM --platform=$BUILDPLATFORM ubuntu:latest@sha256:6a06111c3b0a8600feadb270aec771dc239bf407b25625f93f44ab00e11f4d3c
COPY --from=builder /sbin/tini-static   /bin/tini
COPY --from=ytt   /tmp/ytt  /bin
COPY --from=kapp  /tmp/kapp /bin

USER 5000
COPY k8n /bin/k8n
WORKDIR /config
ENTRYPOINT [ "/bin/tini", "--" , "/bin/k8n" ]
