FROM golang:1.7
RUN apt-get update && apt-get install -y vim graphviz
RUN go get github.com/uber/go-torch \
    && cd $GOPATH/src/github.com/uber/go-torch \
    && git clone https://github.com/brendangregg/FlameGraph.git
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT /entrypoint.sh
CMD bash
