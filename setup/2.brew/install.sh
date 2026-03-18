#!/bin/bash
set -eu

pkgs="${WSH_PKGS_PREPEND:-} ${WSH_PKGS_BASE:-} ${WSH_PKGS_APPEND:-}"
pkgs=$(echo "$pkgs" | tr '\n' ' ' | tr -s ' ' | sed 's/^ //;s/ $//')
[ -z "$pkgs" ] && exit 0

# shellcheck disable=SC2086
brew install $pkgs || {
    echo "ERROR: brew failed to install: $pkgs" >&2
    exit 1
}
