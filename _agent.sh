#!/usr/bin/env bash

export W_JOURNAL_DIR="${W_JOURNAL_DIR:-$HOME/Documents/journal}"

# FLAGS
agent_init=
agent_claude=
agent_copilot=

# ARGS
agent_tool=
agent_prompt_arg=
agent_extra_args=()

_agent_link_dir() {
    local src="$1" dst="$2" label="$3"
    [[ ! -d "$src" ]] && return
    if [[ -L "$dst" ]]; then
        log "  skip: $label already linked"
    elif [[ -e "$dst" ]]; then
        error "$label exists but is not a symlink, skipping"
    else
        ln -s "$src" "$dst"
        log "  + $label -> $src"
    fi
}

agent_do_init() {
    local targets_claude="$agent_claude"
    local targets_copilot="$agent_copilot"
    if [[ -z "$targets_claude" && -z "$targets_copilot" ]]; then
        targets_claude=1
        targets_copilot=1
    fi

    local any_done=

    if [[ -n "$targets_claude" ]]; then
        if [[ ! -d "$HOME/.claude" ]]; then
            log "info: ~/.claude does not exist, skipping Claude setup"
        else
            log "Setting up ~/.claude..."
            _agent_link_dir "$wsh_dir/agent/plugins" "$HOME/.claude/agents" "~/.claude/agents"
            _agent_link_dir "$wsh_dir/agent/skills"  "$HOME/.claude/skill"  "~/.claude/skill"
            any_done=1
        fi
    fi

    if [[ -n "$targets_copilot" ]]; then
        if [[ ! -d "$HOME/.copilot" ]]; then
            log "info: ~/.copilot does not exist, skipping Copilot setup"
        else
            log "Setting up ~/.copilot..."
            _agent_link_dir "$wsh_dir/agent/plugins" "$HOME/.copilot/agents" "~/.copilot/agents"
            _agent_link_dir "$wsh_dir/agent/skills"  "$HOME/.copilot/skill"  "~/.copilot/skill"
            any_done=1
        fi
    fi

    if [[ -z "$any_done" ]]; then
        error "none of the requested targets exist (~/.claude / ~/.copilot)"
        exit 1
    fi
}

agent_do_run() {
    local tool="${agent_tool:-claude}"
    if [[ -n "$agent_prompt_arg" ]]; then
        if [[ ! -f "$agent_prompt_arg" ]]; then
            error "prompt file not found: $agent_prompt_arg"
            exit 1
        fi
        "$tool" --system-prompt "$(cat "$agent_prompt_arg")" "${agent_extra_args[@]}"
    else
        "$tool" "${agent_extra_args[@]}"
    fi
}
