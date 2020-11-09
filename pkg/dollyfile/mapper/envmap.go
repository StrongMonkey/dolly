package mappers

import (
	"fmt"
	"strings"

	"github.com/rancher/wrangler/pkg/data"
	"github.com/rancher/wrangler/pkg/schemas"
	"github.com/rancher/wrangler/pkg/schemas/mappers"
)

type EnvMapper struct {
	mappers.DefaultMapper
	Sep string
}

func NewEnvMap(field string, opts ...string) schemas.Mapper {
	e := EnvMapper{
		DefaultMapper: mappers.DefaultMapper{
			Field: field,
		},
	}

	for _, opt := range opts {
		if strings.HasPrefix(opt, "sep=") {
			e.Sep = strings.TrimPrefix(opt, "sep=")
		}
	}

	return e
}

func (e EnvMapper) ToInternal(data data.Object) error {
	m := data.Map(e.Field)
	if m == nil {
		return nil
	}

	var result []interface{}
	for k, v := range m {
		item := fmt.Sprintf("%s%s%s", k, e.Sep, v)
		result = append(result, item)
	}

	data.Set(e.Field, result)
	return nil
}
