#!/bin/bash
set -eu

apt_pkgs=""
failed=()

_fail() { failed+=("$1: $2"); }

while IFS= read -r pkg; do
    [ -z "$pkg" ] && continue
    case "$pkg" in
        starship)
            curl -sS https://starship.rs/install.sh | sh -s -- --yes \
                || _fail "$pkg" "curl install failed" ;;
        nvm)
            NVM_DIR="${NVM_DIR:-/opt/nvm}"
            NVM_VERSION="${NVM_VERSION:-0.40.1}"
            NODE_VERSION="${NODE_VERSION:-24.13.1}"
            sudo mkdir -p "$NVM_DIR" && sudo chown "$(id -u):$(id -g)" "$NVM_DIR"
            curl -o- "https://raw.githubusercontent.com/nvm-sh/nvm/v${NVM_VERSION}/install.sh" \
                | NVM_DIR="$NVM_DIR" bash \
                && . "$NVM_DIR/nvm.sh" \
                && nvm install "$NODE_VERSION" \
                && nvm alias default "$NODE_VERSION" \
                || _fail "$pkg" "install/setup failed" ;;
        pyenv)
            PYENV_ROOT="${PYENV_ROOT:-/opt/pyenv}"
            PYTHON_VERSION="${PYTHON_VERSION:-3.12.9}"
            export PYENV_ROOT
            sudo mkdir -p "$PYENV_ROOT" && sudo chown "$(id -u):$(id -g)" "$PYENV_ROOT"
            curl https://pyenv.run | bash \
                && export PATH="$PYENV_ROOT/bin:$PATH" \
                && pyenv install "$PYTHON_VERSION" \
                && pyenv global "$PYTHON_VERSION" \
                || _fail "$pkg" "install/setup failed" ;;
        goenv)
            GOENV_ROOT="${GOENV_ROOT:-/opt/goenv}"
            GO_VERSION="${GO_VERSION:-1.25.8}"
            export GOENV_ROOT
            sudo mkdir -p "$GOENV_ROOT" && sudo chown "$(id -u):$(id -g)" "$GOENV_ROOT"
            git clone https://github.com/go-nv/goenv.git "$GOENV_ROOT" \
                && export PATH="$GOENV_ROOT/bin:$PATH" \
                && goenv install "$GO_VERSION" \
                && goenv global "$GO_VERSION" \
                || _fail "$pkg" "install/setup failed" ;;
        zellij)
            ZELLIJ_VERSION=$(curl -s https://api.github.com/repos/zellij-org/zellij/releases/latest \
                | grep '"tag_name"' | cut -d'"' -f4)
            curl -fsSL "https://github.com/zellij-org/zellij/releases/download/${ZELLIJ_VERSION}/zellij-aarch64-unknown-linux-musl.tar.gz" \
                | sudo tar -xz -C /usr/local/bin \
                || _fail "$pkg" "download/extract failed" ;;
        claude-code)
            curl -fsSL https://claude.ai/install.sh | bash \
                || _fail "$pkg" "curl install failed" ;;
        github-copilot)
            curl -fsSL https://gh.io/copilot-install | bash \
                || _fail "$pkg" "curl install failed" ;;
        *)
            apt_pkgs="$apt_pkgs $pkg" ;;
    esac
done

# shellcheck disable=SC2086
if [ -n "$apt_pkgs" ]; then
    sudo apt-get install -y --no-install-recommends $apt_pkgs \
        || _fail "apt-get" "failed to install:$apt_pkgs"
fi

if [ ${#failed[@]} -gt 0 ]; then
    echo "ERROR: the following packages failed to install:" >&2
    for entry in "${failed[@]}"; do
        echo "  - $entry" >&2
    done
    exit 1
fi
