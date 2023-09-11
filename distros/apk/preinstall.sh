#!/usr/bin/env sh

set -e

# creating acmednsproxy group if he isn't already there
if ! getent group acmednsproxy >/dev/null; then
     # Adding system group: acmednsproxy.
    addgroup --system acmednsproxy >/dev/null
fi

# creating acmednsproxy user if he isn't already there
if ! getent passwd acmednsproxy >/dev/null; then
    # Adding system user: acmednsproxy.
    adduser \
      -S \
      -G acmednsproxy \
      -H \
      -h /nonexistent \
      -g "AcmeDNSProxy Server" \
      -s /bin/false \
      -D \
      acmednsproxy  >/dev/null
fi

mkdir -p /var/log/acmednsproxy
chown acmednsproxy:acmednsproxy /var/log/acmednsproxy