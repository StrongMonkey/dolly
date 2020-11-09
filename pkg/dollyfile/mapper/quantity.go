package mappers

import (
	"github.com/rancher/dolly/pkg/dollyfile/stringers"
	"github.com/rancher/wrangler/pkg/data"
	"github.com/rancher/wrangler/pkg/schemas"
	"github.com/rancher/wrangler/pkg/schemas/mappers"
)

type QuantityMapper struct {
	mappers.DefaultMapper
}

func NewQuantity(field string, args ...string) schemas.Mapper {
	return QuantityMapper{
		DefaultMapper: mappers.DefaultMapper{
			Field: field,
		},
	}
}

func (d QuantityMapper) ToInternal(data data.Object) error {
	v, ok := data[d.Field]
	if !ok {
		return nil
	}

	if s, ok := v.(string); ok {
		q, err := stringers.ParseQuantity(s)
		if err != nil {
			return err
		}
		if !q.IsZero() {
			data[d.Field], _ = q.AsInt64()
		}
	}

	return nil
}
