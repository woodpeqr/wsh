#!/usr/bin/env bash
# Example: Using inline flag definitions with warg

# This demonstrates the inline format for defining flags using -A --add
# Format: -A -n names -s (if switch) -d "description"

cd "$(dirname "$0")/.." || exit 1

./warg \
  -A -n v,verbose -s -d "Enable verbose output" \
  -A -n n,name -d "User name" \
  -A -n o,output -d "Output file" \
  -- -v --name "Alice" --output "result.txt"

# Expected output (JSON format):
# {
#   "flags": [
#     {
#       "definition": {
#         "names": ["-v", "--verbose"],
#         "switch": true,
#         "desc": "Enable verbose output"
#       },
#       "present": true,
#       "value": "",
#       "children": []
#     },
#     {
#       "definition": {
#         "names": ["-n", "--name"],
#         "switch": false,
#         "desc": "User name"
#       },
#       "present": false,
#       "value": "Alice",
#       "children": []
#     },
#     {
#       "definition": {
#         "names": ["-o", "--output"],
#         "switch": false,
#         "desc": "Output file"
#       },
#       "present": false,
#       "value": "result.txt",
#       "children": []
#     }
#   ]
# }
