FROM alpine:latest

RUN true \
    && wget -O "/etc/apk/keys/apk@k-moeller.dk-63d10688.rsa.pub.rsa.pub" "https://kalledk.github.io/registry/apk/apk@k-moeller.dk-63d10688.rsa.pub.rsa.pub" \
    && echo "https://kalledk.github.io/registry/apk" >> /etc/apk/repositorie \
    && apk --no-cache add acmednsproxy-tools acmednsproxy \
    && mkdir -p /etc/acmednsproxy \
    && cp /usr/share/doc/acmednsproxy/examples/*.yaml /etc/acmednsproxy/

VOLUME ["/etc/acmednsproxy"]

CMD ["acmednsproxy", "serve", "-c", "/etc/acmednsproxy/acmednsproxy.yaml"]

EXPOSE 8080