# acmednsproxy
Proxy for DNS01 challenges

## Install
```
# https://kalledk.github.io/apk/
# Install key
sudo wget -O "/etc/apk/keys/apk@k-moeller.dk-63d10688.rsa.pub" "https://kalledk.github.io/apk/apk@k-moeller.dk-63d10688.rsa.pub"

# Install repo
echo "https://kalledk.github.io/apk" | sudo tee -a /etc/apk/repositories

# Install packages
apk update
apk add acmednsproxy-tool acmednsproxy-openrc acmednsproxy
rc-update add acmednsproxy
rc-service acmednsproxy start

# Change motd
echo -e "Add Provider\n\n$ sudo vi /etc/acmednsproxy/acmednsproxy.yaml\n" | sudo tee -a /etc/motd
echo -e "Add User\n\n$ sudo adpcrypt add -a /etc/acmednsproxy/auth.yaml\n" | sudo tee -a /etc/motd

# Add auth
adpcrypt add -a /etc/acmednsproxy/auth.yaml
```
