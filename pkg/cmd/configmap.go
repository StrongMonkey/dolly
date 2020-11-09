package cmd

import (
	"github.com/rancher/dolly/pkg/tables"
	cli "github.com/rancher/wrangler-cli"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewConfigCommand() *cobra.Command {
	ps := cli.Command(&Config{}, cobra.Command{
		Short: "Show kubernetes configmap",
	})
	return ps
}

type Config struct {
	Namespace string `name:"namespace" usage:"print resource in one namespace" default:"default" short:"n"`
	Quiet     bool   `name:"quiet" usage:"only print ID" short:"q"`
	Format    string `name:"format" usage:"format(yaml/json/jsoncompact/raw)"`
}

func (c *Config) Run(cmd *cobra.Command, args []string) error {
	var output []runtime.Object
	cms, err := K8sInterface.CoreV1().ConfigMaps(c.Namespace).List(cmd.Context(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for i := range cms.Items {
		output = append(output, &cms.Items[i])
	}
	w := tables.NewConfig(c.Namespace, c.Format, c.Quiet)
	return w.Write(output)
}
