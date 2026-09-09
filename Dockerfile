FROM --platform=$BUILDPLATFORM alpine:latest@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b AS builder
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

FROM --platform=$BUILDPLATFORM ubuntu:latest@sha256:513c074113a871b51a8d16ab445c88779d6452d937a164fb5cc479f32668a41d
COPY --from=builder /sbin/tini-static   /bin/tini
COPY --from=ytt   /tmp/ytt  /bin
COPY --from=kapp  /tmp/kapp /bin

USER 5000
COPY k8n /bin/k8n
WORKDIR /config
ENTRYPOINT [ "/bin/tini", "--" , "/bin/k8n" ]
