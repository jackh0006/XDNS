#!/usr/bin/env sh
# XDNS client quick installer for Linux and Termux.
set -eu

REPOSITORY="jackh0006/XDNS"

log() { printf '%s\n' "[XDNS] $*"; }
die() { printf '%s\n' "[XDNS] Error: $*" >&2; exit 1; }

is_termux=false
if [ "$(uname -o 2>/dev/null || true)" = "Android" ] || [ -n "${TERMUX_VERSION:-}" ]; then
    is_termux=true
fi

default_dir="$HOME/.local/share/XDNS"
if [ "$is_termux" = false ] && [ "$(id -u)" -eq 0 ]; then
    default_dir="/opt/XDNS-client"
fi
INSTALL_DIR="${XDNS_CLIENT_DIR:-$default_dir}"

case "$(uname -m)" in
    x86_64|amd64) arch="AMD64" ;;
    i386|i486|i586|i686|x86) arch="X86" ;;
    aarch64|arm64) arch="ARM64" ;;
    armv7l|armv7|armhf) arch="ARMV7" ;;
    armv6l|armv6) arch="ARMV6" ;;
    armv5l|armv5) arch="ARMV5" ;;
    mips) arch="MIPS" ;;
    mipsel|mipsle) arch="MIPSLE" ;;
    mips64) arch="MIPS64" ;;
    mips64el|mips64le) arch="MIPS64LE" ;;
    riscv64) arch="RISCV64" ;;
    *) die "Unsupported architecture: $(uname -m)" ;;
esac

command -v tar >/dev/null 2>&1 || die "tar is required"
if command -v curl >/dev/null 2>&1; then
    fetch() { curl -fL --retry 3 -o "$1" "$2"; }
elif command -v wget >/dev/null 2>&1; then
    fetch() { wget -O "$1" "$2"; }
else
    die "Install curl or wget first"
fi

tmpdir="$(mktemp -d)"
cleanup() { rm -rf "$tmpdir"; }
trap cleanup EXIT HUP INT TERM

asset="XDNS_Client_Linux_${arch}.tar.gz"
url="https://github.com/${REPOSITORY}/releases/latest/download/${asset}"
log "Downloading ${asset}"
fetch "$tmpdir/client.tar.gz" "$url" || die "Download failed. Check the release page and your network connection."
tar -xzf "$tmpdir/client.tar.gz" -C "$tmpdir" || die "Could not unpack client archive"

binary="$(find "$tmpdir" -maxdepth 1 -type f -name "XDNS_Client_Linux_${arch}_*" | head -n 1)"
[ -n "$binary" ] && [ -f "$binary" ] || die "Client executable was not present in the downloaded archive"

mkdir -p "$INSTALL_DIR"
if [ -f "$INSTALL_DIR/client_config.toml" ]; then
    cp "$INSTALL_DIR/client_config.toml" "$INSTALL_DIR/client_config.toml.backup"
fi
if [ -f "$INSTALL_DIR/client_resolvers.txt" ]; then
    cp "$INSTALL_DIR/client_resolvers.txt" "$INSTALL_DIR/client_resolvers.txt.backup"
fi
cp "$binary" "$INSTALL_DIR/XDNS-client"
chmod 0755 "$INSTALL_DIR/XDNS-client"
for file in client_config.toml client_resolvers.txt client_config.*.toml CONFIG_PRESETS.md; do
    [ -f "$tmpdir/$file" ] || continue
    case "$file" in
        client_config.toml|client_resolvers.txt)
            [ -f "$INSTALL_DIR/$file" ] || cp "$tmpdir/$file" "$INSTALL_DIR/$file"
            ;;
        *) cp "$tmpdir/$file" "$INSTALL_DIR/$file" ;;
    esac
done

log "Installed XDNS client for Linux ${arch} in ${INSTALL_DIR}"
if [ "$is_termux" = true ]; then
    log "Termux detected. Android battery management may suspend background processes."
fi
printf '%s\n' "Next: edit ${INSTALL_DIR}/client_config.toml and ${INSTALL_DIR}/client_resolvers.txt"
printf '%s\n' "Run:  ${INSTALL_DIR}/XDNS-client --config ${INSTALL_DIR}/client_config.toml"
