package cmd

import (
	"regexp"
	"time"

	"github.com/rancher/dolly/pkg/log"
	"github.com/rancher/dolly/pkg/table/types"
	cli "github.com/rancher/wrangler-cli"
	"github.com/rancher/wrangler/pkg/kv"
	"github.com/spf13/cobra"
	"github.com/stern/stern/stern"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewLogCommand() *cobra.Command {
	logs := cli.Command(&Logs{}, cobra.Command{
		Short: "Log deployments/daemonsets/statefulsets/pods",
	})
	return logs
}

type Logs struct {
	Namespace      string `name:"namespace" usage:"specify namespace" default:"default" short:"n"`
	Container      string `name:"container" usage:"Print the logs of a specific container" short:"c"`
	InitContainers bool   `name:"init" usage:"Include or exclude init containers" default:"true"`
	NoColor        bool   `name:"no-color" usage:"Dont show color when logging" default:"false"`
	Output         string `name:"output" usage:"Output format: [default, raw, json]"`
	Previous       bool   `name:"previous" usage:"Print the logs for the previous instance of the container in a pod if it exists, excludes running" short:"p"`
	Since          string `name:"since" desc:"Logs since a certain time, either duration (5s, 2m, 3h) or RFC3339" default:"24h"`
	Tail           int    `name:"tail" usage:"Number of recent lines to print, -1 for all" default:"200" short:"t"`
	Timestamps     bool   `name:"timestamps" usage:"Print the logs with timestamp" default:"false"`
}

func (l *Logs) Run(cmd *cobra.Command, args []string) error {
	config := &stern.Config{
		LabelSelector:  labels.Everything(),
		Timestamps:     l.Timestamps,
		Namespace:      l.Namespace,
		InitContainers: l.InitContainers,
	}

	var result runtime.Object
	var err error
	t, resourceName := kv.Split(args[0], "/")
	if t == types.DeploymentType {
		result, err = K8sInterface.AppsV1().Deployments(l.Namespace).Get(cmd.Context(), resourceName, metav1.GetOptions{})
		if err != nil {
			return err
		}
	} else if t == types.DaemonSetType {
		result, err = K8sInterface.AppsV1().DaemonSets(l.Namespace).Get(cmd.Context(), resourceName, metav1.GetOptions{})
		if err != nil {
			return err
		}
	} else if t == types.PodType {
		result, err = K8sInterface.CoreV1().Pods(l.Namespace).Get(cmd.Context(), resourceName, metav1.GetOptions{})
		if err != nil {
			return err
		}
	}

	podName, sel, err := ToPodNameOrSelector(result)
	if err != nil {
		return err
	}

	if podName == "" {
		config.LabelSelector = sel
		config.PodQuery, err = regexp.Compile("")
	} else {
		config.PodQuery, err = regexp.Compile(regexp.QuoteMeta(podName))
	}

	config.Template, err = log.Format(l.Output, l.NoColor)
	if err != nil {
		return err
	}

	tail := int64(l.Tail)
	config.TailLines = &tail

	config.Since, err = time.ParseDuration(l.Since)
	if err != nil {
		return err
	}

	if len(config.ContainerState) == 0 {
		if l.Previous {
			config.ContainerState = stern.TERMINATED
		} else {
			config.ContainerState = stern.RUNNING
		}
	}

	config.ContainerQuery, err = regexp.Compile(l.Container)
	if err != nil {
		return err
	}

	return log.Output(cmd.Context(), config, l.Namespace, K8sInterface)
}

func ToPodNameOrSelector(obj runtime.Object) (string, labels.Selector, error) {
	switch v := obj.(type) {
	case *corev1.Pod:
		return v.Name, nil, nil
	case *appsv1.Deployment:
		return toSelector(v.Spec.Selector)
	case *appsv1.DaemonSet:
		return toSelector(v.Spec.Selector)
	}

	return "", labels.Nothing(), nil
}

func toSelector(sel *metav1.LabelSelector) (string, labels.Selector, error) {
	l, err := metav1.LabelSelectorAsSelector(sel)
	return "", l, err
}
