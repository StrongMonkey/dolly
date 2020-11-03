# Quick start

## Create a compose file.

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
    - 8082:80/http # port to expose on service -> pods. 80/http assumes service port and container port are the same
    configs: # config map to mount
    - conf:/usr/share/nginx/html/
```

## Dolly up

Run dolly up. It will start watching and applying the compose file. 

```text
$ dolly up -f ./riofile
configmap/conf
deployment.apps/nginx
service/nginx
Forwarding from 127.0.0.1:8082 -> 80
Forwarding from [::1]:8082 -> 80
```

Dolly translates the compose file into k8s resource(configmap, service, deployment) and deploy them into cluster. You can now visit http://127.0.0.1:8082 to access your service. 

If you make any changes to your riofile, changes will automatically applied.

```text
# Edit the file so that configs print Hello Dolly
configs:
  conf:
    index.html: |-
      <!DOCTYPE html>
      <html>
      <body>

      <h1>Hello Dolly</h1>

      </body>
      </html>
services:
  nginx:
    image: nginx # image to use
    ports:
    - 8082:80/http # port to expose on service -> pods. 80/http assumes service port and container port are the same
    configs: # config map to mount
    - conf:/usr/share/nginx/html/
```

```text
$ dolly up -f ./riofile
configmap/conf
deployment.apps/nginx
service/nginx
Forwarding from 127.0.0.1:8082 -> 80
Forwarding from [::1]:8082 -> 80
configmap/conf
deployment.apps/nginx
service/nginx
Forwarding from 127.0.0.1:8082 -> 80
Forwarding from [::1]:8082 -> 80
```

It will re-apply the changed file. If you visit http://127.0.0.1:8082 again(wait for a minute for configmap change to appear), you should see `Hello Dolly` now.
 
Enjoy the journey!
