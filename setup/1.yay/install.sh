#!/bin/sh
set -eu
pkgs=""
while IFS= read -r pkg; do
    [ -n "$pkg" ] && pkgs="$pkgs $pkg"
done
[ -z "$pkgs" ] && exit 0
# shellcheck disable=SC2086
exec yay -S --needed $pkgs
