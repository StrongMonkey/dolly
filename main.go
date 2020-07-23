package main

import (
	"github.com/rancher/dolly/pkg/cmd"
	cli "github.com/rancher/wrangler-cli"

	_ "github.com/rancher/wrangler/pkg/generated/controllers/apps/v1"
	_ "github.com/rancher/wrangler/pkg/generated/controllers/core/v1"
	_ "github.com/rancher/wrangler/pkg/generated/controllers/rbac/v1"
)

func main() {
	cli.Main(cmd.New())
}
