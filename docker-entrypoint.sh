#!/bin/sh
set -eu

config_path="${XDNS_CONFIG:-/data/server_config.toml}"
chown -R xdns:xdns /data
if [ ! -f "$config_path" ]; then
    cp /opt/XDNS/server_config.toml.simple "$config_path"
	chown xdns:xdns "$config_path"
fi

if [ -n "${XDNS_DOMAIN:-}" ]; then
    escaped_domain=$(printf '%s' "$XDNS_DOMAIN" | sed 's/[\\&|]/\\&/g')
    sed -i "s|^DOMAIN = .*|DOMAIN = [\"${escaped_domain}\"]|" "$config_path"
fi

if grep -q 'DOMAIN = \["v.domain.com"\]' "$config_path"; then
    echo "XDNS_DOMAIN must be set for the first start, or edit $config_path" >&2
    exit 2
fi

exec su-exec xdns:xdns /usr/local/bin/XDNS-server \
    -config "$config_path" \
    -metrics-address "${XDNS_METRICS_ADDRESS:-0.0.0.0:9090}" \
    "$@"
