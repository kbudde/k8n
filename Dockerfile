FROM alpine:latest@sha256:7144f7bab3d4c2648d7e59409f15ec52a18006a128c733fcff20d3a4a54ba44a AS builder
WORKDIR /tmp
RUN apk add curl git tini-static 
RUN curl https://zyedidia.github.io/eget.sh | sh 

# https://github.com/carvel-dev/ytt/releases
FROM builder AS ytt
ARG YTT_VERSION=0.46.2
RUN ./eget carvel-dev/ytt -t v${YTT_VERSION}

# https://github.com/carvel-dev/kapp/releases
FROM builder AS kapp
ARG KAPP_VERSION=0.59.1
RUN ./eget carvel-dev/kapp -t v${KAPP_VERSION}

FROM ubuntu@sha256:8eab65df33a6de2844c9aefd19efe8ddb87b7df5e9185a4ab73af936225685bb
COPY --from=builder /sbin/tini-static   /bin/tini
COPY --from=ytt   /tmp/ytt  /bin
COPY --from=kapp  /tmp/kapp /bin

USER 5000
COPY k8n /bin/k8n
WORKDIR /config
ENTRYPOINT [ "/bin/tini", "--" , "/bin/k8n" ]