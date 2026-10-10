#!/bin/sh
# Builds the release archives and checksums.txt into dist/.
set -eu

version="${1:?usage: scripts/release.sh <version>}"
cd "$(dirname "$0")/.."

rm -rf dist
mkdir dist

for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64; do
	os="${target%/*}"
	arch="${target#*/}"
	name="grepdiem_${os}_${arch}"
	stage="dist/$name"
	mkdir "$stage"
	cp LICENSE README.md "$stage"

	exe=grepdiem
	if [ "$os" = windows ]; then
		exe=grepdiem.exe
	fi
	CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -trimpath \
		-ldflags="-s -w -X main.version=$version" -o "$stage/$exe" ./cmd/grepdiem

	if [ "$os" = windows ]; then
		(cd "$stage" && zip -q "../$name.zip" ./*)
	else
		tar -czf "dist/$name.tar.gz" -C "$stage" .
	fi
	rm -r "$stage"
done

(cd dist && sha256sum grepdiem_* >checksums.txt)
