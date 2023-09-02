FROM alpine:latest AS build
RUN apk add --no-cache tini-static

FROM scratch
COPY --from=build /sbin/tini-static /bin/tini

USER 5000
COPY k8n /bin/k8n
WORKDIR /config
ENTRYPOINT [ "/bin/tini", "--" , "/bin/k8n" ]