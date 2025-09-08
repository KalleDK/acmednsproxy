#!/usr/bin/env sh

set -e

MAIL=acme@example.com

export HTTPREQ_MODE=RAW
export HTTPREQ_USERNAME=
export HTTPREQ_PASSWORD=
HTTPREQ_ENDPOINT=https://ns03.example.com:9090

first=0
testserver=""
days=45


while getopts "rtkf" o; do
    case "${o}" in
        k)
            HTTPREQ_ENDPOINT=http://ns03.example.com:8080
            ;;
        r)
            first="first"
            ;;
        t)
            testserver=" --server=https://acme-staging-v02.api.letsencrypt.org/directory "
            ;;
        f)
            days=90
            ;;
        *) echo "Usage: $0 [-r] [-t] [-k] [-f]"
           exit 1
           ;;
    esac
done
export HTTPREQ_ENDPOINT
echo $HTTPREQ_ENDPOINT
wget -q -O - $HTTPREQ_ENDPOINT/ping
if [ "first" = "$first" ]; then
echo lego "$testserver" --path=/var/lib/lego/nas04.ilo --email="$MAIL" --accept-tos --domains="nas04.ilo.example.com" --dns="httpreq" --key-type="rsa2048" run --run-hook="ipmi-certupdater -c /root/nas04.ilo/ipmiconfig.yml"
lego "$testserver" --path=/var/lib/lego/nas04.ilo --email="$MAIL" --accept-tos --domains="nas04.ilo.example.com" --dns="httpreq" --key-type="rsa2048" run --run-hook="ipmi-certupdater -c /root/nas04.ilo/ipmiconfig.yml"
else
echo lego "$testserver" --path=/var/lib/lego/nas04.ilo --email="$MAIL" --domains="nas04.ilo.example.com" --dns="httpreq" --key-type=rsa2048 renew --days=$days --renew-hook="ipmi-certupdater -c /root/nas04.ilo/ipmiconfig.yml"
lego "$testserver" --path=/var/lib/lego/nas04.ilo --email="$MAIL" --domains="nas04.ilo.example.com" --dns="httpreq" --key-type=rsa2048 renew --days=$days --renew-hook="ipmi-certupdater -c /root/nas04.ilo/ipmiconfig.yml"
fi;