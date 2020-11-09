package dollyfile

import (
	"bytes"

	"github.com/rancher/dolly/pkg/template"
	"github.com/rancher/wrangler/pkg/data/convert"
	wyaml "github.com/rancher/wrangler/pkg/yaml"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
)

// Parse converts a textfile into a DollyFile struct
func Parse(contents []byte, namespace string, answers template.AnswerCallback) (*DollyFile, error) {
	k8s, objs, err := isK8SYaml(contents)
	if err != nil {
		return nil, err
	}

	if k8s {
		return &DollyFile{
			Kubernetes: objs,
		}, nil
	}

	data, err := parseData(contents, answers)
	if err != nil {
		return nil, err
	}

	s := Schema.Schema("DollyFile")
	if err := s.Mapper.ToInternal(data); err != nil {
		return nil, err
	}

	rf := &DollyFile{}
	if err := convert.ToObj(data, rf); err != nil {
		return nil, err
	}

	return renderK8sObject(rf, namespace)
}

func renderK8sObject(rf *DollyFile, namespace string) (*DollyFile, error) {
	if rf.Manifest != "" {
		objs, err := wyaml.ToObjects(bytes.NewBufferString(rf.Manifest))
		if err != nil {
			return nil, err
		}
		rf.Kubernetes = objs
	}

	for k, v := range rf.Configs {
		v.Name = k
		rf.Configs[k] = v
	}

	for k, v := range rf.Services {
		v.Name = k
		v.Namespace = namespace
		rf.Services[k] = v
	}
	return rf, nil
}

func isK8SYaml(contents []byte) (bool, []runtime.Object, error) {
	objs, err := wyaml.ToObjects(bytes.NewBuffer(contents))
	if err != nil {
		return false, nil, nil
	}
	if len(objs) > 0 &&
		objs[0].GetObjectKind().GroupVersionKind().Kind != "" {
		return true, objs, nil
	}
	return false, nil, nil
}

func parseData(contents []byte, answers template.AnswerCallback) (map[string]interface{}, error) {
	t := template.Template{
		Content: contents,
	}

	cont, err := t.Parse(answers)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{}
	if err := yaml.Unmarshal(cont, &data); err != nil {
		return nil, err
	}
	return data, nil
}
