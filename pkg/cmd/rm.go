package cmd

import (
	"github.com/rancher/dolly/pkg/table/types"
	cli "github.com/rancher/wrangler-cli"
	"github.com/rancher/wrangler/pkg/kv"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewRmCommand() *cobra.Command {
	rm := cli.Command(&Rm{}, cobra.Command{
		Short: "remove resources",
	})
	return rm
}

type Rm struct {
	Namespace string `name:"namespace" usage:"specify namespace" default:"default"`
}

func (r *Rm) Run(cmd *cobra.Command, args []string) error {
	for _, arg := range args {
		t, resourceName := kv.Split(arg, "/")
		if t == types.DeploymentType {
			if err := K8sInterface.AppsV1().Deployments(r.Namespace).Delete(cmd.Context(), resourceName, metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
				return err
			}
		} else if t == types.DaemonSetType {
			if err := K8sInterface.AppsV1().DaemonSets(r.Namespace).Delete(cmd.Context(), resourceName, metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
				return err
			}
		} else if t == types.PodType {
			if err := K8sInterface.CoreV1().Pods(r.Namespace).Delete(cmd.Context(), resourceName, metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
				return err
			}
		}
	}
	return nil
}
