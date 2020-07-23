FROM golang:1.13.14
ENV GOPATH="/go"
RUN ["mkdir", "-p", "/go/src/github.com/rancher/demo"]
COPY * /go/src/github.com/rancher/demo/
WORKDIR /go/src/github.com/rancher/demo
RUN ["go", "build", "-o", "demo"]
CMD ["./demo"]