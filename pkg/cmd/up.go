package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rancher/dolly/pkg/portforward"
	"github.com/rancher/dolly/pkg/riofile"
	"github.com/rancher/dolly/pkg/template"
	"github.com/rancher/dolly/pkg/types"
	cli "github.com/rancher/wrangler-cli"
	"github.com/rancher/wrangler/pkg/gvk"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

func NewUpCommand() *cobra.Command {
	up := cli.Command(&Up{}, cobra.Command{
		Short: "Applying kubernetes application using riofile",
	})
	return up
}

type Up struct {
	File       string `name:"file" usage:"Path to riofile, can point to local file path, https links or stdin(-)" default:"Riofile" short:"f"`
	NoExpose   bool   `name:"no-expose" usage:"Whether to expose service port on localhost"`
	Namespace  string `name:"namespace" usage:"Namespace to install" default:"default" short:"n"`
	AnswerFile string `name:"answer-file" usage:"Answer file set for riofile" default:"Riofile-answers" short:"a"`
}

func (u *Up) Run(cmd *cobra.Command, args []string) error {
	rf, err := u.parseRioFileFile()
	if err != nil {
		return err
	}

	if err := u.do(rf); err != nil {
		return err
	}

	done := cmd.Context().Done()
	stop := make(chan struct{}, 1)
	toWatch := u.File
	if rf.NeedBuild() {
		toWatch = filepath.Dir(u.File)
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()
	go func() {
		go func() {
			for {
				select {
				// watch for events
				case event := <-watcher.Events:
					if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Rename == fsnotify.Rename {
						stop <- struct{}{}
						rf, err = u.parseRioFileFile()
						if err != nil {
							logrus.Errorf("Failed to parse riofile, error: %v", err)
						}
						if err := u.do(rf); err != nil {
							logrus.Errorf("Failed to apply riofile, error: %v", err)
						}
						if !u.NoExpose {
							go u.portForward(cmd.Context(), rf, stop)
						}
					}
					if err := watcher.Add(event.Name); err != nil {
						logrus.Debugf("Failed to watch %v, error: %v", event.Name, err)
					}

					// watch for errors
				case err := <-watcher.Errors:
					logrus.Debugf("Failed to watch directory %v, error: %v", toWatch, err)
				}
			}
		}()

		filepath.Walk(toWatch, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				if err := watcher.Add(path); err != nil {
					logrus.Errorf("Failed to watch directory %v, error: %v", toWatch, err)
				}
			}
			return nil
		})
		<-done
	}()

	go u.portForward(cmd.Context(), rf, stop)
	<-done
	return nil
}

func (u *Up) parseRioFileFile() (*riofile.Riofile, error) {
	content, answers, err := riofile.LoadFileAndAnswer(u.File, u.AnswerFile)
	if err != nil {
		return nil, err
	}

	return riofile.Parse(content, u.Namespace, template.AnswersFromMap(answers))
}

func (u *Up) do(rf *riofile.Riofile) error {
	if err := rf.Build(false); err != nil {
		return err
	}

	objects := rf.Objects()
	if err := printObjects(objects); err != nil {
		return err
	}

	return Apply.WithDynamicLookup().WithDefaultNamespace(u.Namespace).ApplyObjects(objects...)
}

func (u *Up) portForward(ctx context.Context, rf *riofile.Riofile, stop chan struct{}) {
	var pfOptions []portforward.Option
	for _, svc := range rf.Services {
		pod, err := u.waitForPod(svc)
		if err != nil {
			return
		}
		for _, container := range append([]types.NamedContainer{{Name: svc.Name, Container: svc.Spec.Container}}, svc.Spec.Sidecars...) {
			for _, port := range container.Ports {
				if port.Port != 0 && port.TargetPort != 0 {
					pfOptions = append(pfOptions, portforward.Option{
						Pod:        pod,
						Port:       strconv.Itoa(int(port.Port)),
						TargetPort: strconv.Itoa(int(port.TargetPort)),
						Stdout:     true,
						ReadyChan:  make(chan struct{}, 1),
						StopChan:   make(chan struct{}, 1),
					})
				}
			}
		}
	}

	k8s := kubernetes.NewForConfigOrDie(RestConfig)
	for _, option := range pfOptions {
		go func() {
			if err := portforward.PortForward(RestConfig, k8s, option); err != nil {
				logrus.Errorf("Failed to setup port-forward for pod %s, error: %v", option.Pod.Name, err)
			}
		}()
	}
	select {
	case <-stop:
		for _, option := range pfOptions {
			option.StopChan <- struct{}{}
		}
	}

	return
}

func (u *Up) waitForPod(svc types.Service) (v1.Pod, error) {
	for {
		pods, err := K8sInterface.CoreV1().Pods(u.Namespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%s", svc.Name),
			Limit:         1,
		})
		if err != nil {
			return v1.Pod{}, err
		}

		if len(pods.Items) == 0 {
			time.Sleep(time.Second * 1)
			continue
		}

		return pods.Items[0], nil
	}
}

func printObjects(objects []runtime.Object) error {
	for _, object := range objects {
		gvk, err := gvk.Get(object)
		if err != nil {
			return err
		}
		m, err := meta.Accessor(object)
		if err != nil {
			return err
		}
		fmt.Printf("%s/%s\n", strings.ToLower(gvk.GroupKind().String()), m.GetName())
	}
	return nil
}
