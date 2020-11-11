dolly
========

A CLI tool to convert Dollyfile(docker-compose) into kubernetes manifest and helm chart

## Installing

`make`.

## Running

Dolly is a simple CLI tool that transform Dollyfile(docker-compose syntax) into kubernetes manifest and helm charts.

```text
Create, manage kubernetes application using dollyfile

Usage:
  dolly [flags]
  dolly [command]

Available Commands:
  build       Run docker build using dollyfile syntax
  exec        Exec into pods
  help        Help about any command
  kill        kill/delete pods
  logs        Log deployments/daemonsets/statefulsets/pods
  ps          Show kubernetes deployments/daemonset/statesets
  push        Run docker build and push using dollyfile syntax
  render      Creating helm charts based on dollyfile
  rm          remove resources
  up          Applying kubernetes application using dollyfile

Flags:
      --debug               Enable debug log
  -h, --help                help for dolly
      --kubeconfig string   Path to the kubeconfig file to use for CLI requests.

```

### Quick start

1. Install `dolly` from source code.

```bash
$ git clone git@github.com:StrongMonkey/dolly.git
$ cd dolly
$ make
$ chmod +x ./bin/dolly
$ mv ./bin/dolly /usr/local/bin
```

2. Run `dolly up`.

```bash
$ export KUBECONFIG=/path/to/your/config
$ dolly up -f https://raw.githubusercontent.com/StrongMonkey/dolly/master/example/Dollyfile 
```

### Documentation

Detailed documentation can be found in [here](https://dolly.do.rancher.space)

## License
Copyright (c) 2020 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
