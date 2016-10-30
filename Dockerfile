FROM golang
RUN apt-get update && apt-get install -y vim graphviz bsdmainutils
RUN go get github.com/uber/go-torch \
    && cd $GOPATH/src/github.com/uber/go-torch \
    && git clone https://github.com/brendangregg/FlameGraph.git
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
CMD ["bash"]
