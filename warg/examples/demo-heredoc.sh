#!/usr/bin/env bash
# Example: Using heredoc DSL format for flag definitions

cd "$(dirname "$0")/.." || exit 1

./warg <<'EOF' -- -v --name "Bob"
-v, --verbose Enable verbose output
-n, --name [value] User name
-o, --output [value] Output file
EOF

echo ""
echo "---"
echo ""

# More complex example with context:
./warg <<'EOF' -- -Gcm "Fix bug"
-v, --verbose Enable verbose output
-G, --git Git operations
  -c, --commit Commit changes
  -m, --message [msg] Commit message
  -p, --push Push changes
EOF
