### Managing Application using Riofile

1. Define a [Riofile](https://raw.githubusercontent.com/StrongMonkey/dolly/master/example/Riofile).

```yaml
configs:
  conf:
    index.html: |-
      <!DOCTYPE html>
      <html>
      <body>

      <h1>Hello World</h1>

      </body>
      </html>
services:
  nginx:
    image: nginx # image to use
    ports:
    - 80:80/http # port to expose on service -> pods. 80/http assumes service port and container port are the same
    configs: # config map to mount
    - conf/index.html:/usr/share/nginx/html/index.html
    permission: # namespaced-scoped permission with serviceaccount used by pod
    - "list pods" # allows to list all the pods in the same namespace 
    globalPermission: # cluster-scoped permission with serviceaccount used by pod
    - "list services" # allows to list all the services in all namespaces 
```

2. Run `dolly up`.

```bash
$ dolly up -f https://raw.githubusercontent.com/StrongMonkey/dolly/master/example/Riofile
configmap/conf
deployment.apps/nginx
serviceaccount/nginx
role.rbac.authorization.k8s.io/nginx
rolebinding.rbac.authorization.k8s.io/nginx-nginx
clusterrole.rbac.authorization.k8s.io/nginx
clusterrolebinding.rbac.authorization.k8s.io/nginx-nginx
service/nginx
``` 

  Dolly will translate Riofile into kubernetes native manifest.

3. Creating a helm chart based on Riofile. `dolly render`

```bash
$ dolly render -f https://raw.githubusercontent.com/StrongMonkey/dolly/master/example/Riofile --chart-name riofile-demo --version 0.0.1
tarball riofile-demo-0.0.1.tar.gz created
```

### Using Riofile with source

Riofile can be defined to build images from source and dockerfile. It defines a simple build syntax that is similar to [docker-compose](https://docs.docker.com/compose/compose-file/#build) and call docker build directly on the host. 

1. Go to the example folder.
```bash
$ git clone git@github.com:StrongMonkey/dolly.git
$ cd dolly/example/build
```

The Riofile is defined as:

```yaml
services:
  demo:
   build:
     context: ./
   ports:
   - 80/http
```

2. Run `dolly up`.
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

Dolly translates build parameters defined in Riofile and send it to docker daemon to build the image. 

Note: the docker daemon has to be connected to kubernetes container runtime in order to use the image.

Note: if no image name is define in Riofile, the current working directory name will be used as image name.

### Using Template

Riofile embedded with templating function and variable substitutions.

Variable substitutions follows the syntax

```text
${var^}
${var^^}
${var,}
${var,,}
${var:position}
${var:position:length}
${var#substring}
${var##substring}
${var%substring}
${var%%substring}
${var/substring/replacement}
${var//substring/replacement}
${var/#substring/replacement}
${var/%substring/replacement}
${#var}
${var=default}
${var:=default}
${var:-default}
```

Templating allows user to use basic go template, for example

```yaml
services:
{{- if eq .Values.DEMO "true" }}
  demo:
    image: ${IMAGE}
    cpus: 100
    ports:
    - 80/http
{{- end }}

template:
  goTemplate: true
  envSubst: true
  variables:
  - DEMO
```

### Riofile reference

```yaml
# Configmap
configs:
  config-foo:     # specify name in the section
    key1: |-      # specify key and data in the section
      {{ config1 }}
    key2: |-
      {{ config2 }}

# Service
services:
  service-foo:

    # Scale setting
    scale: 2 # Specify scale of the service. If you pass range `1-10`, it will enable autoscaling which can be scale from 1 to 10. Default to 1 if omitted

    # Revision setting
    app: my-app # Specify app name. Defaults to service name. This is used to aggregate services that belongs to the same app.
    # Container configuration
    image: nginx # Container image. Required if not setting build
    imagePullPolicy: always # Image pull policy. Options: (always/never/ifNotProsent), defaults to ifNotProsent.
    build: # Setting build parameters. Set if you want to build image for source
      args: # Build arguments to pass to buildkit https://docs.docker.com/engine/reference/builder/#understand-how-arg-and-from-interact. Optional
      - foo=bar
      dockerfile: Dockerfile # The name of Dockerfile to look for.  This is the full path relative to the repo root. Defaults to `Dockerfile`.
      context: ./  # Docker build context. Defaults to .
      cache_from: # a list of string. A list of images that the engine uses for cache resolution.
      labels: # map of string. Add metadata to the resulting image using Docker labels.
        foo: bar
      shm_size: # Set the size of the /dev/shm partition for this build’s containers
      network: # Set the network containers connect to for the RUN instructions during build
      target: # build the specified stage as defined inside the Dockerfile
    command: # Container entrypoint, not executed within a shell. The docker image's ENTRYPOINT is used if this is not provided.
    - echo
    args: # Arguments to the entrypoint. The docker image's CMD is used if this is not provided.
    - "hello world"
    workingDir: /home # Container working directory
    ports: # Container ports, format: `$(servicePort:)containerPort/protocol`. Required if user wants to expose service through gateway
    - 8080:80/http,web # Service port 8080 will be mapped to container port 80 with protocol http, named `web`
    - 8080/http,admin,internal=true # Service port 8080 will be mapped to container port 8080 with protocol http, named `admin`, internal port(will not be exposed through gateway)
    env: # Specify environment variable
    - POD_NAME=$(self/name) # Mapped to "metadata.name"
    #
    # "self/name":           "metadata.name",
    # "self/namespace":      "metadata.namespace",
    # "self/labels":         "metadata.labels",
    # "self/annotations":    "metadata.annotations",
    # "self/node":           "spec.nodeName",
    # "self/serviceAccount": "spec.serviceAccountName",
    # "self/hostIp":         "status.hostIP",
    # "self/nodeIp":         "status.hostIP",
    # "self/ip":             "status.podIP",
    #
    cpus: 100m # Cpu request, format 0.5 or 500m. 500m = 0.5 core. If not set, cpu request will not be set. https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
    memory: 100Mi # Memory request. 100Mi, available options. If not set, memory request will not be set. https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
    secrets: # Specify secret to mount. Format: `$name/$key:/path/to/file`. Secret has to be pre-created in the same namespace
    - foo/bar:/my/password
    configs: # Specify configmap to mount. Format: `$name/$key:/path/to/file`.
    - foo/bar:/my/config
    livenessProbe: # LivenessProbe setting. https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/
      httpGet:
        path: /ping
        port: "9997" # port must be string
      initialDelaySeconds: 10
    readinessProbe: # ReadinessProbe https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/
      failureThreshold: 7
      httpGet:
        path: /ready
        port: "9997" # port must be string
    stdin: true # Whether this container should allocate a buffer for stdin in the container runtime
    stdinOnce: true # Whether the container runtime should close the stdin channel after it has been opened by a single attach. When stdin is true the stdin stream will remain open across multiple attach sessions.
    tty: true # Whether this container should allocate a TTY for itself
    runAsUser: 1000 # The UID to run the entrypoint of the container process.
    runAsGroup: 1000 # The GID to run the entrypoint of the container process
    readOnlyRootFilesystem: true # Whether this container has a read-only root filesystem
    privileged: true # Run container in privileged mode.

    nodeAffinity: # Describes node affinity scheduling rules for the pod.
    podAffinity:  # Describes pod affinity scheduling rules (e.g. co-locate this pod in the same node, zone, etc. as some other pod(s)).
    podAntiAffinity: # Describes pod anti-affinity scheduling rules (e.g. avoid putting this pod in the same node, zone, etc. as some other pod(s)).

    hostAliases: # Hostname alias
      - ip: 127.0.0.1
        hostnames:
        - example.com
    hostNetwork: true # Use host networking, defaults to False. If this option is set, the ports that will be used must be specified.
    imagePullSecrets: # Image pull secret https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod
    - secret1
    - secret2

    # Containers: Specify sidecars. Other options are available in container section above, this is limited example
    containers:
    - init: true # Init container
      name: my-init
      image: ubuntu
      args:
      - "echo"
      - "hello world"


    # Permission for service
    #
    #   global_permissions:
    #   - 'create,get,list certmanager.k8s.io/*'
    #
    #   this will give workload abilities to **create, get, list** **all** resources in api group **certmanager.k8s.io**.
    #
    #   If you want to hook up with an existing role:
    #
    #
    #   global_permissions:
    #   - 'role=cluster-admin'
    #
    #
    #   - `permisions`: Specify current namespace permission of workload
    #
    #   Example:
    #
    #   permissions:
    #   - 'create,get,list certmanager.k8s.io/*'
    #
    #
    #   This will give workload abilities to **create, get, list** **all** resources in api group **certmanager.k8s.io** in **current** namespace.
    #
    #   Example:
    #
    #   permissions:
    #   - 'create,get,list /node/proxy'
    #
    #    This will give subresource for node/proxy

    # Optional, will created and mount serviceAccountToken into pods with corresponding permissions
    global_permissions:
    - 'create,get,list certmanager.k8s.io/*'
    permissions:
    - 'create,get,list certmanager.k8s.io/*'

# Use Riofile's answer/question templating
template:
  goTemplate: true # use go templating
  envSubst: true # use ENV vars during templating

# Supply arbitrary kubernetes manifest yaml
kubernetes:
  manifest: |-
    apiVersion: apps/v1
    kind: Deployment
    ....
```