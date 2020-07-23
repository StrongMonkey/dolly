package cmd

import (
	"fmt"
	"strings"

	cli "github.com/rancher/wrangler-cli"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewKillCommand() *cobra.Command {
	kill := cli.Command(&Kill{}, cobra.Command{
		Short: "kill/delete pods",
	})
	return kill
}

type Kill struct {
	Namespace string `name:"namespace" usage:"specify namespace" default:"default"`
}

func (k *Kill) Run(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("require at least one argument")
	}

	for _, arg := range args {
		if err := K8sInterface.CoreV1().Pods(k.Namespace).Delete(cmd.Context(), strings.TrimPrefix(arg, "pod/"), metav1.DeleteOptions{}); err != nil {
			return err
		}
	}
	return nil
}
