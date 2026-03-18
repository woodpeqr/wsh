# FLAGS (set by w.sh during parsing)
setup=
setup_simulate=
setup_delete=
setup_restow=
setup_list=
setup_packages=
setup_shell=

# ARGS
setup_dir_arg=
setup_pkg_args=()

# ── helpers ──────────────────────────────────────────────────────────────────

_setup_dotfiles_dir() {
    echo "${setup_dir_arg:-$wsh_dir/dotfiles}"
}

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
    local repo_root
    repo_root=$(git -C "$wsh_dir" rev-parse --show-toplevel 2>/dev/null) || {
        error "could not determine git repo root"
        exit 1
    }
    log "Initializing dotfiles submodule..."
    git -C "$repo_root" submodule update --init
}

setup_do_stow() {
    local dotfiles_dir
    dotfiles_dir=$(_setup_dotfiles_dir)
    [[ ! -d "$dotfiles_dir" ]] && {
        error "dotfiles dir not found: $dotfiles_dir — run 'w -S' first"
        exit 1
    }

    local -a cmd=(stow --dir="$dotfiles_dir" --target="$HOME")
    [[ -n $setup_simulate ]] && cmd+=(--simulate)
    [[ -n $setup_delete   ]] && cmd+=(--delete)
    [[ -n $setup_restow   ]] && cmd+=(--restow)

    if [[ ${#setup_pkg_args[@]} -gt 0 ]]; then
        cmd+=("${setup_pkg_args[@]}")
    else
        local -a all=()
        for d in "$dotfiles_dir"/*/; do
            [[ -d "$d" ]] && all+=("$(basename "$d")")
        done
        [[ ${#all[@]} -eq 0 ]] && { error "no packages found in $dotfiles_dir"; exit 1; }
        cmd+=("${all[@]}")
    fi

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
    local dotfiles_dir
    dotfiles_dir=$(_setup_dotfiles_dir)
    [[ ! -d "$dotfiles_dir" ]] && { error "dotfiles dir not found: $dotfiles_dir"; exit 1; }

    printf "%-24s %s\n" "PACKAGE" "STATUS"
    printf "%-24s %s\n" "-------" "------"

    for pkg_dir in "$dotfiles_dir"/*/; do
        [[ ! -d "$pkg_dir" ]] && continue
        local pkg; pkg=$(basename "$pkg_dir")
        local all_ok=1 any=0

        # Recursively check leaf files; corresponding path in $HOME must be a symlink
        while IFS= read -r -d '' f; do
            any=1
            local rel="${f#$pkg_dir}"
            _is_stow_linked "$HOME/$rel" || { all_ok=0; break; }
        done < <(find "$pkg_dir" -type f -print0)

        local icon
        if   [[ $any -eq 0    ]]; then icon="  (empty)"
        elif [[ $all_ok -eq 1 ]]; then icon="  ✓"
        else                            icon="  ✗"
        fi

        printf "%-24s %s\n" "$pkg" "$icon"
    done
}

setup_do_packages() {
    local setup_dir="$wsh_dir/setup"
    [[ ! -d "$setup_dir" ]] && { error "setup dir not found: $setup_dir"; exit 1; }

    # Iterate PM dirs in sorted order (numeric prefix controls order)
    local -a pm_dirs=()
    while IFS= read -r -d '' d; do
        pm_dirs+=("$d")
    done < <(find "$setup_dir" -maxdepth 1 -mindepth 1 -type d -print0 | sort -z)

    for pm_dir in "${pm_dirs[@]}"; do
        local pm_name; pm_name=$(basename "$pm_dir")
        pm_name="${pm_name#*.}"   # "1.yay" → "yay"

        command -v "$pm_name" &>/dev/null || { debug "skipping $pm_name: not in PATH"; continue; }
        log "--- $pm_name ---"

        local -a prepend_raw=() base_raw=() append_raw=()

        local pkgs_onlyf; pkgs_onlyf=$(_setup_find_pkgfile "$pm_dir/pkgs_only")
        if [[ -n "$pkgs_onlyf" ]]; then
            _setup_read_pkgs "$pkgs_onlyf" base_raw
        else
            local base_pkgsf;    base_pkgsf=$(_setup_find_pkgfile "$setup_dir/pkgs")
            local pkgs_prependf; pkgs_prependf=$(_setup_find_pkgfile "$pm_dir/pkgs_prepend")
            local pkgs_appendf;  pkgs_appendf=$(_setup_find_pkgfile "$pm_dir/pkgs_append")
            [[ -n "$pkgs_prependf" ]] && _setup_read_pkgs "$pkgs_prependf" prepend_raw
            [[ -n "$base_pkgsf"    ]] && _setup_read_pkgs "$base_pkgsf"    base_raw
            [[ -n "$pkgs_appendf"  ]] && _setup_read_pkgs "$pkgs_appendf"  append_raw
        fi

        local map_sh="$pm_dir/pkg_map.sh"
        local -a prepend_resolved=() base_resolved=() append_resolved=()
        for p in "${prepend_raw[@]}"; do
            local m; m=$(_setup_resolve_pkg "$map_sh" "$p"); [[ -n "$m" ]] && prepend_resolved+=("$m")
        done
        for p in "${base_raw[@]}"; do
            local m; m=$(_setup_resolve_pkg "$map_sh" "$p"); [[ -n "$m" ]] && base_resolved+=("$m")
        done
        for p in "${append_raw[@]}"; do
            local m; m=$(_setup_resolve_pkg "$map_sh" "$p"); [[ -n "$m" ]] && append_resolved+=("$m")
        done

        if [[ ${#prepend_resolved[@]} -eq 0 && ${#base_resolved[@]} -eq 0 && ${#append_resolved[@]} -eq 0 ]]; then
            log "  no packages, skipping"; continue
        fi

        local install_sh="$pm_dir/install.sh"
        [[ ! -f "$install_sh" ]] && { error "install.sh not found: $install_sh"; continue; }
        [[ ! -x "$install_sh" ]] && { error "install.sh not executable: $install_sh"; continue; }

        debug "→ prepend:${#prepend_resolved[@]} base:${#base_resolved[@]} append:${#append_resolved[@]}"
        WSH_PKGS_PREPEND="${prepend_resolved[*]:-}" \
        WSH_PKGS_BASE="${base_resolved[*]:-}" \
        WSH_PKGS_APPEND="${append_resolved[*]:-}" \
        "$install_sh" || return $?
    done
}

setup_do_shell() {
    printf 'export PATH="%s:$PATH"\n' "$wsh_dir" > "$HOME/.zshenv"
    printf 'eval "$(%s/w.sh -IA)"\n' "$wsh_dir" > "$HOME/.zshrc"
    log "wrote ~/.zshenv and ~/.zshrc"
}
