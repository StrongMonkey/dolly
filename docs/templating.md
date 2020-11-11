# Templating

Dollyfile embedded with templating function and variable substitutions.

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
${var=default}
${var:=default}
${var:-default}
```

Templating allows user to use basic go template if goTemplate is turned on

```yaml
services:
  demo:
    image: ${IMAGE}
    cpus: 100
    ports:
    - 80/http

template:
  goTemplate: true
  envSubst: true
  variables:
  - DEMO
```
