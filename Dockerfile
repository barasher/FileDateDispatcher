FROM golang:1.12
WORKDIR $GOPATH/src/github.com/barasher/FileDateDispatcher
RUN apt-get update -y
RUN apt-get install -y libimage-exiftool-perl
COPY . .
RUN GO111MODULE=on go get -d ./...
RUN GO111MODULE=on go build cmd/dispatcher.go
RUN mkdir -p /var/dispatcher/in
RUN mkdir -p /var/dispatcher/out
RUN mkdir -p /etc/dispatcher
COPY docker.json /etc/dispatcher/dispatcher.json
#CMD ["/bin/bash"]
CMD [ "./dispatcher", "-c", "/etc/dispatcher/dispatcher.json", "-s", "/var/dispatcher/in", "-d", "/var/dispatcher/out" ]