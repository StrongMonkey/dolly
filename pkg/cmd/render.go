package cmd

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/rancher/dolly/pkg/riofile"
	"github.com/rancher/dolly/pkg/template"
	cli "github.com/rancher/wrangler-cli"
	gvk2 "github.com/rancher/wrangler/pkg/gvk"
	"github.com/rancher/wrangler/pkg/yaml"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	chartYAMLTemplate = ` 
apiVersion: v2
name: %s
version: %s`
)

func NewRenderCommand() *cobra.Command {
	render := cli.Command(&Render{}, cobra.Command{
		Short: "Creating helm charts based on riofile",
	})
	return render
}

type Render struct {
	File       string `name:"file" usage:"Path to riofile, can point to local file path, https links or stdin(-)" default:"Riofile" short:"f"`
	Namespace  string `name:"namespace" usage:"Namespace to install" default:"default" short:"n"`
	AnswerFile string `name:"answer-file" usage:"Answer file set for riofile" default:"Riofile-answers" short:"a"`
	Version    string `name:"version" usage:"Helm chart version to create" short:"v"`
	ChartName  string `name:"chart-name" usage:"Chart name to be created"`
	// todo: add more parameters
}

func (r Render) Run(cmd *cobra.Command, args []string) error {
	if r.ChartName == "" {
		return fmt.Errorf("chartName is required. Use --chart-name")
	}

	if r.Version == "" {
		return fmt.Errorf("helm version is required. Use --version")
	}

	content, answers, err := riofile.LoadFileAndAnswer(r.File, r.AnswerFile)
	if err != nil {
		return err
	}

	rf, err := riofile.Parse(content, r.Namespace, template.AnswersFromMap(answers))
	if err != nil {
		return err
	}

	if err := rf.Build(true); err != nil {
		return err
	}

	return r.renderHelmCharts(rf.Objects())
}

func (r Render) renderHelmCharts(objects []runtime.Object) error {
	objectsGroupedByGVK := map[schema.GroupVersionKind][]runtime.Object{}
	for _, object := range objects {
		gvk, err := gvk2.Get(object)
		if err != nil {
			return err
		}
		if objectsGroupedByGVK[gvk] == nil {
			objectsGroupedByGVK[gvk] = []runtime.Object{}
		}

		objectsGroupedByGVK[gvk] = append(objectsGroupedByGVK[gvk], object)
	}

	tmpdir, err := ioutil.TempDir("", "helm-chart-rio")
	if err != nil {
		return err
	}

	chartFolder := filepath.Join(tmpdir, r.ChartName)

	if err := os.Mkdir(chartFolder, 0755); err != nil {
		return err
	}

	if err := os.Mkdir(filepath.Join(chartFolder, "templates"), 0755); err != nil {
		return err
	}

	chartYamlFilename := filepath.Join(chartFolder, "Chart.yaml")
	chartFile, err := os.OpenFile(chartYamlFilename, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(chartFile, chartYAMLTemplate, r.ChartName, r.Version); err != nil {
		return err
	}

	for gvk, objects := range objectsGroupedByGVK {
		filename := filepath.Join(chartFolder, "templates", strings.ToLower(fmt.Sprintf("%s.yaml", gvk.Kind)))
		data, err := yaml.Export(objects...)
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(filename, data, 0755); err != nil {
			return err
		}
	}
	tarFilename := fmt.Sprintf("%s-%s.tar.gz", r.ChartName, r.Version)

	output, err := os.OpenFile(tarFilename, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	gw := gzip.NewWriter(output)
	tw := tar.NewWriter(gw)

	filepath.Walk(chartFolder, func(path string, info os.FileInfo, err error) error {
		header, err := tar.FileInfoHeader(info, path)
		if err != nil {
			return err
		}
		relativePath, err := filepath.Rel(tmpdir, path)
		if err != nil {
			return err
		}
		header.Name = relativePath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, file); err != nil {
				return err
			}
		}
		return nil
	})

	if err := tw.Close(); err != nil {
		return err
	}

	if err := gw.Close(); err != nil {
		return err
	}
	fmt.Printf("tarball %s created\n", tarFilename)
	return nil
}
