# Topic subagent

All paths are relative to `$W_JOURNAL_DIR`.

## First summary — find or create topic

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

## Every summary — update topic

Update the topic file:
- Revise the **Story** section to reflect current state — write it as something that could be turned into a talk or blog post:
  - Identify pain points, obstacles, lessons learned — **only if they actually occurred**
  - If none exist, do not fabricate them
  - If the story has no presentable content yet, mark the topic as not ready: add `ready: false` to frontmatter (define `ready` in `metadata.md` first if not already there)
- After updating the Story, rewrite **Progress** to match — it is a narrative outline of the Story, structured like a table of contents in a fiction book. Each bullet is a chapter of the arc; aim for around 5, but let the story determine the count. Sub-bullets add granularity within a chapter — no more than 3, and only when they earn their place. Freely restructure, retitle, or remove bullets as the story evolves; the outline should always reflect the current shape, not the history of edits.
  - Before writing, identify where the story currently sits in its arc (beginning, buildup, obstacle, twist, resolution, etc.). Then decide: does this exchange advance the story to the next stage, or does it deepen the current one? Write accordingly. Do not force a single exchange into a structure the story hasn't earned yet.

## Topic too wide

If the topic is becoming too wide — i.e. it could reasonably be two distinct topics — report back to the main agent:

```
TOPIC_TOO_WIDE: <brief reason>
```

The main agent will then dispatch a split subagent.

---

# Split subagent

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
