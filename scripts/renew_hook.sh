#!/usr/bin/env sh

set -exuo pipefail

MAIL=acme@example.com

NS_USERNAME=user
NS_PASSWORD=password
NS_SERVER=ns03.example.com
NS_ENDPOINT=https://$NS_SERVER:9090

HOOK="ipmi-certupdater -c /root/nas04.ilo/ipmiconfig.yml"

TARGET_NAME=nas04.ilo
TARGET_DOMAIN=nas04.ilo.example.com

first=0
acme_server=https://acme-v02.api.letsencrypt.org/directory
days=45


while getopts "rtkf" o; do
    case "${o}" in
        k)
            NS_ENDPOINT=http://$NS_SERVER:8080
            ;;
        r)
            first="first"
            ;;
        t)
            acme_server=https://acme-staging-v02.api.letsencrypt.org/directory
            ;;
        f)
            days=90
            ;;
        *) echo "Usage: $0 [-r] [-t] [-k] [-f]"
           exit 1
           ;;
    esac
done
export HTTPREQ_MODE=RAW
export HTTPREQ_USERNAME="${NS_USERNAME}"
export HTTPREQ_PASSWORD="${NS_PASSWORD}"
export HTTPREQ_ENDPOINT="${NS_ENDPOINT}"
echo "${HTTPREQ_ENDPOINT}"

LEGO="lego --server=$acme_server --path=/var/lib/lego/$TARGET_NAME --email=$MAIL --dns=httpreq --key-type=rsa2048"

wget -q -O - $HTTPREQ_ENDPOINT/ping
curl --request POST --url $HTTPREQ_ENDPOINT/domain --user "$HTTPREQ_USERNAME:$HTTPREQ_PASSWORD" --data '{"domain": "nas04.ilo.krypto.dk"}'
if [ "first" = "$first" ]; then
$LEGO --accept-tos --domains="$TARGET_DOMAIN" run --run-hook="$HOOK"
else
$LEGO --domains="$TARGET_DOMAIN" renew --days=$days --renew-hook="$HOOK"
fi;