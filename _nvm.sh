#!/bin/env bash
_init_nvm(){
    unset -f nvm node npm pnpm npx
    export NVM_DIR="$HOME/.nvm"
    [ -s "/opt/homebrew/opt/nvm/nvm.sh" ] && \. "/opt/homebrew/opt/nvm/nvm.sh"  # This loads nvm
    [ -s "/opt/homebrew/opt/nvm/etc/bash_completion.d/nvm" ] && \. "/opt/homebrew/opt/nvm/etc/bash_completion.d/nvm"  # This loads nvm bash_completion

}

nvm() { _init_nvm; nvm "$@"; }

node() { _init_nvm; node "$@"; }

npm() { _init_nvm; npm "$@"; }

pnpm() { _init_nvm; pnpm "$@"; }

npx() { _init_nvm; npx "$@"; }

