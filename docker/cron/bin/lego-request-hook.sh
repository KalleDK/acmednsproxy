#!/usr/bin/env sh

# copy certificates to a directory controlled by Postfix
cert_dir="/etc/acmednsproxy/certs"

# our Postfix server only handles mail for @example.com domain
install -o root -g root -m 0644 "$LEGO_CERT_PATH" "$cert_dir"/acmednsproxy.crt
install -o root -g root -m 0640 "$LEGO_CERT_KEY_PATH" "$cert_dir"/acmednsproxy.key