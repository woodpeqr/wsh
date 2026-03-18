#!/bin/bash
set -eu

failed=()
_fail() { failed+=("$1: $2"); }

# Install one space-separated group: batch apt first, then specials
_install_group() {
    local group="$1"
    [[ -z "$group" ]] && return 0

    local apt_pkgs=""
    local -a special_pkgs=()

    while IFS= read -r pkg; do
        [ -z "$pkg" ] && continue
        case "$pkg" in
            starship|nvm|pyenv|goenv|zellij|claude-code|github-copilot)
                special_pkgs+=("$pkg") ;;
            *)
                apt_pkgs="$apt_pkgs $pkg" ;;
        esac
    done <<< "$(printf '%s\n' $group)"

    # shellcheck disable=SC2086
    [ -n "$apt_pkgs" ] && sudo apt-get install -y --no-install-recommends $apt_pkgs \
        || _fail "apt-get" "failed to install:$apt_pkgs"

    for pkg in "${special_pkgs[@]}"; do
        case "$pkg" in
            starship)
                curl -sS https://starship.rs/install.sh | sh -s -- --yes \
                    || _fail "$pkg" "curl install failed" ;;
            nvm)
                # Installs nvm only — language version installed by Dockerfile
                NVM_DIR="${NVM_DIR:-/opt/nvm}"
                NVM_VERSION="${NVM_VERSION:-0.40.1}"
                sudo mkdir -p "$NVM_DIR" && sudo chown "$(id -u):$(id -g)" "$NVM_DIR"
                curl -o- "https://raw.githubusercontent.com/nvm-sh/nvm/v${NVM_VERSION}/install.sh" \
                    | NVM_DIR="$NVM_DIR" PROFILE=/dev/null bash \
                    || _fail "$pkg" "install failed" ;;
            pyenv)
                # Installs pyenv only — language version installed by Dockerfile
                PYENV_ROOT="${PYENV_ROOT:-/opt/pyenv}"
                export PYENV_ROOT
                sudo git clone https://github.com/pyenv/pyenv.git "$PYENV_ROOT" \
                    && sudo chown -R "$(id -u):$(id -g)" "$PYENV_ROOT" \
                    || _fail "$pkg" "install failed" ;;
            goenv)
                # Installs goenv only — language version installed by Dockerfile
                GOENV_ROOT="${GOENV_ROOT:-/opt/goenv}"
                export GOENV_ROOT
                sudo mkdir -p "$GOENV_ROOT" && sudo chown "$(id -u):$(id -g)" "$GOENV_ROOT"
                git clone https://github.com/go-nv/goenv.git "$GOENV_ROOT" \
                    || _fail "$pkg" "install failed" ;;
            zellij)
                _zarch=$(uname -m)
                ZELLIJ_VERSION=$(curl -s https://api.github.com/repos/zellij-org/zellij/releases/latest \
                    | grep '"tag_name"' | cut -d'"' -f4)
                curl -fsSL "https://github.com/zellij-org/zellij/releases/download/${ZELLIJ_VERSION}/zellij-${_zarch}-unknown-linux-musl.tar.gz" \
                    | sudo tar -xz -C /usr/local/bin \
                    || _fail "$pkg" "download/extract failed" ;;
            claude-code)
                curl -fsSL https://claude.ai/install.sh | bash \
                    || _fail "$pkg" "curl install failed" ;;
            github-copilot)
                curl -fsSL https://gh.io/copilot-install | bash \
                    || _fail "$pkg" "curl install failed" ;;
        esac
    done
}

_install_group "${WSH_PKGS_PREPEND:-}"
_install_group "${WSH_PKGS_BASE:-}"
_install_group "${WSH_PKGS_APPEND:-}"

if [ ${#failed[@]} -gt 0 ]; then
    echo "ERROR: the following packages failed to install:" >&2
    for entry in "${failed[@]}"; do echo "  - $entry" >&2; done
    exit 1
fi
