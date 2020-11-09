package mappers

import (
	"github.com/rancher/wrangler/pkg/data"
	"github.com/rancher/wrangler/pkg/data/convert"
	"github.com/rancher/wrangler/pkg/schemas"
	"github.com/rancher/wrangler/pkg/schemas/mappers"
)

type ConfigMapMapper struct {
	mappers.DefaultMapper
}

func NewConfigMapMapper(field string) schemas.Mapper {
	return ConfigMapMapper{
		DefaultMapper: mappers.DefaultMapper{
			Field: field,
		},
	}
}

func (d ConfigMapMapper) FromInternal(data data.Object) {
	newData, ok := data[d.Field]
	if !ok {
		return
	}

	delete(data, d.Field)
	for k, v := range convert.ToMapInterface(newData) {
		data[k] = v
	}
}

func (d ConfigMapMapper) ToInternal(data data.Object) error {
	newData := map[string]interface{}{}
	for k, v := range data {
		if k != "labels" && v != "annotations" {
			delete(data, k)
			newData[k] = v
		}
	}

	if len(newData) > 0 {
		data[d.Field] = newData
	}

	return nil
}
