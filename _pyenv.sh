#!/bin/env bash
_init_pyenv() {
    unset -f pyenv python python3 pip pip3

    export PYENV_ROOT="$HOME/.pyenv"
    [[ -d $PYENV_ROOT/bin ]] && export PATH="$PYENV_ROOT/bin:$PATH"
    eval "$(command pyenv init - zsh)"
}

pyenv() { _init_pyenv; pyenv "$@"; }
python() { _init_pyenv; python "$@"; }
python3() { _init_pyenv; python3 "$@"; }
pip() { _init_pyenv; pip "$@"; }
pip3() { _init_pyenv; pip3 "$@"; }
