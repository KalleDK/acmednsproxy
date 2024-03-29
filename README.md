# acmednsproxy
Proxy for DNS01 challenges

## Install
```
su
apk update
apk add sudo
addgroup sudo
addgroup sudo km
# relogin

# https://kalledk.github.io/apk/
# Install key
sudo wget -O "/etc/apk/keys/apk@k-moeller.dk-63d10688.rsa.pub" "https://kalledk.github.io/apk/apk@k-moeller.dk-63d10688.rsa.pub"

# Install repo
echo "https://kalledk.github.io/apk" | sudo tee -a /etc/apk/repositories

# Install packages
sudo apk update
sudo apk add acmednsproxy-tool acmednsproxy-openrc acmednsproxy

# Install completion
sudo apk add bash-completion bash shadow
chsh -s /bin/bash $USER
adpcrypt completion bash | sudo tee /usr/share/bash-completion/completions/adpcrypt
acmednsproxy completion bash | sudo tee /usr/share/bash-completion/completions/acmednsproxy

# Change motd
echo -e "Add Provider\n\n$ sudo vi /etc/acmednsproxy/acmednsproxy.yaml\n" | sudo tee -a /etc/motd
echo -e "Add User\n\n$ sudo adpcrypt add -a /etc/acmednsproxy/auth.yaml\n" | sudo tee -a /etc/motd

# Add auth
sudo adpcrypt add -a /etc/acmednsproxy/auth.yaml

# Add provider
sudo vi /etc/acmednsproxy/acmednsproxy.yaml
provider:
  type: multi
  config:
    - type: cloudflare
      domain: example.com
      config:
        zoneid: ZpRCkagvcx6PuNC6b5a17gMtfnbv156Ozq7tsRfhk
        authtoken: 7dZhgfbonk5j5hgFTQ_xYIrhgfrKx
    - type: cloudflare
      domain: sub.example.com
      config:
        zoneid: Zpe7Lvcx9jRCkag6PuNCx5a17gMtf156Ozq7tsRfhk
        authtoken: 7dZbonhgfArJFDl3FTQ_xYIhgfhgfhrrKx


# Starting service
sudo rc-update add acmednsproxy
sudo rc-service acmednsproxy start
sudo apk add lego
sudo adpcrypt add -u ns01 -d $(hostname -f) -a /etc/acmednsproxy/auth.yaml

sudo su -l
lego --path=/var/lib/lego --email="acme@example.com" --domains="$(hostname -f)" --dns="httpreq" run

cat /usr/sbin/on_certrenewal_acmednsproxy
#############
#!/usr/bin/env sh

cp /var/lib/lego/certificates/$(hostname -f).key /etc/acmednsproxy/server.key
chown root:acmednsproxy /etc/acmednsproxy/server.key
chmod g+r /etc/acmednsproxy/server.key
cp /var/lib/lego/certificates/$(hostname -f).crt /etc/acmednsproxy/server.crt
chown root:acmednsproxy /etc/acmednsproxy/server.crt
chmod g+r /etc/acmednsproxy/server.crt
service acmednsproxy restart
###########

cat /etc/periodic/daily/001-updatecert.sh
chmod +x /etc/periodic/daily/001-updatecert.sh
#############
#!/usr/bin/env sh

set -e

MAIL=acme@example.com

export HTTPREQ_MODE=RAW
export HTTPREQ_USERNAME=ns01
export HTTPREQ_PASSWORD=aykB2dZrIm87WTsApDuYEx45
HTTPREQ_ENDPOINT=https://ns01.example.com:9090

#!/bin/sh

first=0
testserver=""
days=45


while getopts "rtkf" o; do
    case "${o}" in
        k)
            HTTPREQ_ENDPOINT=http://ns01.example.com:8080
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
    esac
done
export HTTPREQ_ENDPOINT
echo $HTTPREQ_ENDPOINT
wget -q -O - $HTTPREQ_ENDPOINT/ping
if [ "first" = "$first" ]; then
echo lego $testserver --path=/var/lib/lego --email="$MAIL" --domains="$(hostname -f)" --dns="httpreq" run
lego $testserver --path=/var/lib/lego --email="$MAIL" --domains="$(hostname -f)" --dns="httpreq" run
else
echo lego $testserver --path=/var/lib/lego --email="$MAIL" --domains="$(hostname -f)" --dns="httpreq" renew --days=$days --renew-hook="/usr/sbin/on_certrenewal_acmednsproxy"
lego $testserver --path=/var/lib/lego --email="$MAIL" --domains="$(hostname -f)" --dns="httpreq" renew --days=$days --renew-hook="/usr/sbin/on_certrenewal_acmednsproxy"
fi;


#############


sudo vi /etc/acmednsproxy/acmednsproxy.yaml
certfile: /etc/acmednsproxy/server.crt
keyfile: /etc/acmednsproxy/server.key

```
