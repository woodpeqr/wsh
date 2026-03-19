#!/bin/sh
case "$1" in
    fd)                      echo "fd-find" ;;
    zsh-autosuggestions)     echo "zsh-autosuggestions" ;;
    zsh-syntax-highlighting) echo "zsh-syntax-highlighting" ;;
    zellij)                  echo "" ;;
    *)                       echo "$1" ;;
esac
