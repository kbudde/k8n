FROM --platform=$BUILDPLATFORM alpine:latest@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6 AS builder
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

FROM --platform=$BUILDPLATFORM ubuntu:latest@sha256:da6fc2be547864451aa253836dd926da33623312df4a9a243e35dc877c378a78
COPY --from=builder /sbin/tini-static   /bin/tini
COPY --from=ytt   /tmp/ytt  /bin
COPY --from=kapp  /tmp/kapp /bin

USER 5000
COPY k8n /bin/k8n
WORKDIR /config
ENTRYPOINT [ "/bin/tini", "--" , "/bin/k8n" ]
