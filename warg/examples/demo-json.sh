#!/usr/bin/env bash
# Example: Using JSON for flag definitions

cd "$(dirname "$0")/.." || exit 1

# Method 1: JSON string as argument
./warg '[
  {
    "names": ["-v", "--verbose"],
    "switch": true,
    "desc": "Enable verbose output"
  },
  {
    "names": ["-n", "--name"],
    "switch": false,
    "desc": "User name"
  },
  {
    "names": ["-G", "--git"],
    "switch": true,
    "desc": "Git operations",
    "children": [
      {
        "names": ["-c", "--commit"],
        "switch": true,
        "desc": "Commit changes"
      },
      {
        "names": ["-m", "--message"],
        "switch": false,
        "desc": "Commit message"
      }
    ]
  }
]' -- -v -Gcm "Initial commit"

echo ""
echo "---"
echo ""

# Method 2: JSON from stdin
cat <<'EOF' | ./warg -- -v --name "Bob"
[
  {
    "names": ["-v", "--verbose"],
    "switch": true,
    "desc": "Enable verbose output"
  },
  {
    "names": ["-n", "--name"],
    "switch": false,
    "desc": "User name"
  }
]
EOF
