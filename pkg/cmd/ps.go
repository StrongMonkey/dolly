package cmd

import (
	"github.com/rancher/dolly/pkg/tables"
	cli "github.com/rancher/wrangler-cli"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewPsCommand() *cobra.Command {
	ps := cli.Command(&Ps{}, cobra.Command{
		Short: "Show kubernetes deployments/daemonset/statesets",
	})
	return ps
}

type Ps struct {
	All       bool   `name:"all" usage:"print resource in all namespaces" short:"a"`
	Namespace string `name:"namespace" usage:"print resource in one namespace" default:"default" short:"n"`
	Pod       bool   `name:"pod" usage:"only show pods" short:"p"`
	Quiet     bool   `name:"quiet" usage:"only print ID" short:"q"`
	Format    string `name:"format" usage:"format(yaml/json/jsoncompact/raw)"`
}

func (p *Ps) Run(cmd *cobra.Command, args []string) error {
	namespace := p.Namespace
	if p.All {
		namespace = ""
	}

	var output []runtime.Object
	if p.Pod {
		pods, err := K8sInterface.CoreV1().Pods(namespace).List(cmd.Context(), metav1.ListOptions{})
		if err != nil {
			return err
		}
		for i := range pods.Items {
			output = append(output, &pods.Items[i])
		}
		w := tables.NewPods(namespace, p.Format, p.Quiet)
		return w.Write(output)
	}

	deployments, err := K8sInterface.AppsV1().Deployments(namespace).List(cmd.Context(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	daemonsets, err := K8sInterface.AppsV1().DaemonSets(namespace).List(cmd.Context(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	statefulsets, err := K8sInterface.AppsV1().StatefulSets(namespace).List(cmd.Context(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for i := range deployments.Items {
		output = append(output, &deployments.Items[i])
	}

	for i := range daemonsets.Items {
		output = append(output, &daemonsets.Items[i])
	}

	for i := range statefulsets.Items {
		output = append(output, &statefulsets.Items[i])
	}

	w := tables.NewService(namespace, p.Format, p.Quiet)
	return w.Write(output)
}
