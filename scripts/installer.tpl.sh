#!/usr/bin/env sh
set -e

RELEASES_URL="https://github.com/ekristen/distillery/releases"
FILE_BASENAME="distillery"
LATEST="__VERSION__"

test -z "$VERSION" && VERSION="$LATEST"

test -z "$VERSION" && {
	echo "Unable to get distillery version." >&2
	exit 1
}

TMP_DIR="$(mktemp -d)"
# shellcheck disable=SC2064 # intentionally expands here
trap "rm -rf \"$TMP_DIR\"" EXIT INT TERM

OS="$(uname -s | awk '{print tolower($0)}')"
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64) ARCH="amd64" ;; # Normalize x86_64 to amd64
    aarch64) ARCH="arm64" ;; # Normalize aarch64 to arm64
esac
TAR_FILE="${FILE_BASENAME}-${VERSION}-${OS}-${ARCH}.tar.gz"

(
	cd "$TMP_DIR"
	echo "Downloading distillery $VERSION..."
	curl -sfLO "$RELEASES_URL/download/$VERSION/$TAR_FILE"
	curl -sfLO "$RELEASES_URL/download/$VERSION/checksums.txt"
	echo "Verifying checksums..."
	sha256sum --ignore-missing --quiet --check checksums.txt
	if command -v cosign >/dev/null 2>&1; then
		echo "Verifying signatures..."
		REF="refs/tags/$VERSION"
		cosign verify-blob \
			--certificate-identity-regexp "https://github.com/ekristen/distillery.*/.github/workflows/.*.yml@$REF" \
			--certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
			--cert "$RELEASES_URL/download/$VERSION/checksums.txt.pem" \
			--signature "$RELEASES_URL/download/$VERSION/checksums.txt.sig" \
			checksums.txt
	else
		echo "Could not verify signatures, cosign is not installed."
	fi
)

tar -xf "$TMP_DIR/$TAR_FILE" -C "$TMP_DIR"
"$TMP_DIR/dist" "install" "ekristen/distillery" "$@"
