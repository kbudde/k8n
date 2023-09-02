FROM alpine:latest@sha256:7144f7bab3d4c2648d7e59409f15ec52a18006a128c733fcff20d3a4a54ba44a AS build
RUN apk add --no-cache tini-static

FROM scratch
COPY --from=build /sbin/tini-static /bin/tini

USER 5000
COPY k8n /bin/k8n
WORKDIR /config
ENTRYPOINT [ "/bin/tini", "--" , "/bin/k8n" ]