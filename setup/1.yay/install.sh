#!/bin/bash
set -eu

pkgs=""
while IFS= read -r pkg; do
    [ -n "$pkg" ] && pkgs="$pkgs $pkg"
done
[ -z "$pkgs" ] && exit 0

# shellcheck disable=SC2086
yay -S --needed $pkgs || {
    echo "ERROR: yay failed to install: $pkgs" >&2
    exit 1
}
