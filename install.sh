#!/bin/sh
# Installs the latest grepdiem release into ~/.local/bin, or $GREPDIEM_BIN.
set -eu

fail() {
	echo "grepdiem: $1" >&2
	exit 1
}

fetch() {
	if command -v curl >/dev/null 2>&1; then
		curl -fsSL -o "$2" "$1"
	elif command -v wget >/dev/null 2>&1; then
		wget -qO "$2" "$1"
	else
		fail "curl or wget is needed to download the release"
	fi
}

sha256() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | cut -d' ' -f1
	elif command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$1" | cut -d' ' -f1
	else
		fail "sha256sum or shasum is needed to check the download"
	fi
}

# The work is in a function so a download cut short runs nothing.
main() {
	base="${GREPDIEM_RELEASE_URL:-https://github.com/Matter121/grepdiem/releases/latest/download}"
	bin_dir="${GREPDIEM_BIN:-$HOME/.local/bin}"

	case "$(uname -s)" in
	Linux) os=linux ;;
	Darwin) os=darwin ;;
	*) fail "no prebuilt binary for $(uname -s); see the README for other ways to install" ;;
	esac
	case "$(uname -m)" in
	x86_64 | amd64) arch=amd64 ;;
	arm64 | aarch64) arch=arm64 ;;
	*) fail "no prebuilt binary for $(uname -m); see the README for other ways to install" ;;
	esac

	archive="grepdiem_${os}_${arch}.tar.gz"
	tmp="$(mktemp -d)"
	trap 'rm -rf "$tmp"' EXIT

	fetch "$base/$archive" "$tmp/$archive" || fail "could not download $base/$archive"
	fetch "$base/checksums.txt" "$tmp/checksums.txt" || fail "could not download $base/checksums.txt"

	want="$(awk -v f="$archive" '$2 == f { print $1 }' "$tmp/checksums.txt")"
	[ -n "$want" ] || fail "checksums.txt does not list $archive"
	[ "$(sha256 "$tmp/$archive")" = "$want" ] || fail "$archive does not match its checksum"

	tar -xzf "$tmp/$archive" -C "$tmp"
	mkdir -p "$bin_dir"
	# Moved into place so a grepdiem that is running keeps its old file.
	cp "$tmp/grepdiem" "$bin_dir/.grepdiem.new"
	chmod 755 "$bin_dir/.grepdiem.new"
	mv "$bin_dir/.grepdiem.new" "$bin_dir/grepdiem"

	echo "Installed $("$bin_dir/grepdiem" --version) to $bin_dir/grepdiem"
	case ":$PATH:" in
	*":$bin_dir:"*) ;;
	*) echo "$bin_dir is not on your PATH; add it to run grepdiem by name." ;;
	esac
}

main
