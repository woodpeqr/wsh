#!/usr/bin/bash

stdout() {
    printf '%s\n' "$@"
}

log() {
    printf "%s\n" "$@" >&2 
}

debug(){
    if [[ -n $debug ]]; then
        log "v: $*"
    fi
}

error(){
    log "ERROR: $*"
}

to_array() {
    local arr_name=$1
    shift # first arg is the arr name, so let's ignore it from now on
    IFS=$'\n' read -r -d '' -a $arr_name <<< "$*" || true
}


