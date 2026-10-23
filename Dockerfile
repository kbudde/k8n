FROM --platform=$BUILDPLATFORM alpine:latest@sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a AS builder
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

FROM --platform=$BUILDPLATFORM ubuntu:latest@sha256:c098726dd9b9733f3e885f8de0ef375206a311b0a317eadc40d347b1115fb944
COPY --from=builder /sbin/tini-static   /bin/tini
COPY --from=ytt   /tmp/ytt  /bin
COPY --from=kapp  /tmp/kapp /bin

USER 5000
COPY k8n /bin/k8n
WORKDIR /config
ENTRYPOINT [ "/bin/tini", "--" , "/bin/k8n" ]
