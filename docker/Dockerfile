FROM alpine:latest
LABEL org.opencontainers.image.source=https://github.com/KalleDK/acmednsproxy
LABEL org.opencontainers.image.description="ACME DNS Proxy"
ARG ADP_VERSION=0.1.51
RUN true \
    && wget -O "/etc/apk/keys/apk@k-moeller.dk-63d10688.rsa.pub.rsa.pub" "https://kalledk.github.io/registry/apk/apk@k-moeller.dk-63d10688.rsa.pub.rsa.pub" \
    && echo "https://kalledk.github.io/registry/apk" >> /etc/apk/repositories
RUN true \
    && apk --no-cache add acmednsproxy-tools=${ADP_VERSION} acmednsproxy=${ADP_VERSION} \
    && mkdir -p /etc/acmednsproxy \
    && cp /usr/share/doc/acmednsproxy/examples/*.yaml /etc/acmednsproxy/

VOLUME ["/etc/acmednsproxy"]

CMD ["acmednsproxy", "serve", "-c", "/etc/acmednsproxy/acmednsproxy.yaml"]

EXPOSE 8080