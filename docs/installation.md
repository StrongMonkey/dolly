# Installation 

Dolly is a simple CLI which can be installed on anywhere. 

To install from source code:

```text
git clone git@github.com:StrongMonkey/dolly.git
cd dolly
make
chmod +x ./bin/dolly
mv ./bin/dolly /usr/local/bin
```

To install from the latest release, download the latest release from github pages.

Once it is finished, run `dolly` to make sure it is properly installed:

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

Setup `KUBECONFIG` to point to your k8s cluster and you are ready to rock.