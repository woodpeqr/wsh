#!/usr/bin/bash

# GLOBAL VARS
SOLIDTIME_URL=${SOLIDTIME_URL:="https://chrono.woodpeqr.net"}
solidtime_api_url="$SOLIDTIME_URL/api/v1"
SOLIDTIME_ORG=${SOLIDTIME_ORG:="Stiffy's Organization"}
SOLIDTIME_PROJECT=${SOLIDTIME_PROJECT:="Work"}

work_hours_per_day=8

# LIBS
source "$(dirname $(realpath "$0"))/_utils.sh"

# LOCAL VARS
solidtime_user_id=
solidtime_org_id=
solidtime_project_id=

solidtime_get() {
    local endpoint=$1
    shift
    local url="$solidtime_api_url/$endpoint"

    if [[ -z $SOLIDTIME_API_KEY ]]; then
        error "\$SOLIDTIME_API_KEY env var is not set!"        
    fi

    local query
    query=$(printf "%s&" "$@")
    query="${query%&}"

    [[ -n "$query" ]] && url="$url?$query"

    curl \
        -H "Authorization: Bearer $SOLIDTIME_API_KEY" \
        -H "Content-Type: application/json" \
        -H "Accept: application/json" \
        -L \
        "$url"
}

solidtime_set_user_id(){
    solidtime_user_id=$(solidtime_get "/users/me" | jq .data.id)
}

solidtime_set_org_id(){
    memberships=$(solidtime_get "/users/me/memberships")
    echo $memberships
}

solidtime_set_project_id() {
    command ...
}



