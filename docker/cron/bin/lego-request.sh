#!/usr/bin/env sh

lego --server $LEGO_SERVER --path=/var/lib/lego --email="$LEGO_MAIL" --domains="$LEGO_DOMAIN" --dns="$LEGO_DNS" -a run --run-hook="/usr/local/bin/lego-request-hook.sh"