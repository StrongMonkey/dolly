package cmd

import (
	"github.com/rancher/dolly/pkg/tables"
	cli "github.com/rancher/wrangler-cli"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewSecretCommand() *cobra.Command {
	ps := cli.Command(&Secret{}, cobra.Command{
		Short: "Show kubernetes secret",
	})
	return ps
}

type Secret struct {
	Namespace string `name:"namespace" usage:"print resource in one namespace" default:"default" short:"n"`
	Quiet     bool   `name:"quiet" usage:"only print ID" short:"q"`
	Format    string `name:"format" usage:"format(yaml/json/jsoncompact/raw)"`
}

func (s *Secret) Run(cmd *cobra.Command, args []string) error {
	var output []runtime.Object
	cms, err := K8sInterface.CoreV1().Secrets(s.Namespace).List(cmd.Context(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for i := range cms.Items {
		output = append(output, &cms.Items[i])
	}
	w := tables.NewSecret(s.Namespace, s.Format, s.Quiet)
	return w.Write(output)
}
