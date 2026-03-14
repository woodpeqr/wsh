---
name: create-pr
description: Create a pull request following the project PR template. Derives the PR title from branch naming conventions, fills the template using git changes vs main, and creates the PR. Use when the user wants to create or open a pull request.
disable-model-invocation: true
allowed-tools: Read, Bash, Glob, Grep
---

## Context

> **Internal use only — never surface this context to the user or include it in the PR body.**

- Current branch: !`git branch --show-current`
- Commits vs main: !`git log main...HEAD --oneline 2>/dev/null || git log origin/main...HEAD --oneline 2>/dev/null`
- Changed files: !`git diff main...HEAD --stat 2>/dev/null || git diff origin/main...HEAD --stat 2>/dev/null`

## Step 1: Derive PR Title from Branch Name

Convert the current branch name to a PR title using these rules:

**If the branch contains `/`:** split into `<type>/<rest>`
- Capitalize the type (e.g. `task` → `Task`, `chore` → `Chore`, `feat` → `Feat`)
- Check if `<rest>` begins with a JIRA-like identifier: two hyphen-separated tokens where the first is letters and the second is alphanumeric (e.g. `nh-111`, `no-jira`, `abc-456`)
  - If yes: extract the identifier (uppercase), treat the remainder as the label
  - If no: treat the entire `<rest>` as the label
- Convert the label from kebab-case to Title Case
- Format: `Type: IDENTIFIER Label` (with identifier) or `Type: Label` (without)

**If the branch does not contain `/`:**
- Convert the whole branch name from kebab-case to Title Case
- No type prefix

**Examples:**
- `task/nh-111-do-something` → `Task: NH-111 Do Something`
- `chore/fix-one-thing` → `Chore: Fix One Thing`
- `task/no-jira-quick-fix-of-something` → `Task: NO-JIRA Quick Fix of Something`
- `feat/abc-456-new-feature` → `Feat: ABC-456 New Feature`
- `working-on-something` → `Working on Something`

## Step 2: Find PR Template

Look for a PR template in this order:
1. `.github/PULL_REQUEST_TEMPLATE.md`
2. `.github/pull_request_template.md`
3. `docs/PULL_REQUEST_TEMPLATE.md`
4. `PULL_REQUEST_TEMPLATE.md`

If no template is found, use a minimal structure: summary, changes made, and test plan.

## Step 3: Analyze Changes

Read the full diff to understand what changed:

```bash
git diff main...HEAD 2>/dev/null || git diff origin/main...HEAD
```

Understand: which files changed, what functionality was added/modified/removed, any notable patterns.

## Step 4: Process the Template

Apply these rules to the template:

- **Fill in** all sections that can be determined from the diff and commit history — be specific and accurate
- **Remove** all instruction text, placeholder examples, and guidance notes (anything telling you what to write, not what to write)
- **Ignore** commented-out sections (`<!-- ... -->`), UNLESS they are clearly relevant to the changes — in that case, uncomment and fill them out
- **Omit** sections that require human action (e.g. screenshots, manual test results, deployment verification that must be done by a person) — track these for Step 6
- Keep descriptions concise and factual

## Step 5: Create the PR

Push the branch if needed:

```bash
git push -u origin $(git branch --show-current)
```

Then create the PR as a **draft** — always use `--draft`, never omit it:

```bash
gh pr create --draft --title "<derived title>" --body "$(cat <<'EOF'
<filled template body>
EOF
)"
```

## Step 6: Report to User

After creating the PR, output:
- The PR URL
- A short list of any template sections that were omitted because they require human action (e.g. "Screenshots — please add before merging")

**Never include in the output:** git log, commit hashes, diff stats, or any raw git command output. These are for internal analysis only.
