#!/usr/bin/env sh

lego --server $LEGO_SERVER --path=/var/lib/lego --email="$LEGO_MAIL" --domains="$LEGO_DOMAIN" --dns="$LEGO_DNS" renew --renew-hook="/usr/local/bin/lego-renew-hook.sh"