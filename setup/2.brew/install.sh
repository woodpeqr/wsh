#!/bin/bash
set -u

pkgs=""
while IFS= read -r pkg; do
    [ -n "$pkg" ] && pkgs="$pkgs $pkg"
done
[ -z "$pkgs" ] && exit 0

# shellcheck disable=SC2086
brew install $pkgs || {
    echo "ERROR: brew failed to install: $pkgs" >&2
    exit 1
}
