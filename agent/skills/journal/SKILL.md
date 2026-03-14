---
name: journal
description: Log this exchange to the session journal. Dispatch topic and speech subagents. Use after each agent response.
---

All paths are relative to `$W_JOURNAL_DIR`.
The session note path is printed at startup in the session context.

## Step 1 — Append summary to note

Append to the session note:

```bash
echo '---' >> Notes/<note-file>
echo '<summary>' >> Notes/<note-file>
```

Summaries are flexible in length — one liner or multiple paragraphs depending on what happened. Focus on:
- What was done and why
- Problems solved — highlight what the **user** solved, without excessive praise
- Genuine obstacles the agent worked through — don't omit real hurdles

## Step 2 — Dispatch topic subagent

Run in parallel with step 3. Pass:
- The note path
- The summary just written
- Whether this is the **first summary** in this note

On the first summary, the topic agent prepends a wikilink to the note file.

## Step 3 — Dispatch speech subagent

Run in parallel with step 2. Pass the user's raw message from this exchange.
