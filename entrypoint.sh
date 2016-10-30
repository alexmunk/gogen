#!/bin/bash

set -e

export PATH=$PATH:/go/src/github.com/uber/go-torch/FlameGraph
export HOME=/go/src/github.com/coccyx/gogen
cat > $HOME/.profile <<- EOF
export PS1="\[\033[36m\]\u\[\033[m\]@\[\033[32m\] \[\033[33;1m\]\w\[\033[m\] (\$(git branch 2>/dev/null | grep '^*' | colrm 1 2)) \$ "
EOF

cat > $HOME/.vimrc <<- EOF
syntax on
EOF

cd $HOME

"$@"
