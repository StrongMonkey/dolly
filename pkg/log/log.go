package log

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"text/template"

	"github.com/fatih/color"
	"github.com/stern/stern/stern"
	"k8s.io/client-go/kubernetes"
)

func Output(ctx context.Context, conf *stern.Config, namespace string, k8s kubernetes.Interface) error {
	logCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	podInterface := k8s.CoreV1().Pods(namespace)
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
				tailsMutex.Lock()
				tails[id].Close()
				delete(tails, id)
				tailsMutex.Unlock()
			}
			tailOpts := &stern.TailOptions{
				SinceSeconds: int64(conf.Since.Seconds()),
				Timestamps:   conf.Timestamps,
				TailLines:    conf.TailLines,
				Exclude:      conf.Exclude,
				Include:      conf.Include,
				Namespace:    true,
			}
			newTail := stern.NewTail("", a.Namespace, a.Pod, a.Container, conf.Template, tailOpts)
			tailsMutex.Lock()
			tails[id] = newTail
			tailsMutex.Unlock()
			newTail.Start(logCtx, podInterface)
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

	<-logCtx.Done()
	return nil
}

// Format is based on both wercker/stern and linkerd/stern templating
func Format(format string, noColor bool) (*template.Template, error) {
	var tpl string
	switch format {
	case "json":
		tpl = "{{json .}}\n"
	case "raw":
		tpl = "{{.Message}}"
	default:
		tpl = "{{color .PodColor .PodName}} {{color .ContainerColor .ContainerName}} {{.Message}}\n"
		if noColor {
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
