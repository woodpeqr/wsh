GOENV_ROOT="${GOENV_ROOT:-$HOME/.goenv}"
export GOENV_ROOT
[[ -d $GOENV_ROOT/bin ]] && export PATH="$GOENV_ROOT/bin:$PATH"
eval "$(command goenv init - zsh)"
