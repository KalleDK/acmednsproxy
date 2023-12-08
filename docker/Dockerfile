FROM alpine:latest
LABEL org.opencontainers.image.source=https://github.com/KalleDK/acmednsproxy
LABEL org.opencontainers.image.description="ACME DNS Proxy"

ARG ADP_VERSION=0.1.52

COPY dist/acmednsproxy_${ADP_VERSION}_x86_64.apk /tmp/acmednsproxy.apk
COPY dist/acmednsproxy-tools_${ADP_VERSION}_x86_64.apk /tmp/acmednsproxy-tools.apk

RUN apk add --no-cache --allow-untrusted /tmp/acmednsproxy.apk /tmp/acmednsproxy-tools.apk
RUN mkdir -p /etc/acmednsproxy
RUN cp /usr/share/doc/acmednsproxy/examples/*.yaml /etc/acmednsproxy/

VOLUME ["/etc/acmednsproxy"]

CMD ["acmednsproxy", "serve", "-c", "/etc/acmednsproxy/acmednsproxy.yaml"]

EXPOSE 8080