git_sign_branch() {
    git rebase --exec 'git commit --amend --no-edit -S' "$(git merge-base HEAD "${1:-main}")"
}
