#!/usr/bin/env bash
#set -euo pipefail
IFS=$'\t\n'

help_format="  %-3s %-20s %s\n"
help_format_1="  $help_format"
wsh_dir=$(dirname $(realpath "$0"))
plugin_dir="$wsh_dir/plugins"
shim_dir="$wsh_dir/shims"
dotfiles_dir="$wsh_dir/dotfiles"

# FLAGS
## GLOBAL
debug=
## INIT
init=
init_pyenv=
init_nvm=
init_goenv=
init_plugins=
init_functions=
##SETUP
setup=
setup_stow=
setup_simulate=
setup_delete=
setup_restow=
setup_list=
setup_packages=
setup_shell=
##TIME
time=
##GIT
git=
git_creds=
##AGENT
agent=
agent_super=

# ARGS
## SETUP
setup_pkg_args=()
## TIME
time_from_arg=
time_start_arg=
## GIT
git_cred_arg=

# LIBS
source "$wsh_dir/_utils.sh"
source "$wsh_dir/_time.sh"
source "$wsh_dir/_agent.sh"
source "$wsh_dir/_setup.sh"

print_option() {
    local level="$1"
    local short="$2"
    local long="$3"
    local arg="$4"
    local desc="$5"

    local indent="  "
    for ((i = 0; i < level; i++)); do
        indent="  $indent"
    done

    if [[ -n "$arg" ]]; then
        printf "${indent}%-3s %-20s %s\n" "$short" "$long <$arg>" "$desc"
    else
        printf "${indent}%-3s %-20s %s\n" "$short" "$long" "$desc"
    fi
}

usage_init() {
    print_option 0 "-I" "--init" "" "init options"
    print_option 1 "-h" "--help" "" "show this menu"
    print_option 1 "-n" "--nvm" "" "init nvm"
    print_option 1 "-p" "--pyenv" "" "init pyenv"
    print_option 1 "-z" "--plugins" "" "init zsh plugins from \$ZSH_PLUGIN_DIRS"
    print_option 1 "-f" "--functions" "" "init shell utility functions"
    print_option 1 "-g" "--goenv" "" "init goenv"
    print_option 1 "-A" "--all" "" "init all (nvm, pyenv, goenv, plugins, functions)"
}

usage_setup() {
    print_option 0 "-S" "--setup" "" "dotfiles / package setup"
    print_option 1 "-h" "--help" "" "show this menu"
    print_option 1 "-n" "--no,--simulate" "" "dry-run (passed to stow)"
    print_option 1 "-S" "--stow" "" "stow listed packages"
    print_option 1 "-D" "--delete" "" "unstow listed packages"
    print_option 1 "-R" "--restow" "" "restow listed packages"
    print_option 1 "-l" "--list" "" "list packages with deployed status"
    print_option 1 "-P" "--packages" "" "run SBOM package install"
    print_option 1 "-s" "--shell" "" "write ~/.zshrc and ~/.zshenv"
}

usage_time() {
    print_option 0 "-T" "--time" "" "time tracker options"
    print_option 1 "-h" "--help" "" "show this menu"
    print_option 1 "-b" "--shift-begin" "" "show shift begin time"
    print_option 1 "-e" "--shift-end" "" "show shift end time"
    print_option 1 "-r" "--earliest-leave" "" "show earliest leave time"
    print_option 1 "-w" "--worked" "" "show worked time today"
    print_option 1 "-t" "--total" "" "show total time worked since --from days ago"
    print_option 1 "-o" "--offline" "" "do not query the server for time; REQUIRES --start"
    print_option 1 "-i" "--overtime" "" "show overtime worked since --from days ago, or if --live is specified, include current shift"
    print_option 1 "-l" "--live" "" "include current shift in overtime calculation"
    print_option 1 "-f" "--from" "days" "from how many days ago to calculate overtime, 30 by default"
    print_option 1 "-s" "--start" "timestamp" "override shift start time and don't query the server for this info; REQUIRED when --offline"
}

usage_git() {
    print_option 0 "-G" "--git" "" "git options"
    print_option 1 "-h" "--help" "" "show this menu"
    print_option 1 "-c" "--credentials" "operation" "get git credentials for git from \$GIT_CREDS env var, when operation==get"
}

usage_agent() {
    print_option 0 "-A" "--agent" "" "agent options"
    print_option 1 "-h" "--help" "" "show this menu"
    print_option 1 "-I" "--init" "" "symlink agents/ and skill/ into tool config dirs"
    print_option 2 "-h" "--help" "" "show this menu"
    print_option 2 "-a" "--claude" "" "claude only (~/.claude)"
    print_option 2 "-o" "--copilot" "" "copilot only (~/.copilot)"
    print_option 1 "-a" "--claude" "" "run claude"
    print_option 1 "-o" "--copilot" "" "run copilot"
    print_option 1 "-S" "--super" "" "bypass permission checks"
    print_option 1 "-p" "--prompt" "file" "run tool with a .md file as system prompt"
    print_option 1 "--" "" "" "pass remaining args to the tool"
}

usage() {
    log ""
    log "usage: $(basename $0) OPERATION [...]"
    log ""
    print_option 0 "-h" "--help" "" "show this menu"
    print_option 0 "-v" "--verbose" "" "extensive logging"
    usage_init
    usage_setup
    usage_time
    usage_git
    usage_agent
}

pre_process_flags() {
    for arg in "$@"; do
        case "$arg" in
        --*)
            stdout "$arg"
            ;;
        -*)
            for ((i = 1; i < ${#arg}; i++)); do
                stdout "-${arg:i:1}"
            done
            ;;
        *)
            stdout "$arg"
            ;;
        esac
    done
}

unknown_flag() {
    error "unknown argument: $1"
    log "mayhaps it was used in the wrong context?"
    usage
    exit 1
}

no_arg_for_flag() {
    error "$1 requires an argument $2"
    usage
    exit 1
}

validate_time_from() {
    if [[ ! "$1" =~ ^[0-9]+$ ]] || [[ "$1" -lt 1 ]]; then
        error "days must be a positive integer, not $1"
        exit 1
    fi
}

validate_time_start() {
    if [[ ! "$1" =~ ^([0-9]|[0-1][0-9]|2[0-4]):([0-9]|[0-5][0-9])$ ]]; then
        error "timestamp must be in format of H:M, HH:MM and their permutations, not $1"
        exit 1
    fi
}

if [[ ${#@} -eq 0 ]]; then
    usage
    exit 1
fi
to_array flags "$(pre_process_flags "$@")"

# Pre-strip -v/--verbose so it works anywhere (including inside contexts)
new_flags=()
for _f in "${flags[@]}"; do
    case "$_f" in
    --verbose | -v) debug=1 ;;
    *) new_flags+=("$_f") ;;
    esac
done
flags=("${new_flags[@]}")
unset new_flags _f

for ((i = 0; i < "${#flags[@]}"; i++)); do
    case "${flags[$i]}" in
    --help | -h)
        usage
        exit 0
        ;;
    --init | -I)
        init=1
        ((i++)) || true
        for (( ; i < "${#flags[@]}"; i++)); do
            case "${flags[$i]}" in
            --help | -h)
                usage_init
                exit 0
                ;;
            --pyenv | -p)
                init_pyenv=1
                ;;
            --nvm | -n)
                init_nvm=1
                ;;
            --goenv | -g)
                init_goenv=1
                ;;
            --plugins | -z)
                init_plugins=1
                ;;
            --functions | -f)
                init_functions=1
                ;;
            --all | -A)
                init_nvm=1
                init_pyenv=1
                init_goenv=1
                init_plugins=1
                init_functions=1
                ;;
            *)
                unknown_flag "${flags[$i]}"
                ;;
            esac
        done
        ;;
    --setup | -S)
        setup=1
        ((i++)) || true
        for (( ; i < "${#flags[@]}"; i++)); do
            case "${flags[$i]}" in
            --help | -h)
                usage_setup
                exit 0
                ;;
            --no | --simulate | -n)
                setup_simulate=1
                ;;
            --stow | -S)
                setup_stow=1
                ;;
            --delete | -D)
                setup_delete=1
                ;;
            --restow | -R)
                setup_restow=1
                ;;
            --list | -l)
                setup_list=1
                ;;
            --packages | -P)
                setup_packages=1
                ;;
            --shell | -s)
                setup_shell=1
                ;;
            --* | -*)
                unknown_flag "${flags[$i]}"
                ;;
            *)
                setup_pkg_args+=("${flags[$i]}")
                ;;
            esac
        done
        ;;
    --time | -T)
        time=1
        ((i++)) || true
        for (( ; i < "${#flags[@]}"; i++)); do
            case "${flags[$i]}" in
            --help | -h)
                usage_time
                exit 0
                ;;
            --offline | -o)
                time_offline=1
                ;;
            --from | -f)
                time_from=1
                if [[ $((i + 1)) -lt "${#flags[@]}" ]]; then
                    ((i++))
                    time_from_arg="${flags[$i]}"
                    validate_time_from "$time_from_arg"
                else
                    no_arg_for_flag "${flags[$i]}" "days"
                fi
                ;;
            --start | -s)
                time_start=1
                if [[ $((i + 1)) -lt "${#flags[@]}" ]]; then
                    ((i++))
                    time_start_arg="${flags[$i]}"
                    validate_time_start "$time_start_arg"
                else
                    no_arg_for_flag "${flags[$i]}" "timestamp"
                fi
                ;;
            --total | -t)
                time_total=1
                ;;
            --overtime | -i)
                time_overtime=1
                ;;
            --shift-begin | -b)
                time_shift_begin=1
                ;;
            --shift-end | -e)
                time_shift_end=1
                ;;
            --earliest-leave | -r)
                time_earliest_leave=1
                ;;
            --worked | -w)
                time_worked=1
                ;;
            --live | -l)
                time_live=1
                ;;
            *)
                unknown_flag "${flags[$i]}"
                ;;
            esac
        done
        ;;
    --agent | -A)
        agent=1
        ((i++)) || true
        for (( ; i < "${#flags[@]}"; i++)); do
            case "${flags[$i]}" in
            --help | -h)
                usage_agent
                exit 0
                ;;
            --init | -I)
                agent_init=1
                ((i++)) || true
                for (( ; i < "${#flags[@]}"; i++)); do
                    case "${flags[$i]}" in
                    --help | -h)
                        usage_agent
                        exit 0
                        ;;
                    --claude | -a)
                        agent_claude=1
                        ;;
                    --copilot | -o)
                        agent_copilot=1
                        ;;
                    *)
                        unknown_flag "${flags[$i]}"
                        ;;
                    esac
                done
                ;;
            --claude | -a)
                agent_tool=claude
                ;;
            --copilot | -o)
                agent_tool=copilot
                ;;
            --super | -S)
                agent_super=1
                ;;
            --prompt | -p)
                if [[ $((i + 1)) -lt "${#flags[@]}" ]]; then
                    ((i++))
                    agent_prompt_arg="${flags[$i]}"
                else
                    no_arg_for_flag "${flags[$i]}" "file"
                fi
                ;;
            --)
                for ((i++; i < "${#flags[@]}"; i++)); do
                    agent_args+=("${flags[$i]}")
                done
                break
                ;;
            --* | -*)
                unknown_flag "${flags[$i]}"
                ;;
            esac
        done
        ;;
    --git | -G)
        git=1
        ((i++)) || true
        for (( ; i < "${#flags[@]}"; i++)); do
            case "${flags[$i]}" in
            --help | -h)
                usage_git
                exit 0
                ;;
            --credentials | -c)
                git_creds=1
                if [[ $((i + 1)) -lt "${#flags[@]}" ]]; then
                    ((i++))
                    git_cred_arg="${flags[$i]}"
                else
                    no_arg_for_flag "${flags[$i]}" "operation"
                fi
                ;;
            *)
                if [[ -n $git_creds ]]; then
                    continue # if creds are used, then we eat everything
                else
                    unknown_flag "${flags[$i]}"
                fi
                ;;
            esac
        done
        ;;
    *)
        unknown_flag "${flags[i]}"
        ;;
    esac
done

if [[ -n $init ]]; then
    debug "initializing autocompletions"
    cat "$wsh_dir/init/base.sh"
fi
if [[ -n $init_pyenv ]]; then
    debug "initializing pyenv"
    cat "$wsh_dir/init/pyenv.sh"
fi
if [[ -n $init_nvm ]]; then
    debug "initializing nvm"
    cat "$wsh_dir/init/nvm.sh"
fi
if [[ -n $init_goenv ]]; then
    debug "initializing goenv"
    cat "$wsh_dir/init/goenv.sh"
fi
if [[ -n $init_plugins ]]; then
    debug "initializing zsh plugins"
    if [[ -n "$ZSH_PLUGIN_DIRS" ]]; then
        IFS=':' read -ra _plugin_dirs <<<"$ZSH_PLUGIN_DIRS"
        for _plugin_dir in "${_plugin_dirs[@]}"; do
            for _plugin_file in "$_plugin_dir"/*.zsh; do
                [[ -f "$_plugin_file" ]] && echo "source \"$_plugin_file\""
            done
        done
    fi
fi
if [[ -n $init_functions ]]; then
    debug "initializing shell functions"
    cat "$wsh_dir/init/functions.sh"
fi
if [[ -n $setup ]]; then
    setup_do_init #TODO: maybe check if we're already inited and skip if so?
    if [[ -n $setup_list ]]; then
        setup_do_list
    elif [[ -n $setup_packages ]]; then
        setup_do_packages || exit $?
    elif [[ -n $setup_stow || -n $setup_simulate || -n $setup_delete || -n $setup_restow || ${#setup_pkg_args[@]} -gt 0 ]]; then
        setup_do_stow
    fi
    if [[ -n $setup_shell ]]; then
        setup_do_shell
    fi
fi
if [[ -n $time ]]; then
    _time_validate_env
    debug "yay"
    solidtime_get "/users/me/memberships"
fi
if [[ -n $agent ]]; then
    if [[ -n $agent_init ]]; then
        agent_do_init
    else
        agent_do_run
    fi
fi
if [[ -n $git ]]; then
    debug "git options"
    if [[ -n $git_creds ]]; then
        case "$git_cred_arg" in
        get)
            if [[ -z "$GIT_CREDS" ]]; then
                error "GIT_CREDS env var is not set"
                exit 1
            fi

            IFS='=' read -r git_user git_token <<<"$GIT_CREDS"
            if [[ -z "$git_user" || -z "$git_token" ]]; then
                error "GIT_CREDS env var is not in the correct format, must be user=token"
                exit 1
            fi
            stdout "username=$git_user"
            stdout "password=$git_token"
            ;;
        *)
            exit 0 # do nothing for unknown operations
            ;;
        esac
    fi
fi
