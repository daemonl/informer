FROM golang:1.10
ENV GOPATH=/go
COPY . /go/src/github.com/daemonl/informer
RUN go build -o /informer github.com/daemonl/informer
WORKDIR /go/src/github.com/daemonl/informer
CMD ["/informer"]

