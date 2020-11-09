package mappers

import (
	"github.com/rancher/wrangler/pkg/data"
	s "github.com/rancher/wrangler/pkg/schemas"
)

type HostNetwork struct {
	Field string
}

func NewHostNetwork(field string, _ ...string) s.Mapper {
	return HostNetwork{Field: field}
}

func (d HostNetwork) FromInternal(data data.Object) {
	if _, ok := data[d.Field]; ok {
		delete(data, d.Field)
		data["net"] = "host"
	}
}

func (d HostNetwork) ToInternal(data data.Object) error {
	if net, ok := data["net"]; ok && net == "host" {
		delete(data, "net")
		data[d.Field] = true
	}
	return nil
}

func (d HostNetwork) ModifySchema(schema *s.Schema, schemas *s.Schemas) error {
	schema.ResourceFields["net"] = s.Field{}
	return nil
}
