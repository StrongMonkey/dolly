package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rancher/dolly/pkg/rdns"

	"github.com/fsnotify/fsnotify"
	"github.com/rancher/dolly/pkg/dollyfile"
	"github.com/rancher/dolly/pkg/log"
	"github.com/rancher/dolly/pkg/portforward"
	"github.com/rancher/dolly/pkg/template"
	"github.com/rancher/dolly/pkg/types/convert/deployment"
	"github.com/rancher/dolly/pkg/types/convert/ingress"
	"github.com/rancher/dolly/pkg/types/convert/rbac"
	"github.com/rancher/dolly/pkg/types/convert/service"
	"github.com/rancher/dolly/pkg/types/convert/volume"
	cli "github.com/rancher/wrangler-cli"
	"github.com/rancher/wrangler/pkg/gvk"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stern/stern/stern"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

func NewUpCommand() *cobra.Command {
	up := cli.Command(&Up{}, cobra.Command{
		Short: "Applying kubernetes application using dollyfile",
	})
	return up
}

type Up struct {
	File       string `name:"file" usage:"Path to dollyfile, can point to local file path, https links or stdin(-)" default:"DollyFile" short:"f"`
	NoExpose   bool   `name:"no-expose" usage:"Whether to expose service port on localhost"`
	NoWatch    bool   `name:"no-watch" usage:"Whether to watch dollyfile and apply changes"`
	Namespace  string `name:"namespace" u2sage:"Namespace to install" default:"default" short:"n"`
	AnswerFile string `name:"answer-file" usage:"Answer file set for dollyfile" default:"DollyFile-answers" short:"a"`
}

func (u *Up) Run(cmd *cobra.Command, args []string) error {
	if err := u.setupRDNS(cmd.Context()); err != nil {
		return err
	}

	rf, err := u.parseDollyFile()
	if err != nil {
		return err
	}

	if err := u.do(rf); err != nil {
		return err
	}

	if !u.NoExpose {
		go u.portForward(cmd.Context(), rf)
	}

	if !u.NoWatch {
		ctx, _ := context.WithCancel(cmd.Context())
		go u.Watch(ctx, rf)
	}

	return u.Log(cmd.Context(), rf)
}

func (u *Up) setupRDNS(ctx context.Context) (err error) {
	RdnsDomain, err = rdns.GetDomain(ctx, K8sInterface)
	return err
}

func (u *Up) Watch(ctx context.Context, rf *dollyfile.DollyFile) {
	toWatch := u.File
	if rf.NeedBuild() {
		toWatch = filepath.Dir(u.File)
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logrus.Error(err)
	}
	defer watcher.Close()
	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Rename == fsnotify.Rename {
					rf, err = u.parseDollyFile()
					if err != nil {
						logrus.Errorf("Failed to parse dollyfile, error: %v", err)
					}
					if err := u.do(rf); err != nil {
						logrus.Errorf("Failed to apply dollyfile, error: %v", err)
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
	<-ctx.Done()
	return
}

func (u *Up) parseDollyFile() (*dollyfile.DollyFile, error) {
	content, answers, err := dollyfile.LoadFileAndAnswer(u.File, u.AnswerFile)
	if err != nil {
		return nil, err
	}

	rf, err := dollyfile.Parse(content, u.Namespace, template.AnswersFromMap(answers))
	if err != nil {
		return nil, err
	}

	for k, svc := range rf.Services {
		svc.Spec.Hostnames = append(svc.Spec.Hostnames, fmt.Sprintf("%s-%s.%s", svc.Name, svc.Namespace, RdnsDomain))
		rf.Services[k] = svc
	}

	rf.Plugins = []dollyfile.Plugin{
		deployment.Plugin{},
		service.Plugin{},
		rbac.Plugin{},
		volume.Plugin{},
		ingress.Plugin{},
	}
	return rf, nil
}

func (u *Up) do(rf *dollyfile.DollyFile) error {
	if err := rf.Build(false); err != nil {
		return err
	}

	objects := rf.Objects()
	if err := printObjects(objects); err != nil {
		return err
	}

	return Apply.WithDynamicLookup().WithDefaultNamespace(u.Namespace).ApplyObjects(objects...)
}

func (u *Up) portForward(ctx context.Context, rf *dollyfile.DollyFile) {
	for _, svc := range rf.Services {
		fmt.Printf("http://%s-%s.%s:9080 ----> %s\n", svc.Name, svc.Namespace, RdnsDomain, svc.Name)
	}

	pod, err := u.waitForIngressPod()
	if err != nil {
		return
	}
	pfOption := portforward.Option{
		Pod:        pod,
		Port:       strconv.Itoa(9080),
		TargetPort: strconv.Itoa(80),
		Stdout:     Debug == true,
		ReadyChan:  make(chan struct{}, 1),
		StopChan:   make(chan struct{}, 1),
	}

	k8s := kubernetes.NewForConfigOrDie(RestConfig)
	if err := portforward.PortForward(RestConfig, k8s, pfOption); err != nil {
		logrus.Errorf("Failed to setup port-forward for ingress pod, error: %v", err)
	}

	return
}

func (u *Up) waitForIngressPod() (v1.Pod, error) {
	// this code assumes traefik ingress in k3s
	for {
		pods, err := K8sInterface.CoreV1().Pods("kube-system").List(context.Background(), metav1.ListOptions{
			LabelSelector: "app=traefik",
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

func (u *Up) Log(ctx context.Context, rf *dollyfile.DollyFile) error {
	var selectors []string
	for _, svc := range rf.Services {
		selectors = append(selectors, fmt.Sprintf("app=%v", svc.Name))
	}
	labelSelector, err := metav1.ParseToLabelSelector(strings.Join(selectors, ","))
	if err != nil {
		return err
	}
	ls, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		return err
	}

	template, err := log.Format("", false)
	if err != nil {
		return err
	}

	podQuery, err := regexp.Compile("")
	if err != nil {
		return err
	}

	containerQuery, err := regexp.Compile("")
	if err != nil {
		return err
	}

	config := &stern.Config{
		ContainerQuery: containerQuery,
		PodQuery:       podQuery,
		LabelSelector:  ls,
		Namespace:      u.Namespace,
		InitContainers: true,
		Template:       template,
		TailLines:      &[]int64{200}[0],
		Since:          time.Hour * 24,
		ContainerState: stern.RUNNING,
	}

	return log.Output(ctx, config, u.Namespace, K8sInterface)
}

func printObjects(objects []runtime.Object) error {
	if !Debug {
		return nil
	}
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
