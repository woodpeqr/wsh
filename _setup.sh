# FLAGS (set by w.sh during parsing)
setup=
setup_simulate=
setup_delete=
setup_restow=
setup_list=
setup_packages=
setup_shell=

# ARGS
setup_pkg_args=()

# Execute a .sh pkg script and collect output into a named array; skip blank lines and comments
_setup_read_pkgs() {
    local file="$1" arr_name="$2"
    [[ ! -f "$file" ]] && return
    while IFS= read -r line; do
        [[ -z "$line" || "$line" == \#* ]] && continue
        eval "${arr_name}+=(\"\$line\")"
    done < <(bash "$file")
}

# Find a .sh pkg file; outputs path or nothing
_setup_find_pkgfile() {
    local base="$1"
    [[ -f "${base}.sh" ]] && echo "${base}.sh" || true
}

# Resolve a command name to PM-specific name via pkg_map.sh; empty = skip
_setup_resolve_pkg() {
    local map_sh="$1" pkg="$2"
    if [[ -f "$map_sh" && -x "$map_sh" ]]; then
        "$map_sh" "$pkg"
    else
        echo "$pkg"
    fi
}

# ── public functions ──────────────────────────────────────────────────────────

setup_do_init() {
    log "Initializing dotfiles submodule..."
    git -C "$wsh_dir" submodule update --init
}

setup_do_stow() {
    [[ ! -d "$dotfiles_dir" ]] && {
        error "dotfiles dir not found: $dotfiles_dir — run 'w -S' first"
        exit 1
    }

    local -a cmd=(stow --dir="$dotfiles_dir" --target="$HOME")
    if [[ -n $setup_stow ]]; then
        cmd+=(--stow)
    elif [[ -n $setup_delete ]]; then
        cmd+=(--delete)
    elif [[ -n $setup_restow ]]; then
        cmd+=(--restow)
    fi

    [[ -n $setup_simulate ]] && cmd+=(--simulate)

    cmd+=("${setup_pkg_args[@]}")

    debug "running: ${cmd[*]}"
    "${cmd[@]}"
}

_is_stow_linked() {
    local path="$1"
    while [[ "$path" != "$HOME" && "$path" != "/" ]]; do
        [[ -L "$path" ]] && return 0
        path="${path%/*}"
    done
    return 1
}

setup_do_list() {
    [[ ! -d "$dotfiles_dir" ]] && {
        error "dotfiles dir not found: $dotfiles_dir"
        exit 1
    }

    printf "%-24s %s\n" "PACKAGE" "STATUS"
    printf "%-24s %s\n" "-------" "------"

    for pkg_dir in "$dotfiles_dir"/*/; do
        [[ ! -d "$pkg_dir" ]] && continue
        local pkg
        pkg=$(basename "$pkg_dir")
        local all_ok=1 any=0

        # Recursively check leaf files; corresponding path in $HOME must be a symlink
        while IFS= read -r -d '' f; do
            any=1
            local rel="${f#$pkg_dir}"
            _is_stow_linked "$HOME/$rel" || {
                all_ok=0
                break
            }
        done < <(find "$pkg_dir" -type f -print0)

        local icon
        if [[ $any -eq 0 ]]; then
            icon="  (empty)"
        elif [[ $all_ok -eq 1 ]]; then
            icon="  ✓"
        else
            icon="  ✗"
        fi

        printf "%-24s %s\n" "$pkg" "$icon"
    done
}

setup_do_packages() {
    local setup_dir="$wsh_dir/setup"
    [[ ! -d "$setup_dir" ]] && {
        error "setup dir not found: $setup_dir"
        exit 1
    }

    # Iterate PM dirs in sorted order (numeric prefix controls order)
    local -a pm_dirs=()
    while IFS= read -r -d '' d; do
        pm_dirs+=("$d")
    done < <(find "$setup_dir" -maxdepth 1 -mindepth 1 -type d -print0 | sort -z)

    for pm_dir in "${pm_dirs[@]}"; do
        local pm_name
        pm_name=$(basename "$pm_dir")
        pm_name="${pm_name#*.}" # "1.yay" → "yay"

        command -v "$pm_name" &>/dev/null || {
            debug "skipping $pm_name: not in PATH"
            continue
        }
        log "--- $pm_name ---"

        local -a prepend_raw=() base_raw=() append_raw=()

        local pkgs_onlyf
        pkgs_onlyf=$(_setup_find_pkgfile "$pm_dir/pkgs_only")
        if [[ -n "$pkgs_onlyf" ]]; then
            _setup_read_pkgs "$pkgs_onlyf" base_raw
        else
            local base_pkgsf
            base_pkgsf=$(_setup_find_pkgfile "$setup_dir/pkgs")
            local pkgs_prependf
            pkgs_prependf=$(_setup_find_pkgfile "$pm_dir/pkgs_prepend")
            local pkgs_appendf
            pkgs_appendf=$(_setup_find_pkgfile "$pm_dir/pkgs_append")
            [[ -n "$pkgs_prependf" ]] && _setup_read_pkgs "$pkgs_prependf" prepend_raw
            [[ -n "$base_pkgsf" ]] && _setup_read_pkgs "$base_pkgsf" base_raw
            [[ -n "$pkgs_appendf" ]] && _setup_read_pkgs "$pkgs_appendf" append_raw
        fi

        local map_sh="$pm_dir/pkg_map.sh"
        local -a prepend_resolved=() base_resolved=() append_resolved=()
        for p in "${prepend_raw[@]}"; do
            local m
            m=$(_setup_resolve_pkg "$map_sh" "$p")
            [[ -n "$m" ]] && prepend_resolved+=("$m")
        done
        for p in "${base_raw[@]}"; do
            local m
            m=$(_setup_resolve_pkg "$map_sh" "$p")
            [[ -n "$m" ]] && base_resolved+=("$m")
        done
        for p in "${append_raw[@]}"; do
            local m
            m=$(_setup_resolve_pkg "$map_sh" "$p")
            [[ -n "$m" ]] && append_resolved+=("$m")
        done

        if [[ ${#prepend_resolved[@]} -eq 0 && ${#base_resolved[@]} -eq 0 && ${#append_resolved[@]} -eq 0 ]]; then
            log "  no packages, skipping"
            continue
        fi

        local install_sh="$pm_dir/install.sh"
        [[ ! -f "$install_sh" ]] && {
            error "install.sh not found: $install_sh"
            continue
        }
        [[ ! -x "$install_sh" ]] && {
            error "install.sh not executable: $install_sh"
            continue
        }

        debug "→ prepend:${#prepend_resolved[@]} base:${#base_resolved[@]} append:${#append_resolved[@]}"
        WSH_PKGS_PREPEND="${prepend_resolved[*]:-}" \
            WSH_PKGS_BASE="${base_resolved[*]:-}" \
            WSH_PKGS_APPEND="${append_resolved[*]:-}" \
            "$install_sh" || return $?
    done
}

_setup_has_env_var() {
    grep -qE "^[#[:space:]]*(export )?$1=" "$HOME/.zshenv" 2>/dev/null
}

_setup_stub_env_var() {
    local var="$1" desc="$2" default_val="$3"
    _setup_has_env_var "$var" && return
    printf '\n# %s\n# export %s=%s\n' "$desc" "$var" "$default_val" >> "$HOME/.zshenv"
}

setup_do_shell() {
    if grep -q 'WSH_SETUP_DONE=1' "$HOME/.zshenv" 2>/dev/null; then
        log "shell setup already done (WSH_SETUP_DONE=1 in ~/.zshenv) — skipping"
        return
    fi

    printf 'export PATH="%s:$PATH"\n' "$wsh_dir" >> "$HOME/.zshenv"
    printf 'eval "$(%s/w.sh -IA)"\n' "$wsh_dir" >> "$HOME/.zshrc"

    # ZSH_PLUGIN_DIRS: write sensible defaults if not already set
    if ! _setup_has_env_var "ZSH_PLUGIN_DIRS"; then
        printf '\n# Colon-separated list of dirs containing .zsh plugin files (loaded by w.sh -Iz)\nexport ZSH_PLUGIN_DIRS="/usr/share/zsh/plugins:$HOME/.zsh/plugins:$HOME/.local/share/zsh/plugins"\n' >> "$HOME/.zshenv"
    fi

    # Required env var stubs
    _setup_stub_env_var "SOLIDTIME_API_KEY"  "Required: Solidtime API bearer token (w.sh -T commands)" ""
    _setup_stub_env_var "SOLIDTIME_URL"      "Required: Solidtime API base URL, e.g. https://app.solidtime.io (w.sh -T)" ""
    _setup_stub_env_var "SOLIDTIME_ORG"      "Required: Solidtime organization name (w.sh -T)" ""
    _setup_stub_env_var "SOLIDTIME_PROJECT"  "Required: Solidtime project name (w.sh -T)" "Work"
    _setup_stub_env_var "GIT_CREDS"          "Required for w.sh -Gc: git credentials, format user=token" ""
    # Optional env var stubs
    _setup_stub_env_var "NVM_DIR"            "Optional: NVM directory, defaults to ~/.nvm (w.sh -In)" ""
    _setup_stub_env_var "W_JOURNAL_DIR"      "Optional: journal dir for agent session hooks, defaults to ~/Documents/journal (w.sh -A)" ""

    printf '\nexport WSH_SETUP_DONE=1\n' >> "$HOME/.zshenv"
    export WSH_SETUP_DONE=1

    log "wrote ~/.zshenv and ~/.zshrc"
}
