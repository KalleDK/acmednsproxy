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
      --system \
      --disabled-login \
      --ingroup acmednsproxy \
      --no-create-home \
      --home /nonexistent \
      --gecos "AcmeDNSProxy Server" \
      --shell /bin/false \
      acmednsproxy  >/dev/null
fi
