#!/usr/bin/env sh

/usr/local/bin/lego-request-hook.sh

curl -X POST -k https://server:9090/reload