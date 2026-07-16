# acmednsproxy

**acmednsproxy** is a lightweight HTTP proxy for [ACME DNS-01 challenges](https://letsencrypt.org/docs/challenge-types/#dns-01-challenge).  
It sits between an ACME client (such as [lego](https://github.com/go-acme/lego)) and your actual DNS provider, adding authentication and multi-domain routing so you can:

- Issue certificates on machines that have **no direct access** to DNS provider credentials.
- Delegate specific domains to different Cloudflare zones or backend proxies.
- Rotate credentials without touching every machine that requests certificates.

```
ACME client (lego)
      │  HTTP Basic-Auth
      ▼
acmednsproxy  ──────►  Cloudflare API
                   └──►  another acmednsproxy  ──►  Cloudflare API
```

---

## Table of Contents

- [How it works](#how-it-works)
- [Installation](#installation)
- [Quick start](#quick-start)
- [Configuration](#configuration)
  - [Main config (acmednsproxy.yaml)](#main-config-acmednsproxyyaml)
  - [Auth file (auth.yaml)](#auth-file-authyaml)
  - [Provider file (provider.yaml)](#provider-file-provideryaml)
- [HTTP API](#http-api)
- [Managing users (adpcrypt)](#managing-users-adpcrypt)
- [Signal handling](#signal-handling)
- [Building from source](#building-from-source)

---

## How it works

1. An ACME client configures lego's `httpreq` DNS provider to point at this proxy.
2. When a certificate needs to be issued or renewed, lego sends a `POST /present` request with Basic-Auth credentials and the challenge domain + key.
3. acmednsproxy verifies the credentials against `auth.yaml`, finds the correct DNS backend in `provider.yaml`, and creates the TXT record.
4. After the ACME server verifies the challenge, lego sends `POST /cleanup` to remove the record.

---

## Installation

### Alpine Linux (from the project's APK repository)

```sh
# Add the signing key
sudo wget -O /etc/apk/keys/apk@k-moeller.dk-63d10688.rsa.pub \
  https://kalledk.github.io/apk/apk@k-moeller.dk-63d10688.rsa.pub

# Add the repository
echo "https://kalledk.github.io/apk" | sudo tee -a /etc/apk/repositories

# Install
sudo apk update
sudo apk add acmednsproxy acmednsproxy-tool acmednsproxy-openrc
```

### From source

```sh
go install github.com/KalleDK/acmednsproxy/cmd/acmednsproxy@latest
go install github.com/KalleDK/acmednsproxy/cmd/adpcrypt@latest
```

---

## Quick start

### 1. Create the auth file

```sh
adpcrypt init -a /etc/acmednsproxy/auth.yaml
adpcrypt add  -u myuser -d example.com -a /etc/acmednsproxy/auth.yaml
```

The `add` command prints the generated password to stdout:

```
HTTPREQ_MODE=RAW \
HTTPREQ_USERNAME=myuser \
HTTPREQ_PASSWORD=<generated-password>
```

Save these for use in your ACME client environment.

### 2. Create the provider file (`/etc/acmednsproxy/provider.yaml`)

```yaml
- type: cloudflare
  zones:
    example.com: <cloudflare-zone-id>
  authtoken: <cloudflare-api-token>
```

### 3. Create the main config (`/etc/acmednsproxy/acmednsproxy.yaml`)

```yaml
listen: ":9090"
tls:
  certfile: /etc/acmednsproxy/server.crt
  keyfile:  /etc/acmednsproxy/server.key
authenticator: /etc/acmednsproxy/auth.yaml
provider:      /etc/acmednsproxy/provider.yaml
```

Omit the `tls` block to run plain HTTP on port 8080.

### 4. Start the server

```sh
acmednsproxy serve -c /etc/acmednsproxy/acmednsproxy.yaml
```

### 5. Configure lego

```sh
export HTTPREQ_MODE=RAW
export HTTPREQ_USERNAME=myuser
export HTTPREQ_PASSWORD=<generated-password>
export HTTPREQ_ENDPOINT=https://<proxy-host>:9090

lego --email=you@example.com --domains=example.com --dns=httpreq run
```

---

## Configuration

### Main config (`acmednsproxy.yaml`)

| Key             | Type   | Default | Description |
|-----------------|--------|---------|-------------|
| `listen`        | string | `:9090` (TLS) / `:8080` (plain) | TCP address to listen on. |
| `tls.certfile`  | string | –       | Path to PEM certificate. Relative paths are resolved against the config file directory. |
| `tls.keyfile`   | string | –       | Path to PEM private key. |
| `authenticator` | string | –       | Path to the auth YAML file. |
| `provider`      | string | –       | Path to the provider YAML file. |

### Auth file (`auth.yaml`)

Two authenticator backends are available:

#### `simpleauth` – bcrypt-hashed username/password per domain

```yaml
type: simpleauth
permissions:
  example.com:
    alice: $2a$12$...   # bcrypt hash of alice's password
    bob:   $2a$12$...
  other.example.com:
    carol: $2a$12$...
```

Manage entries with the `adpcrypt` tool (see [Managing users](#managing-users-adpcrypt)).

#### `noauth` – no credentials required, domain allowlist only

```yaml
type: noauth
domains:
  - example.com
  - other.example.com
```

Any request for a listed domain is permitted regardless of credentials.  
Useful for testing or internal networks.

### Provider file (`provider.yaml`)

The file is a YAML list; each entry has a `type` field and backend-specific fields.

#### `cloudflare`

```yaml
- type: cloudflare
  # Map of domain name → Cloudflare zone ID.
  # The proxy can handle any subdomain of the listed domains.
  zones:
    example.com:     <zone-id-1>
    sub.example.com: <zone-id-2>   # more-specific zone takes precedence
  authtoken: <cloudflare-api-token>  # needs Zone/DNS/Edit permissions
  ttl: 120           # optional, seconds (minimum 120)
  httptimeout: 30    # optional, seconds
```

#### `httpreq` – forward to another acmednsproxy

```yaml
- type: httpreq
  endpoint: https://ns01.example.com:9090
  username: user
  password: secret
  httptimeout: 10    # optional, seconds
```

Multiple backends can be listed; the proxy selects the most-specific matching
domain for each request:

```yaml
- type: cloudflare
  zones:
    example.com: <zone-id>
  authtoken: <token>

- type: httpreq
  endpoint: https://internal-ns:9090
  username: user
  password: secret
```

---

## HTTP API

All mutating endpoints require HTTP Basic-Auth.

| Method | Path       | Auth | Description |
|--------|------------|------|-------------|
| GET    | `/ping`    | No   | Liveness probe. Returns current server time. |
| POST   | `/domain`  | Yes  | Verify that credentials are valid for a domain. |
| POST   | `/present` | Yes  | Create a DNS TXT challenge record. |
| POST   | `/cleanup` | Yes  | Remove a DNS TXT challenge record. |
| POST   | `/reload`  | No   | Reload config and TLS certificate from disk. |

### `POST /present` and `POST /cleanup`

Accepts either **default mode** (pre-computed FQDN and value, as sent by lego's httpreq provider in default mode):

```json
{ "fqdn": "_acme-challenge.example.com.", "value": "abc123" }
```

Or **RAW mode** (domain + keyAuth, as sent when `HTTPREQ_MODE=RAW`):

```json
{ "domain": "example.com", "token": "token", "keyAuth": "key.auth" }
```

### `POST /domain`

```json
{ "domain": "example.com" }
```

Returns `200 OK` with `{ "status": "ok", "domain": "...", "user": "..." }` if
the credentials grant access to the domain, or `401` otherwise.

### `POST /reload`

Triggers an in-place reload of `auth.yaml`, `provider.yaml`, and the TLS
certificate.  The server continues serving in-flight requests with the old
config while the new config is loaded.

> **Security note:** the `/reload` endpoint is unauthenticated.  Restrict
> access to trusted networks or localhost via firewall rules.

A reload can also be triggered by sending `SIGHUP` to the process.

---

## Managing users (`adpcrypt`)

`adpcrypt` manages entries in a `simpleauth` auth file.

```sh
# Initialise a new auth file
adpcrypt init [-a auth.yaml]

# Add a user (generates a random password by default)
adpcrypt add  -u alice -d example.com [-a auth.yaml]

# Add a user and prompt for a password
adpcrypt add  -u alice -d example.com -k [-a auth.yaml]

# Remove a user from a domain
adpcrypt del  -u alice -d example.com [-a auth.yaml]

# Verify credentials
adpcrypt verify -u alice -d example.com -k [-a auth.yaml]
```

The default auth file path is `auth.yaml` in the current directory.

---

## Signal handling

| Signal  | Effect |
|---------|--------|
| `SIGHUP` | Reload `auth.yaml`, `provider.yaml`, and the TLS certificate without dropping connections. |
| `SIGTERM` / `SIGINT` | Graceful shutdown. |

---

## Building from source

```sh
git clone https://github.com/KalleDK/acmednsproxy
cd acmednsproxy
go build ./cmd/acmednsproxy
go build ./cmd/adpcrypt
```

Run the tests:

```sh
go test ./...
```
