# Issue #1: CLI Definition Format

## Problem
External programs (bash scripts, Python, etc.) need a way to define their flag structure to warg. The format must be:
- Concise (minimal boilerplate)
- Bash-friendly (easy to write in shell scripts)
- Expressive enough to define hierarchical flags with types and descriptions

## Implications
- This is the **primary interface** for non-Go users
- Poor UX here means low adoption
- Format affects how warg is invoked in every script
- Must support infinite nesting depth
- Must distinguish between switch flags and value flags
- Must allow multiple names per flag (short/long forms)

## Requirements
1. Must be writable in bash without complex escaping
2. Must support hierarchical flag definitions
3. Must specify: names, type (switch/value), description, children
4. Should be compact for simple cases but scale to complex flag trees
5. Should have good error messages when definition is malformed

## Options to Evaluate

### Option A: Inline Flag Definition
\`\`\`bash
warg -D "n,name;string;User name" \
     -D "v,verbose;bool;Verbose output" \
     -D "G,git;context;Git operations" \
     -D "G.c,commit;bool;Commit changes" \
     -D "G.m,message;string;Commit message" \
     -- "$@"
\`\`\`
**Pros:** Self-contained, no extra files
**Cons:** Gets unwieldy with many flags, harder to see structure

### Option B: JSON/YAML Config File
\`\`\`json
{
  "flags": [
    {"names": ["-n", "--name"], "type": "string", "desc": "User name"},
    {"names": ["-G", "--git"], "type": "context", "desc": "Git operations",
     "children": [
       {"names": ["-c", "--commit"], "type": "bool", "desc": "Commit changes"},
       {"names": ["-m", "--message"], "type": "string", "desc": "Commit message"}
     ]}
  ]
}
\`\`\`
**Pros:** Clean structure, easy to see hierarchy
**Cons:** Requires extra file, more verbose, bash-unfriendly

### Option C: Heredoc DSL
\`\`\`bash
warg parse "$@" <<EOF
-n, --name       : string : User name
-v, --verbose    : bool   : Verbose output
-G, --git        : context : Git operations
  -c, --commit   : bool   : Commit changes
  -m, --message  : string : Commit message
EOF
\`\`\`
**Pros:** Readable, indentation shows hierarchy, bash-friendly
**Cons:** Custom parsing needed, syntax rules to define

### Option D: Environment Variable Array
\`\`\`bash
WARG_DEFS=(
  "n,name:string:User name"
  "v,verbose:bool:Verbose output"
  "G,git:context:Git operations"
  "G.c,commit:bool:Commit changes"
  "G.m,message:string:Commit message"
)
warg parse "${WARG_DEFS[@]}" -- "$@"
\`\`\`
**Pros:** Bash native, fairly concise
**Cons:** Dotted notation for hierarchy is implicit, not visual

## Decision Needed
Which format(s) to support? Multiple formats for different use cases?
