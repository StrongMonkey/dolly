package tables

import (
	"encoding/base64"

	"github.com/docker/go-units"
	"github.com/rancher/dolly/pkg/table"
	v1 "k8s.io/api/core/v1"
)

func NewConfig(namespace, format string, quiet bool) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{id .Obj}}"},
		{"CREATED", "{{.Obj.CreationTimestamp | ago}}"},
		{"SIZE", "{{.Obj | size}}"},
	}, namespace, quiet, format)

	writer.AddFormatFunc("size", Base64Size)

	return &tableWriter{
		writer: writer,
	}
}

func Base64Size(data interface{}) (string, error) {
	c, ok := data.(*v1.ConfigMap)
	if !ok {
		return "", nil
	}

	size := len(c.Data) + len(c.BinaryData)
	if size > 0 {
		for _, v := range c.Data {
			size += len(v)
		}
		for _, v := range c.BinaryData {
			size += len(base64.StdEncoding.EncodeToString(v))
		}
	}

	return units.HumanSize(float64(size)), nil
}
