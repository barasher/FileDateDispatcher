FROM golang:1.12
WORKDIR $GOPATH/src/github.com/barasher/FileDateDispatcher
RUN apt-get update -y
RUN apt-get install -y libimage-exiftool-perl
COPY . .
RUN GO111MODULE=on go get -d ./...
RUN GO111MODULE=on go build cmd/dispatcher.go
RUN mkdir /var/in
RUN mkdir /var/out
CMD [ "./dispatcher", "-s", "/var/in", "-d", "/var/out" ]
#CMD ["/bin/bash"]