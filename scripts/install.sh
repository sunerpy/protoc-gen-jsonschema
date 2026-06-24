#!/bin/sh
# protoc-gen-jsonschema one-liner installer (Linux / macOS).
#
#   curl -fsSL https://raw.githubusercontent.com/sunerpy/protoc-gen-jsonschema/main/scripts/install.sh | sh
#
# Env overrides:
#   PGJ_VERSION      pin a release (e.g. 0.0.8 or v0.0.8); default: latest
#   PGJ_INSTALL_DIR  install destination; default: $HOME/.local/bin
set -eu

REPO="sunerpy/protoc-gen-jsonschema"
BIN="protoc-gen-jsonschema"

err() {
	printf 'error: %s\n' "$1" >&2
	exit 1
}
info() { printf '%s\n' "$1" >&2; }

# Pick a downloader once.
if command -v curl >/dev/null 2>&1; then
	download() { curl -fsSL "$1" -o "$2"; }
	fetch() { curl -fsSL "$1"; }
elif command -v wget >/dev/null 2>&1; then
	download() { wget -qO "$2" "$1"; }
	fetch() { wget -qO - "$1"; }
else
	err "need curl or wget to download releases"
fi
command -v tar >/dev/null 2>&1 || err "need tar to extract the release archive"

# Detect OS. GoReleaser names archives with raw GOOS values.
os=$(uname -s)
case "$os" in
Linux) os_part="linux" ;;
Darwin) os_part="darwin" ;;
*) err "unsupported OS: $os (supported: Linux, Darwin)" ;;
esac

# Detect arch. GoReleaser names archives with raw GOARCH values.
arch=$(uname -m)
case "$arch" in
x86_64 | amd64) arch_part="amd64" ;;
arm64 | aarch64) arch_part="arm64" ;;
*) err "unsupported arch: $arch (supported: amd64, arm64)" ;;
esac

ext="tar.gz"

# Resolve version: env override or latest-release API.
if [ "${PGJ_VERSION:-}" != "" ]; then
	version=$(printf '%s' "$PGJ_VERSION" | sed 's/^v//')
else
	info "Resolving latest release..."
	api="https://api.github.com/repos/${REPO}/releases/latest"
	tag=$(fetch "$api" | grep -o '"tag_name"[ ]*:[ ]*"[^"]*"' | head -1 | sed 's/.*"tag_name"[ ]*:[ ]*"\([^"]*\)".*/\1/')
	[ "${tag:-}" != "" ] || err "could not resolve latest release tag from $api"
	version=$(printf '%s' "$tag" | sed 's/^v//')
fi

# Asset name matches .goreleaser.yaml: {ProjectName}_{Version}_{Os}_{Arch}.{ext}
asset="${BIN}_${version}_${os_part}_${arch_part}.${ext}"
url="https://github.com/${REPO}/releases/download/v${version}/${asset}"
install_dir="${PGJ_INSTALL_DIR:-$HOME/.local/bin}"

info "Installing ${BIN} v${version} (${os_part}/${arch_part})"
info "  from: ${url}"
info "  to:   ${install_dir}/${BIN}"

tmp=$(mktemp -d 2>/dev/null || mktemp -d -t "$BIN")
trap 'rm -rf "$tmp"' EXIT INT TERM

download "$url" "$tmp/$asset" || err "download failed: $url"
tar -xzf "$tmp/$asset" -C "$tmp" || err "failed to extract $asset"

mkdir -p "$install_dir"
mv "$tmp/$BIN" "$install_dir/$BIN" || err "failed to install to $install_dir"
chmod +x "$install_dir/$BIN"

info "Installed ${BIN} to ${install_dir}/${BIN}"

# PATH hint.
case ":$PATH:" in
*":$install_dir:"*) ;;
*) info "NOTE: $install_dir is not on your PATH. Add: export PATH=\"$install_dir:\$PATH\"" ;;
esac
