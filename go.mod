module github.com/rancher/dolly

go 1.12

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.2.0+incompatible
	github.com/matryer/moq => github.com/rancher/moq v0.0.0-20190404221404-ee5226d43009
	github.com/wercker/stern => github.com/StrongMonkey/stern v0.0.0-20200722225427-1373118cb3e0
)

require (
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/davecgh/go-spew v1.1.1
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/go-units v0.4.0
	github.com/drone/envsubst v1.0.2
	github.com/fatih/color v1.7.0
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/mattn/go-shellwords v1.0.10
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/rancher/wrangler v0.6.2-0.20200714200521-c61fae623942
	github.com/rancher/wrangler-cli v0.0.0-20200712180548-91e38f783aa5
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify v1.6.1
	github.com/wercker/stern v0.0.0-20190705090245-4fa46dd6987f
	golang.org/x/net v0.0.0-20200707034311-ab3426394381 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	gopkg.in/yaml.v2 v2.2.8
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
	k8s.io/api v0.18.5
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v0.18.0
	k8s.io/utils v0.0.0-20200619165400-6e3d28b6ed19 // indirect
	sigs.k8s.io/yaml v1.2.0
)
