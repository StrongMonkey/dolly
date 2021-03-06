package mappers

import (
	"github.com/mattn/go-shellwords"
	"github.com/rancher/wrangler/pkg/data"
	"github.com/rancher/wrangler/pkg/data/convert"
	"github.com/rancher/wrangler/pkg/schemas"
	"github.com/rancher/wrangler/pkg/schemas/mappers"
)

type Shlex struct {
	mappers.DefaultMapper
}

func NewShlex(field string, _ ...string) schemas.Mapper {
	return &Shlex{
		DefaultMapper: mappers.DefaultMapper{
			Field: field,
		},
	}
}

func (d Shlex) FromInternal(data data.Object) {
	v, ok := data[d.Field]
	if !ok {
		return
	}

	parts := convert.ToStringSlice(v)
	if len(parts) == 1 {
		data[d.Field] = parts[0]
	}
}

func (d Shlex) ToInternal(data data.Object) error {
	v, ok := data[d.Field]
	if !ok {
		return nil
	}

	if str, ok := v.(string); ok {
		parts, err := shellwords.Parse(str)
		if err != nil {
			return err
		}
		data[d.Field] = parts
	}

	return nil
}
