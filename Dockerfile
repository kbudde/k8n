FROM --platform=$BUILDPLATFORM alpine:latest@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf AS builder
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

FROM --platform=$BUILDPLATFORM ubuntu:latest@sha256:065aa8626841f16057b78b027b464cd59c4fac36c204c72dcbf48d66f5fa8a96
COPY --from=builder /sbin/tini-static   /bin/tini
COPY --from=ytt   /tmp/ytt  /bin
COPY --from=kapp  /tmp/kapp /bin

USER 5000
COPY k8n /bin/k8n
WORKDIR /config
ENTRYPOINT [ "/bin/tini", "--" , "/bin/k8n" ]
