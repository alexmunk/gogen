#!/bin/bash

set -e

export PATH=$PATH:/go/src/github.com/uber/go-torch/FlameGraph
cat > $HOME/.bashrc <<- EOF
export PS1="\[\033[36m\]\u\[\033[m\]@\[\033[32m\] \[\033[33;1m\]\w\[\033[m\] (\$(git branch 2>/dev/null | grep '^*' | colrm 1 2)) \$ "
EOF

cat > $HOME/.vimrc <<- EOF
syntax on
EOF

cat > $HOME/.gitconfig <<- EOF
[color]
  diff = auto
  status = auto
  branch = auto
  interactive = auto
  ui = true
  pager = true
EOF

chmod 0600 /root/.ssh/id_rsa

cd $HOME

"$@"
