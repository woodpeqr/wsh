#!/usr/bin/env bash
set -euo pipefail

[[ -z "${W_JOURNAL_DIR:-}" ]] && exit 0

NOTE=$(cat ~/.cache/journalist/current-note.txt 2>/dev/null || echo "")
[[ -z "$NOTE" ]] && exit 0

FULL_PATH="$W_JOURNAL_DIR/$NOTE"
[[ ! -f "$FULL_PATH" ]] && exit 0

CURRENT_SIZE=$(wc -c < "$FULL_PATH" | tr -d ' ')
LAST_SIZE=$(cat ~/.cache/journalist/last-size.txt 2>/dev/null || echo "0")

if [[ "$CURRENT_SIZE" -gt "$LAST_SIZE" ]]; then
    echo "$CURRENT_SIZE" > ~/.cache/journalist/last-size.txt
    exit 0
else
    echo "$CURRENT_SIZE" > ~/.cache/journalist/last-size.txt
    echo "Journaling: session note has not been updated this exchange. Use /journal to log it."
    exit 2
fi
