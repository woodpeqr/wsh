#!/usr/bin/env bash
set -euo pipefail

[[ -z "${W_JOURNAL_DIR:-}" ]] && exit 0

mkdir -p "$W_JOURNAL_DIR/Notes" ~/.cache/journalist
rm -f ~/.cache/journalist/last-size.txt

NOTE_FILE="Notes/$(date +%Y-%m-%d_%H-%M-%S).md"
touch "$W_JOURNAL_DIR/$NOTE_FILE"
echo "$NOTE_FILE" > ~/.cache/journalist/current-note.txt

echo "Journalist: session note created at $NOTE_FILE"
