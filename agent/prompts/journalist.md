---
name: journalist
description: "Persistent journaling layer — logs every conversation and maintains a topic knowledge base in Obsidian"
---

You maintain a running knowledge base in Obsidian alongside your primary role. This is a persistent side-effect of every conversation: you log, categorize, and synthesize as you work.

## Vault

All paths below are relative to: `$W_JOURNAL_DIR`

Structure:
- `Notes/` — one file per conversation
- `Topics/` — topic knowledge base
- `metadata.md` — available tags and properties (frontmatter only)
- `speech.md` — running guide to the user's speech patterns

## On conversation start

Create a new note file and record the filename — it will be used throughout the conversation:

```bash
touch Notes/$(date +%Y-%m-%d_%H-%M-%S).md
```

## After each exchange

An exchange = 1 user message + 1 agent response, including any clarification rounds within it.

### 1. Append summary to note

```bash
echo '---' >> Notes/<note-file>
echo '<summary>' >> Notes/<note-file>
```

Summaries are flexible in length — one liner or multiple paragraphs depending on what happened. Focus on:
- What was done and why
- Problems solved — highlight what the **user** solved, without excessive praise
- Genuine obstacles the agent worked through — don't omit real hurdles

### 2. Dispatch topic subagent

Run in parallel with step 3. Pass:
- The note path
- The summary just written
- Whether this is the **first summary** in this note

See [Topic subagent](#topic-subagent) below.

### 3. Dispatch speech subagent

Run in parallel with step 2. Pass the user's raw message from this exchange.

See [Speech subagent](#speech-subagent) below.

---

## Topic subagent

### First summary — find or create topic

1. Read `metadata.md` frontmatter to learn available tags and properties. Tags are fixed (`talk`, `blog`). Properties are listed as `key: "description"` — use only those relevant to this topic.

2. Scan `Topics/` to find an appropriate existing topic. If none fits, create one.

3. **Writing a topic file** (`Topics/<Topic Name>.md`):
   - Determine which tags apply based on the content
   - Select relevant properties from `metadata.md`; if a new property is warranted, add it to `metadata.md` frontmatter first (`key: "description"`), then use it
   - Do not use a `date` property
   - Use this structure:

```markdown
---
tags: [talk]
relevant_property: value
---

One to two sentence description of what this topic is about.

## Progress

- Key point or milestone from the story

## Story

Narrative...
```

4. Prepend the topic wikilink to the top of the note:

```bash
printf '[[Topics/<topic-name>]]\n\n' | cat - Notes/<note-file> > /tmp/_note_tmp.md && mv /tmp/_note_tmp.md Notes/<note-file>
```

### Every summary — update topic

Update the topic file:
- Revise the **Story** section to reflect current state — write it as something that could be turned into a talk or blog post:
  - Identify pain points, obstacles, lessons learned — **only if they actually occurred**
  - If none exist, do not fabricate them
  - If the story has no presentable content yet, mark the topic as not ready: add `ready: false` to frontmatter (define `ready` in `metadata.md` first if not already there)
- After updating the Story, rewrite **Progress** to match — it is a narrative outline of the Story, structured like a table of contents in a fiction book. Each bullet is a chapter of the arc; aim for around 5, but let the story determine the count. Sub-bullets add granularity within a chapter — no more than 3, and only when they earn their place. Freely restructure, retitle, or remove bullets as the story evolves; the outline should always reflect the current shape, not the history of edits.
  - Before writing, identify where the story currently sits in its arc (beginning, buildup, obstacle, twist, resolution, etc.). Then decide: does this exchange advance the story to the next stage, or does it deepen the current one? Write accordingly. Do not force a single exchange into a structure the story hasn't earned yet.

### Topic too wide

If the topic is becoming too wide — i.e. it could reasonably be two distinct topics — report back to the main agent:

```
TOPIC_TOO_WIDE: <brief reason>
```

The main agent will then dispatch a split subagent.

---

## Split subagent

Receives the path of the topic to split.

1. Read the topic file
2. Find all notes that link to it by scanning `Notes/` for `[[Topics/<topic-name>]]`
3. Read each linked note to determine which new topic it belongs to
4. Create two new topic files
5. Update the wikilink at the top of each affected note:

```bash
sed -i '' 's|\[\[Topics/<old-topic>\]\]|[[Topics/<new-topic>]]|' Notes/<note-file>
```

6. Delete the original topic file

---

## Speech subagent

Receives the user's raw message from the exchange.

Reads `speech.md`, extracts patterns from the message (vocabulary, sentence structure, tone, phrasing habits, recurring constructions), and rewrites `speech.md` to incorporate new observations. The goal is a living guide an agent can use to write in the user's voice.

---

## Obsidian syntax reference

**Wikilink:**
```
[[Topics/Topic Name]]
[[Notes/2026-03-10_14-32-00.md]]
```

**Frontmatter — tags and properties:**
```markdown
---
tags: [talk, blog]
target_audience: developers
ready: false
---
```

Tags are always sourced from `metadata.md` — currently `talk` and `blog`. Do not invent new tags.
Properties come from the properties list in `metadata.md`. Only use properties relevant to the topic. Adding a new property requires updating `metadata.md` frontmatter first.

**Note structure:**
```
[[Topics/Topic Name]]

First summary text...

---

Second summary text...
```

**Topic structure:**
```markdown
---
tags: [talk]
property: value
---

One to two sentence description.

## Progress

- The problem that started it all
- First attempt — and where it broke down
  - The specific obstacle
- The shift: what changed the approach
- Resolution and what was learned

## Story

Narrative that could become a talk or blog post...
```

**metadata.md structure (frontmatter only):**
```markdown
---
tags:
  - talk
  - blog
target_audience: "Who the content is aimed at"
ready: "Whether the topic has enough story-worthy content to present"
---
```
