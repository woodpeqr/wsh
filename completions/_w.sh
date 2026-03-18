#compdef w.sh

_w_sh() {
    local context state
    local -a args
    local i top_ctx agent_init_ctx agent_tool_choice after_separator dd_pos

    # Walk the already-typed words to determine context
    top_ctx=''
    agent_init_ctx=0
    agent_tool_choice='claude'
    after_separator=0
    dd_pos=0
    i=1
    while [[ $i -lt $CURRENT ]]; do
        local w="${words[$i]}"
        case "$w" in
            # Expand single-char bundles like -Ipn -> -I -p -n for context detection
            -*[!-]*)
                if [[ "$w" == --* ]]; then
                    case "$w" in
                        --init)
                            if [[ "$top_ctx" == agent ]]; then
                                agent_init_ctx=1
                            else
                                top_ctx=init
                            fi
                            ;;
                        --setup)    top_ctx=setup ;;
                        --time)     top_ctx=time ;;
                        --git)      top_ctx=git ;;
                        --agent)    top_ctx=agent ;;
                        --claude)
                            if [[ "$top_ctx" == agent && $agent_init_ctx -eq 0 ]]; then
                                agent_tool_choice=claude
                            fi
                            ;;
                        --copilot)
                            if [[ "$top_ctx" == agent && $agent_init_ctx -eq 0 ]]; then
                                agent_tool_choice=copilot
                            fi
                            ;;
                        --)
                            if [[ "$top_ctx" == agent ]]; then
                                after_separator=1
                                dd_pos=$i
                                break
                            fi
                            ;;
                    esac
                else
                    # Bundled short flags: check each char
                    local bundle="${w#-}"
                    local j
                    for (( j=0; j < ${#bundle}; j++ )); do
                        local c="${bundle:$j:1}"
                        case "$c" in
                            I)
                                if [[ "$top_ctx" == agent ]]; then
                                    agent_init_ctx=1
                                else
                                    top_ctx=init
                                fi
                                ;;
                            S) top_ctx=setup ;;
                            T) top_ctx=time ;;
                            G) top_ctx=git ;;
                            A) top_ctx=agent; agent_init_ctx=0 ;;
                            a)
                                if [[ "$top_ctx" == agent && $agent_init_ctx -eq 0 ]]; then
                                    agent_tool_choice=claude
                                fi
                                ;;
                            o)
                                if [[ "$top_ctx" == agent && $agent_init_ctx -eq 0 ]]; then
                                    agent_tool_choice=copilot
                                fi
                                ;;
                        esac
                    done
                fi
                ;;
        esac
        # Consume argument-taking flags
        case "$w" in
            -f|--from|-s|--start|-c|--credentials|-p|--prompt|-d|--dir)
                (( i++ ))   # skip next word (the argument)
                ;;
        esac
        (( i++ ))
    done

    # After -- in agent context: delegate completions to the chosen tool
    if [[ $after_separator -eq 1 && "$top_ctx" == agent ]]; then
        local tool_cmd="${agent_tool_choice:-claude}"
        words=("$tool_cmd" "${(@)words[$((dd_pos+1)),-1]}")
        CURRENT=$(( CURRENT - dd_pos + 1 ))
        _normal
        return
    fi

    # Check if we're right after an argument-taking flag
    local prev="${words[$CURRENT-1]}"
    case "$prev" in
        -f|--from)
            _message 'days (positive integer)'
            return
            ;;
        -s|--start)
            _message 'HH:MM timestamp'
            return
            ;;
        -c|--credentials)
            local -a cred_ops
            cred_ops=(get store erase)
            _describe 'operation' cred_ops
            return
            ;;
        -p|--prompt)
            if [[ "$top_ctx" == agent ]]; then
                _files -g '*.md'
                return
            fi
            ;;
        -d|--dir)
            _files -/
            return
            ;;
    esac

    # Now offer completions based on context
    case "$top_ctx" in
        '')
            # No context yet — top-level flags
            local -a top_flags
            top_flags=(
                '-h:show help'
                '--help:show help'
                '-v:verbose logging'
                '--verbose:verbose logging'
                '-I:init options'
                '--init:init options'
                '-S:setup options'
                '--setup:setup options'
                '-T:time tracker options'
                '--time:time tracker options'
                '-G:git options'
                '--git:git options'
                '-A:agent options'
                '--agent:agent options'
            )
            _describe 'option' top_flags
            ;;

        init)
            local -a init_flags
            init_flags=(
                '-h:show help'            '--help:show help'
                '-n:init nvm'             '--nvm:init nvm'
                '-p:init pyenv'           '--pyenv:init pyenv'
                '-z:init zsh plugins from $ZSH_PLUGIN_DIRS'
                '--plugins:init zsh plugins from $ZSH_PLUGIN_DIRS'
                '-f:init shell utility functions'
                '--functions:init shell utility functions'
            )
            _describe 'init option' init_flags
            ;;

        setup)
            local -a setup_flags
            setup_flags=(
                '-h:show help'           '--help:show help'
                '-n:dry-run'             '--no:dry-run'          '--simulate:dry-run'
                '-d:dotfiles directory'  '--dir:dotfiles directory'
                '-D:unstow packages'     '--delete:unstow packages'
                '-R:restow packages'     '--restow:restow packages'
                '-l:list packages'       '--list:list packages'
                '-P:install packages'    '--packages:install packages'
            )
            _describe 'setup option' setup_flags
            ;;

        time)
            local -a time_flags
            time_flags=(
                '-h:show help'
                '--help:show help'
                '-b:show shift begin time'
                '--shift-begin:show shift begin time'
                '-e:show shift end time'
                '--shift-end:show shift end time'
                '-r:show earliest leave time'
                '--earliest-leave:show earliest leave time'
                '-w:show worked time today'
                '--worked:show worked time today'
                '-t:show total time worked'
                '--total:show total time worked'
                '-i:show overtime worked'
                '--overtime:show overtime worked'
                '-l:include current shift in overtime'
                '--live:include current shift in overtime'
                '-o:do not query server (requires --start)'
                '--offline:do not query server (requires --start)'
                '-f:days ago to calculate from (default 30)'
                '--from:days ago to calculate from (default 30)'
                '-s:override shift start time (HH:MM)'
                '--start:override shift start time (HH:MM)'
            )
            _describe 'time option' time_flags
            ;;

        git)
            local -a git_flags
            git_flags=(
                '-h:show help'
                '--help:show help'
                '-c:git credential helper'
                '--credentials:git credential helper'
            )
            _describe 'git option' git_flags
            ;;

        agent)
            if [[ $agent_init_ctx -eq 1 ]]; then
                local -a agent_init_flags
                agent_init_flags=(
                    '-h:show help'
                    '--help:show help'
                    '-a:claude only'
                    '--claude:claude only'
                    '-o:copilot only'
                    '--copilot:copilot only'
                )
                _describe 'agent init option' agent_init_flags
            else
                local -a agent_flags
                agent_flags=(
                    '-h:show help'
                    '--help:show help'
                    '-I:init symlinks'
                    '--init:init symlinks'
                    '-a:run claude'
                    '--claude:run claude'
                    '-o:run copilot'
                    '--copilot:run copilot'
                    '-S:bypass permission checks'
                    '--super:bypass permission checks'
                    '-p:system prompt .md file'
                    '--prompt:system prompt .md file'
                    '--:pass remaining args to tool'
                )
                _describe 'agent option' agent_flags
            fi
            ;;
    esac
}

_w_sh "$@"
