#!/bin/env bash
export NVM_DIR="$HOME/.nvm"

_init_nvm(){
    unset -f nvm node npm pnpm npx
}

nvm() { _init_nvm; nvm "$@"; }

node() { _init_nvm; node "$@"; }

npm() { _init_nvm; npm "$@"; }

pnpm() { _init_nvm; pnpm "$@"; }

npx() { _init_nvm; npx "$@"; }

