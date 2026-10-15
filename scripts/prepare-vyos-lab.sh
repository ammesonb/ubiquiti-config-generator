#!/usr/bin/env bash
set -euo pipefail

baseline=/usr/local/share/ubq-lab/baseline.env
# shellcheck source=../integration/lab/baseline.env
source "$baseline"
archive=vyos-rootfs.tar.xz

# The cache contains only the prepared filesystem and its verification metadata.
if [[ -f "$archive" && -f "$archive.sha256" && -f baseline.used.env ]] &&
    cmp -s "$baseline" baseline.used.env && sha256sum --check "$archive.sha256"; then
    echo 'Using cached VyOS filesystem.'
    exit 0
fi

iso="vyos-${VYOS_VERSION}-generic-amd64.iso"
url="https://github.com/vyos/vyos-nightly-build/releases/download/${VYOS_VERSION}/${iso}"
if [[ ! -f "$iso" ]] || ! echo "${VYOS_ISO_SHA256}  ${iso}" | sha256sum --check --status; then
    curl --fail --location --retry 3 --output "${iso}.part" "$url"
    echo "${VYOS_ISO_SHA256}  ${iso}.part" | sha256sum --check
    mv "${iso}.part" "$iso"
fi
curl --fail --location --retry 3 --output iso-to-oci \
    "https://raw.githubusercontent.com/vyos/vyos-build/${VYOS_CONVERTER_REVISION}/scripts/iso-to-oci"
bash iso-to-oci "$iso"
mv "vyos-${VYOS_VERSION}-oci-amd64.tar.xz" "$archive"
sha256sum "$archive" > "$archive.sha256"
cp "$baseline" baseline.used.env
