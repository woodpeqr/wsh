fpath=($HOME/.config/wsh/completions $fpath)
autoload -Uz compinit
compinit
eval "$(starship init zsh)"
bindkey -v
