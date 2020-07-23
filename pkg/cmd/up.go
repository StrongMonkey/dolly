package cmd

import (
	"fmt"
	"strings"

	"github.com/rancher/dolly/pkg/riofile"
	"github.com/rancher/dolly/pkg/template"
	cli "github.com/rancher/wrangler-cli"
	"github.com/rancher/wrangler/pkg/gvk"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewUpCommand() *cobra.Command {
	up := cli.Command(&Up{}, cobra.Command{
		Short: "Applying kubernetes application using riofile",
	})
	return up
}

type Up struct {
	File       string `name:"file" usage:"Path to riofile, can point to local file path, https links or stdin(-)" default:"Riofile" short:"f"`
	Namespace  string `name:"namespace" usage:"Namespace to install" default:"default" short:"n"`
	AnswerFile string `name:"answer-file" usage:"Answer file set for riofile" default:"Riofile-answers" short:"a"`
}

func (u *Up) Run(cmd *cobra.Command, args []string) error {
	content, answers, err := riofile.LoadFileAndAnswer(u.File, u.AnswerFile)
	if err != nil {
		return err
	}

	rf, err := riofile.Parse(content, u.Namespace, template.AnswersFromMap(answers))
	if err != nil {
		return err
	}

	if err := rf.Build(false); err != nil {
		return err
	}

	objects := rf.Objects()
	if err := printObjects(objects); err != nil {
		return err
	}

	return Apply.WithDynamicLookup().WithDefaultNamespace(u.Namespace).ApplyObjects(objects...)
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
