# Reference

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

# Use Dollyfile's answer/question templating
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