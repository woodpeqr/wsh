#!/usr/bin/env bash
# Comprehensive test demonstrating all warg parsing features
set -e

echo "=== Complex warg parsing test ==="
echo ""

# Test 1: Combined flags with context
echo "Test 1: Combined context and flags"
../warg <<'EOF' -- -v -Gcm "Initial commit"
-v, --verbose Verbose output
-G, --git Git operations
  -c, --commit Commit changes
  -m, --message [msg] Commit message
  -p, --push Push changes
EOF
echo ""

# Test 2: Context resolution (parent flag accessible from child context)
echo "Test 2: Context resolution - verbose flag from git context"
../warg <<'EOF' -- -G -v -c
-v, --verbose Verbose output
-G, --git Git operations
  -c, --commit Commit changes
EOF
echo ""

# Test 3: Long flags only
echo "Test 3: Long flags"
../warg <<'EOF' -- --verbose --name "Alice Smith"
--verbose Verbose output
--name [value] User name
EOF
echo ""

# Test 4: JSON format
echo "Test 4: JSON input format"
echo '[{"names":["-v","--verbose"],"switch":true,"desc":"Verbose"},{"names":["-n","--name"],"switch":false,"desc":"Name"}]' | \
  ../warg -- -v --name "test user"
echo ""

# Test 5: Inline format
echo "Test 5: Inline -A format"
../warg \
  -A -n v,verbose -s -d "Verbose" \
  -A -n n,name -d "Name" \
  -- -v --name "Bob"
echo ""

echo "✅ All complex tests passed!"
