package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sync"
	"text/template"
	"time"

	"github.com/fatih/color"
	"github.com/rancher/dolly/pkg/table/types"
	cli "github.com/rancher/wrangler-cli"
	"github.com/rancher/wrangler/pkg/kv"
	"github.com/spf13/cobra"
	"github.com/wercker/stern/stern"
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
	Namespace      string `name:"namespace" usage:"specify namespace" default:"default"`
	Container      string `name:"container" usage:"Print the logs of a specific container" short:"c"`
	InitContainers bool   `name:"init" usage:"Include or exclude init containers" default:"true"`
	NoColor        bool   `name:"no-color" usage:"Dont show color when logging" default:"false" short:"n"`
	Output         string `name:"output" usage:"Output format: [default, raw, json]"`
	Previous       bool   `name:"previous" usage:"Print the logs for the previous instance of the container in a pod if it exists, excludes running" short:"p"`
	Since          string `name:"since" desc:"Logs since a certain time, either duration (5s, 2m, 3h) or RFC3339" default:"24h"`
	Tail           int    `name:"tail" usage:"Number of recent lines to print, -1 for all" default:"200" short:"t"`
	Timestamps     bool   `name:"timestamps" usage:"Print the logs with timestamp" default:"false"`
}

func (l *Logs) Run(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("require exactly one parameter")
	}
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

	config.Template, err = l.logFormat()
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
			config.ContainerState = []string{stern.TERMINATED}
		} else {
			config.ContainerState = []string{stern.RUNNING, stern.WAITING}
		}
	}

	config.ContainerQuery, err = regexp.Compile(l.Container)
	if err != nil {
		return err
	}

	return l.output(cmd.Context(), config)
}

func (l *Logs) output(ctx context.Context, conf *stern.Config) error {
	sigCh := make(chan os.Signal, 1)
	logCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	podInterface := K8sInterface.CoreV1().Pods(l.Namespace)
	tails := make(map[string]*stern.Tail)
	tailsMutex := sync.RWMutex{}

	// See: https://github.com/linkerd/linkerd2/blob/c5a85e587c143d31f814d807e0e39cb4ad5e3572/cli/cmd/logs.go#L223-L227
	logC := make(chan string, 1024)
	go func() {
		for {
			select {
			case str := <-logC:
				fmt.Fprintf(os.Stdout, str)
			case <-logCtx.Done():
				break
			}
		}
	}()

	added, removed, err := stern.Watch(
		logCtx,
		podInterface,
		conf.PodQuery,
		conf.ContainerQuery,
		conf.ExcludeContainerQuery,
		conf.InitContainers,
		conf.ContainerState,
		conf.LabelSelector,
	)
	if err != nil {
		return err
	}

	go func() {
		for a := range added {
			id := a.GetID()
			tailsMutex.RLock()
			existing := tails[id]
			tailsMutex.RUnlock()
			if existing != nil {
				if existing.Active {
					continue
				} else { // cleanup failed tail to restart
					tailsMutex.Lock()
					tails[id].Close()
					delete(tails, id)
					tailsMutex.Unlock()
				}
			}
			tailOpts := &stern.TailOptions{
				SinceSeconds: int64(conf.Since.Seconds()),
				Timestamps:   conf.Timestamps,
				TailLines:    conf.TailLines,
				Exclude:      conf.Exclude,
				Include:      conf.Include,
				Namespace:    true,
			}
			newTail := stern.NewTail(a.Namespace, a.Pod, a.Container, conf.Template, tailOpts)
			tailsMutex.Lock()
			tails[id] = newTail
			tailsMutex.Unlock()
			newTail.Start(logCtx, podInterface, logC)
		}
	}()

	go func() {
		for r := range removed {
			id := r.GetID()
			tailsMutex.RLock()
			existing := tails[id]
			tailsMutex.RUnlock()
			if existing == nil {
				continue
			}
			tailsMutex.Lock()
			tails[id].Close()
			delete(tails, id)
			tailsMutex.Unlock()
		}
	}()

	<-sigCh
	return nil
}

// logFormat is based on both wercker/stern and linkerd/stern templating
func (l *Logs) logFormat() (*template.Template, error) {
	var tpl string
	switch l.Output {
	case "json":
		tpl = "{{json .}}\n"
	case "raw":
		tpl = "{{.Message}}"
	default:
		tpl = "{{color .PodColor .PodName}} {{color .ContainerColor .ContainerName}} {{.Message}}"
		if l.NoColor {
			tpl = "{{.PodName}} {{.ContainerName}} {{.Message}}"
		}
	}
	funk := map[string]interface{}{
		"json": func(in interface{}) (string, error) {
			b, err := json.Marshal(in)
			if err != nil {
				return "", err
			}
			return string(b), nil
		},
		"color": func(color color.Color, text string) string {
			return color.SprintFunc()(text)
		},
	}
	template, err := template.New("log").Funcs(funk).Parse(tpl)
	if err != nil {
		return nil, err
	}
	return template, nil
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
