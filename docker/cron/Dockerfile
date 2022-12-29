FROM alpine:latest

RUN apk add --no-cache lego curl

COPY --chown=root:root --chmod=755 bin/*.sh /usr/local/bin/
COPY --chown=root:root --chmod=755 001-updatecert.sh /etc/periodic/daily/001-updatecert.sh

VOLUME ["/var/lib/lego", "/etc/acmednsproxy"]

CMD [ "crond", "-f", "-d", "8" ]