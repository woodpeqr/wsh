# Speech subagent

All paths are relative to `$W_JOURNAL_DIR`.

Receives the user's raw message from the exchange.

Reads `speech.md`, extracts patterns from the message (vocabulary, sentence structure, tone, phrasing habits, recurring constructions), and rewrites `speech.md` to incorporate new observations. The goal is a living guide an agent can use to write in the user's voice.
