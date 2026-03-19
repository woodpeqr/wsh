PYENV_ROOT="${PYENV_ROOT:-$HOME/.pyenv}"
export PYENV_ROOT
[[ -d $PYENV_ROOT/bin ]] && export PATH="$PYENV_ROOT/bin:$PATH"
eval "$(command pyenv init - zsh)"
