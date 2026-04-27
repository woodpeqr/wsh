# fpath entry for completions is injected by w.sh (uses $wsh_dir/completions)
autoload -Uz compinit
compinit
eval "$(starship init zsh)"
bindkey -v
