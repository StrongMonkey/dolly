package tables

import "github.com/rancher/dolly/pkg/table"

func NewSecret(namespace, format string, quiet bool) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{id .Obj}}"},
		{"TYPE", "{{.Obj.Type}}"},
		{"DATA", "{{.Obj.Data | len}}"},
		{"CREATED", "{{.Obj.CreationTimestamp | ago}}"},
	}, namespace, quiet, format)

	return &tableWriter{
		writer: writer,
	}
}
