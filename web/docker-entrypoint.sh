#!/bin/sh
set -eu

# Bootstrap TLS certs for nginx. If /etc/nginx/ssl/fullchain.pem is missing,
# generate a self-signed cert with SANs covering the hostnames the operator
# is likely to hit. The Go server's SSL settings endpoints later manage the
# same files for regeneration and Let's Encrypt handoff.

SSL_DIR=/etc/nginx/ssl
CERT="${SSL_DIR}/fullchain.pem"
KEY="${SSL_DIR}/privkey.pem"
META="${SSL_DIR}/mode"

mkdir -p "${SSL_DIR}" "${SSL_DIR}/acme-webroot"

if [ ! -s "${CERT}" ] || [ ! -s "${KEY}" ]; then
    hostname_short="$(hostname -s 2>/dev/null || hostname 2>/dev/null || echo stack-manager)"
    hostname_fqdn="$(hostname -f 2>/dev/null || echo "${hostname_short}")"

    # HOST_URL is set in .env to the public URL operators visit (e.g.
    # https://docker02:8993). Prefer its hostname over the container's
    # random Docker-assigned hostname for the cert CN.
    host_from_url=""
    if [ -n "${HOST_URL:-}" ]; then
        host_from_url="$(printf '%s' "${HOST_URL}" | sed -E 's|^[a-zA-Z]+://||; s|/.*$||; s|:[0-9]+$||')"
    fi

    common_name="${SSL_CN:-${host_from_url:-${hostname_fqdn}}}"

    sans="DNS:${common_name}"
    if [ "${hostname_short}" != "${common_name}" ]; then
        sans="${sans},DNS:${hostname_short}"
    fi
    if [ "${hostname_fqdn}" != "${common_name}" ] && [ "${hostname_fqdn}" != "${hostname_short}" ]; then
        sans="${sans},DNS:${hostname_fqdn}"
    fi
    sans="${sans},DNS:localhost,IP:127.0.0.1"

    if [ -n "${SSL_EXTRA_SANS:-}" ]; then
        # Convert comma list to newline list and iterate line-by-line so we
        # never touch IFS (semgrep flags global IFS reassignment).
        printf '%s\n' "${SSL_EXTRA_SANS}" | tr ',' '\n' | while IFS= read -r extra; do
            extra="$(printf '%s' "${extra}" | tr -d ' ')"
            [ -z "${extra}" ] && continue
            case "${extra}" in
                *:*) printf 'raw:%s\n' "${extra}" ;;
                *[!0-9.]*) printf 'dns:%s\n' "${extra}" ;;
                *) printf 'ip:%s\n' "${extra}" ;;
            esac
        done > /tmp/sans-extra.txt
        while IFS= read -r line; do
            case "${line}" in
                raw:*) sans="${sans},${line#raw:}" ;;
                dns:*) sans="${sans},DNS:${line#dns:}" ;;
                ip:*)  sans="${sans},IP:${line#ip:}" ;;
            esac
        done < /tmp/sans-extra.txt
        rm -f /tmp/sans-extra.txt
    fi

    umask 077
    openssl req -x509 -nodes -newkey rsa:2048 \
        -days "${SSL_SELF_SIGNED_DAYS:-3650}" \
        -subj "/CN=${common_name}" \
        -addext "subjectAltName=${sans}" \
        -addext "keyUsage=digitalSignature,keyEncipherment" \
        -addext "extendedKeyUsage=serverAuth" \
        -keyout "${KEY}" -out "${CERT}"
    printf 'self-signed\n' > "${META}"
    echo "docker-entrypoint: generated self-signed TLS cert with SANs ${sans}"
fi

chmod 600 "${KEY}" 2>/dev/null || true
chmod 644 "${CERT}" 2>/dev/null || true

exec "$@"
