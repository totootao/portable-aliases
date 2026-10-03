#!/usr/bin/env bash
# =============================================================================
# aliases.sh  ——  由 aliases.json 自动生成, 请勿手改(保存时会被覆盖)
# 编辑请走 Web 管理界面, 或改 aliases.json 后运行: go run main.go generate
# 本文件可直接 source, 与 Docker 镜像共用, 保证可移植。
# =============================================================================
shopt -s expand_aliases 2>/dev/null || true

# ------------------------- 别名与函数(受 Web 管理) -------------------------
alias ll='ls -alF --color=auto'
alias la='ls -A --color=auto'
alias l='ls -CF --color=auto'
alias lt='ls -alFtr --color=auto'
alias ..='cd ..'
alias ...='cd ../..'
alias ....='cd ../../..'
alias df='df -h'
alias du='du -h'
alias free='free -m'
alias psg='ps aux | grep -i'
alias ports='ss -tulanp'
alias myip='curl -s https://ifconfig.me'
alias ping='ping -c 4'
alias g='git'
alias gs='git status -s'
alias ga='git add'
alias gc='git commit -m'
alias gp='git push'
alias gl='git pull'
mkcd() {
mkdir -p "$1" && cd "$1"
}
ff() {
find . -type f -iname "*$1*"
}
extract() {
if [ -f "$1" ]; then
  case "$1" in
    *.tar.bz2) tar xjf "$1" ;;
    *.tar.gz)  tar xzf "$1" ;;
    *.bz2)     bunzip2 "$1" ;;
    *.gz)      gunzip "$1" ;;
    *.tar)     tar xf "$1" ;;
    *.tbz2)    tar xjf "$1" ;;
    *.tgz)     tar xzf "$1" ;;
    *.zip)     unzip "$1" ;;
    *.Z)       uncompress "$1" ;;
    *.7z)      7z x "$1" ;;
    *) echo "'$1' 不是可识别的压缩格式" ;;
  esac
else
  echo "'$1' 不是有效文件"
fi
}
bigfiles() {
du -h "$@" 2>/dev/null | sort -rh | head -10
}
