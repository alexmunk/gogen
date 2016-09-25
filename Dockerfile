FROM golang:1.7
RUN apt-get update && apt-get install -y vim graphviz
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT /entrypoint.sh
CMD bash
