# FLAGS (set by w.sh during parsing)
setup=
setup_simulate=
setup_delete=
setup_restow=
setup_list=
setup_packages=

# ARGS
setup_dir_arg=
setup_pkg_args=()

# ── helpers ──────────────────────────────────────────────────────────────────

_setup_dotfiles_dir() {
    echo "${setup_dir_arg:-$wsh_dir/dotfiles}"
}

# Read a pkg file into a named array; skip blank lines and comments
_setup_read_pkgs() {
    local file="$1" arr_name="$2"
    while IFS= read -r line; do
        [[ -z "$line" || "$line" == \#* ]] && continue
        eval "${arr_name}+=(\"\$line\")"
    done < "$file"
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
            [[ ! -L "$HOME/$rel" ]] && { all_ok=0; break; }
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

    local base_pkgs="$setup_dir/pkgs.txt"

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

        local -a raw=()

        if [[ -f "$pm_dir/pkgs_only.txt" ]]; then
            # Override present: use exclusively (empty file = install nothing)
            _setup_read_pkgs "$pm_dir/pkgs_only.txt" raw
        else
            # Combine base + additions, then deduplicate preserving order via awk
            local -a combined=()
            [[ -f "$base_pkgs"           ]] && _setup_read_pkgs "$base_pkgs"           combined
            [[ -f "$pm_dir/pkgs_add.txt" ]] && _setup_read_pkgs "$pm_dir/pkgs_add.txt" combined
            while IFS= read -r line; do
                raw+=("$line")
            done < <(printf '%s\n' "${combined[@]}" | awk '!seen[$0]++')
        fi

        [[ ${#raw[@]} -eq 0 ]] && { log "  no packages, skipping"; continue; }

        # Resolve command names → package names via pkg_map.txt
        local map_file="$pm_dir/pkg_map.txt"
        local -a resolved=()
        for p in "${raw[@]}"; do
            local mapped="$p"
            if [[ -f "$map_file" ]]; then
                local entry; entry=$(grep -m1 "^${p}=" "$map_file" 2>/dev/null || true)
                [[ -n "$entry" ]] && mapped="${entry#*=}"
            fi
            resolved+=("$mapped")
        done

        local install_sh="$pm_dir/install.sh"
        [[ ! -f "$install_sh" ]] && { error "install.sh not found: $install_sh"; continue; }
        [[ ! -x "$install_sh" ]] && { error "install.sh not executable: $install_sh"; continue; }

        debug "→ ${#resolved[@]} packages to $pm_name"
        printf '%s\n' "${resolved[@]}" | "$install_sh"
    done
}
