package mappers

import (
	"github.com/rancher/wrangler/pkg/schemas"
	"github.com/rancher/wrangler/pkg/schemas/mappers"
)

func NewObject() schemas.Mapper {
	return schemas.Mappers{
		mappers.Drop{
			Optional: true,
			Field:    "status",
		},
		&mappers.Embed{
			Field:    "spec",
			Optional: true,
		},
	}
}
