#!/bin/sh
case "$1" in
    fd)                      echo "fd-find" ;;
    zsh-autosuggestions)     echo "zsh-autosuggestions" ;;
    zsh-syntax-highlighting) echo "zsh-syntax-highlighting" ;;
    *)                       echo "$1" ;;
esac
