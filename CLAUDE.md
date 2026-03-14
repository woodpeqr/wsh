# wsh

Shell utility script (`w.sh`) with modular time tracking, git credential helper, agent launcher, and shell initialization.

## Completions

The file `completions/_w.sh` is the zsh completion script for `w.sh`. **It must be kept in sync with `w.sh` at all times.**

Whenever any of the following change in `w.sh` or its sourced modules (`_time.sh`, `_agent.sh`, `_utils.sh`):
- A flag is added, removed, or renamed
- A flag's argument type changes
- A new context (`-I`, `-T`, `-G`, `-A`, `-S`) is added or removed
- A sub-flag moves between contexts

...then `completions/_w.sh` must be updated to match before committing.

## Env vars

| Variable | Required | Used by | Notes |
|---|---|---|---|
| `SOLIDTIME_API_KEY` | yes | `-T` time commands | Bearer token |
| `SOLIDTIME_URL` | yes | `-T` time commands | API base URL |
| `SOLIDTIME_ORG` | yes | `-T` time commands | Organization name |
| `SOLIDTIME_PROJECT` | yes | `-T` time commands | Project name (default: Work) |
| `GIT_CREDS` | yes (for `-Gc`) | `-G -c get` | Format: `user=token` |
| `NVM_DIR` | no | `-I -n` | Defaults to `$HOME/.nvm` |
| `ZSH_PLUGIN_DIRS` | no | `-I -z` | Colon-separated list of dirs containing `.zsh` plugin files |
| `W_JOURNAL_DIR` | no | `-A` agent session hooks | Defaults to `$HOME/Documents/journal` |
