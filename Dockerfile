FROM alpine:latest@sha256:c5b1261d6d3e43071626931fc004f70149baeba2c8ec672bd4f27761f8e1ad6b AS builder
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

FROM ubuntu:latest@sha256:77906da86b60585ce12215807090eb327e7386c8fafb5402369e421f44eff17e
COPY --from=builder /sbin/tini-static   /bin/tini
COPY --from=ytt   /tmp/ytt  /bin
COPY --from=kapp  /tmp/kapp /bin

USER 5000
COPY k8n /bin/k8n
WORKDIR /config
ENTRYPOINT [ "/bin/tini", "--" , "/bin/k8n" ]