#!/bin/sh

set -eu

repository="kpb/beanstalk"
install_dir="${INSTALL_DIR:-${HOME:?HOME must be set}/.local/bin}"

fail() {
	printf '%s\n' "error: $*" >&2
	exit 1
}

if [ -n "${VERSION:-}" ]; then
	version="$VERSION"
else
	version_url=$(curl --fail --location --silent --show-error --output /dev/null --write-out '%{url_effective}' "https://github.com/$repository/releases/latest") || fail "finding the latest release"
	version=${version_url##*/}
fi

case "$version" in
	v*) ;;
	*) fail "invalid release version: $version" ;;
esac

case "$(uname -s)" in
	Linux) os="linux" ;;
	Darwin) os="darwin" ;;
	*) fail "unsupported operating system: $(uname -s)" ;;
esac

case "$(uname -m)" in
	x86_64 | amd64) arch="amd64" ;;
	aarch64 | arm64) arch="arm64" ;;
	*) fail "unsupported architecture: $(uname -m)" ;;
esac

release_version=${version#v}
archive="beanstalk_${release_version}_${os}_${arch}.tar.gz"
release_url="https://github.com/$repository/releases/download/$version"
temporary_directory=$(mktemp -d)
trap 'rm -rf "$temporary_directory"' 0 HUP INT TERM

curl --fail --location --silent --show-error --output "$temporary_directory/$archive" "$release_url/$archive" || fail "downloading $archive"
curl --fail --location --silent --show-error --output "$temporary_directory/checksums.txt" "$release_url/checksums.txt" || fail "downloading checksums"

expected_checksum=$(awk -v archive="$archive" '$2 == archive { print $1; exit }' "$temporary_directory/checksums.txt")
[ -n "$expected_checksum" ] || fail "checksum for $archive was not found"

if command -v sha256sum >/dev/null 2>&1; then
	actual_checksum=$(sha256sum "$temporary_directory/$archive" | awk '{ print $1 }')
elif command -v shasum >/dev/null 2>&1; then
	actual_checksum=$(shasum -a 256 "$temporary_directory/$archive" | awk '{ print $1 }')
else
	fail "sha256sum or shasum is required to verify the download"
fi

[ "$actual_checksum" = "$expected_checksum" ] || fail "checksum verification failed for $archive"

tar -xzf "$temporary_directory/$archive" -C "$temporary_directory"
[ -f "$temporary_directory/beanstalk" ] || fail "archive does not contain the beanstalk binary"

mkdir -p "$install_dir"
install -m 0755 "$temporary_directory/beanstalk" "$install_dir/beanstalk"
printf 'Installed beanstalk %s to %s/beanstalk\n' "$version" "$install_dir"
