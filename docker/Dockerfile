FROM alpine:latest

RUN true \
    && wget -O "/etc/apk/keys/alpine@k-moeller.dk-62068d1b.rsa.pub" "https://kalledk.github.io/acmednsproxy/apk/alpine@k-moeller.dk-62068d1b.rsa.pub" \
    && echo "https://kalledk.github.io/acmednsproxy/apk" >> /etc/apk/repositories \
    && apk --no-cache add acmednsproxy-tools acmednsproxy \
    && mkdir -p /etc/acmednsproxy \
    && cp /usr/share/doc/acmednsproxy/examples/*.yaml /etc/acmednsproxy/

VOLUME ["/etc/acmednsproxy"]

CMD ["acmednsproxy", "serve", "-c", "/etc/acmednsproxy/acmednsproxy.yaml"]

EXPOSE 8080