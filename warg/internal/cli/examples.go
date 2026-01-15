package cli

import "fmt"

// Example demonstrates all three CLI definition formats

func ExampleInlineFormat() {
	// Define flags using inline format
	defs := []string{
		"n,name;string;User name",
		"v,verbose;bool;Verbose output",
		"G,git;context;Git operations",
		"G.c,commit;bool;Commit changes",
		"G.m,message;string;Commit message",
	}

	flags, err := ParseInlineDefinitions(defs)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Parsed %d root flags\n", len(flags))
	// Output: Parsed 3 root flags
}

func ExampleJSONFormat() {
	jsonData := `{
		"flags": [
			{
				"names": ["-n", "--name"],
				"type": "string",
				"desc": "User name"
			},
			{
				"names": ["-G", "--git"],
				"type": "context",
				"desc": "Git operations",
				"children": [
					{
						"names": ["-c", "--commit"],
						"type": "bool",
						"desc": "Commit changes"
					}
				]
			}
		]
	}`

	flags, err := ParseJSONDefinitions([]byte(jsonData))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Parsed %d flags\n", len(flags))
	// Output: Parsed 2 flags
}

func ExampleHeredocFormat() {
	input := `-n, --name       : string : User name
-v, --verbose    : bool   : Verbose output
-G, --git        : context : Git operations
  -c, --commit   : bool   : Commit changes
  -m, --message  : string : Commit message`

	flags, err := ParseHeredocDefinition(input)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Parsed %d root flags\n", len(flags))
	// Output: Parsed 3 root flags
}

// Usage examples for bash scripts

func BashExampleInline() string {
	return `#!/bin/bash
# Example: Using inline definitions

warg parse \
  -D "n,name;string;User name" \
  -D "v,verbose;bool;Verbose output" \
  -D "G,git;context;Git operations" \
  -D "G.c,commit;bool;Commit changes" \
  -D "G.m,message;string;Commit message" \
  -- "$@"
`
}

func BashExampleJSON() string {
	return `#!/bin/bash
# Example: Using JSON config file

cat > flags.json <<'EOF'
{
  "flags": [
    {"names": ["-n", "--name"], "type": "string", "desc": "User name"},
    {"names": ["-v", "--verbose"], "type": "bool", "desc": "Verbose output"},
    {"names": ["-G", "--git"], "type": "context", "desc": "Git operations",
     "children": [
       {"names": ["-c", "--commit"], "type": "bool", "desc": "Commit changes"},
       {"names": ["-m", "--message"], "type": "string", "desc": "Commit message"}
     ]}
  ]
}
EOF

warg parse --config flags.json -- "$@"
`
}

func BashExampleHeredoc() string {
	return `#!/bin/bash
# Example: Using heredoc DSL

warg parse "$@" <<EOF
-n, --name       : string : User name
-v, --verbose    : bool   : Verbose output
-G, --git        : context : Git operations
  -c, --commit   : bool   : Commit changes
  -m, --message  : string : Commit message
EOF
`
}

func BashExampleEnvironment() string {
	return `#!/bin/bash
# Example: Using environment variables (alternative inline format)

export WARG_DEFS=(
  "n,name:string:User name"
  "v,verbose:bool:Verbose output"
  "G,git:context:Git operations"
  "G.c,commit:bool:Commit changes"
  "G.m,message:string:Commit message"
)

warg parse "${WARG_DEFS[@]}" -- "$@"
`
}
