# Build

Dollyfile can be defined to build images from source and dockerfile. It defines a simple build syntax that is similar to [docker-compose](https://docs.docker.com/compose/compose-file/#build) and call docker build directly on the host. 

Go to the example folder.
```bash
$ git clone git@github.com:StrongMonkey/dolly.git
$ cd dolly/example/build
```

The Dollyfile is defined as:

```yaml
services:
  demo:
   build:
     context: ./
   ports:
   - 80/http
```

Run `dolly up`.
```bash
$ dolly build
Sending build context to Docker daemon  4.096kB
Step 1/7 : FROM golang:1.13.14
 ---> afae231e0b45
Step 2/7 : ENV GOPATH="/go"
 ---> Using cache
 ---> 96492ecf656e
Step 3/7 : RUN ["mkdir", "-p", "/go/src/github.com/rancher/demo"]
 ---> Using cache
 ---> 57f9a193f5e9
Step 4/7 : COPY * /go/src/github.com/rancher/demo/
 ---> Using cache
 ---> 76e9d067d47d
Step 5/7 : WORKDIR /go/src/github.com/rancher/demo
 ---> Using cache
 ---> 409a4a0d8451
Step 6/7 : RUN ["go", "build", "-o", "demo"]
 ---> Using cache
 ---> ce85ab5e58d5
Step 7/7 : CMD ["./demo"]
 ---> Using cache
 ---> fa7d16808743
Successfully built fa7d16808743
Successfully tagged build:latest
deployment.apps/demo
service/demo
```

Dolly translates build parameters defined in Dollyfile and send it to docker daemon to build the image. 

Note: the docker daemon has to be connected to kubernetes container runtime in order to use the image.

Note: if no image name is define in Dollyfile, the current working directory name will be used as image name.