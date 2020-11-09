package dollyfile

import (
	riofilemapper "github.com/rancher/dolly/pkg/dollyfile/mapper"
	"github.com/rancher/dolly/pkg/dollyfile/stringers"
	"github.com/rancher/dolly/pkg/types"
	"github.com/rancher/wrangler/pkg/schemas"
	m "github.com/rancher/wrangler/pkg/schemas/mappers"
	corev1 "k8s.io/api/core/v1"
)

var (
	Schema = schemas.EmptySchemas()
)

func init() {
	Schema.DefaultPostMapper = func() schemas.Mapper {
		return m.JSONKeys{}
	}

	Schema.
		Init(mappers).
		Init(services).
		Init(configs).
		TypeName("DollyFile", DollyFile{}).
		MustImport(DollyFile{})
}

func mappers(schemas *schemas.Schemas) *schemas.Schemas {
	return objectToSlice(schemas).
		AddFieldMapper("alias", m.NewAlias).
		AddFieldMapper("duration", riofilemapper.NewDuration).
		AddFieldMapper("quantity", riofilemapper.NewQuantity).
		AddFieldMapper("enum", m.NewEnum).
		AddFieldMapper("hostNetwork", riofilemapper.NewHostNetwork).
		AddFieldMapper("envmap", riofilemapper.NewEnvMap).
		AddFieldMapper("shlex", riofilemapper.NewShlex)
}

func objectToSlice(schemas *schemas.Schemas) *schemas.Schemas {
	schemas.AddFieldMapper("configs", riofilemapper.NewObjectsToSliceFactory(
		func() riofilemapper.MaybeStringer {
			return &stringers.ConfigsStringer{}
		},
		func(str string) (interface{}, error) {
			return stringers.ParseConfig(str)
		}))
	schemas.AddFieldMapper("secrets", riofilemapper.NewObjectsToSliceFactory(
		func() riofilemapper.MaybeStringer {
			return &stringers.SecretsStringer{}
		},
		func(str string) (interface{}, error) {
			return stringers.ParseSecret(str)
		}))
	schemas.AddFieldMapper("dnsOptions", riofilemapper.NewObjectsToSliceFactory(
		func() riofilemapper.MaybeStringer {
			return &stringers.PodDNSConfigOptionStringer{}
		},
		func(str string) (interface{}, error) {
			return stringers.ParseDNSOptions(str)
		}))
	schemas.AddFieldMapper("env", riofilemapper.NewObjectsToSliceFactory(
		func() riofilemapper.MaybeStringer {
			return &stringers.EnvStringer{}
		},
		func(str string) (interface{}, error) {
			return stringers.ParseEnv(str)
		}))
	schemas.AddFieldMapper("ports", riofilemapper.NewObjectsToSliceFactory(
		func() riofilemapper.MaybeStringer {
			return &stringers.ContainerPortStringer{}
		},
		func(str string) (interface{}, error) {
			return stringers.ParsePort(str)
		}))
	schemas.AddFieldMapper("hosts", riofilemapper.NewObjectsToSliceFactory(
		func() riofilemapper.MaybeStringer {
			return &stringers.HostAliasStringer{}
		},
		func(str string) (interface{}, error) {
			return stringers.ParseHostAlias(str)
		}))
	schemas.AddFieldMapper("volumes", riofilemapper.NewObjectsToSliceFactory(
		func() riofilemapper.MaybeStringer {
			return &stringers.VolumeStringer{}
		},
		func(str string) (interface{}, error) {
			return stringers.ParseVolume(str)
		}))
	schemas.AddFieldMapper("permissions", riofilemapper.NewObjectsToSliceFactory(
		func() riofilemapper.MaybeStringer {
			return &stringers.PermissionStringer{}
		},
		func(str string) (interface{}, error) {
			return stringers.ParsePermission(str)
		}))

	return schemas
}

func configs(schemas *schemas.Schemas) *schemas.Schemas {
	schemas.AddMapperForType(corev1.ConfigMap{},
		riofilemapper.NewObject(),
		riofilemapper.NewConfigMapMapper("data"))
	return schemas
}

func services(schemas *schemas.Schemas) *schemas.Schemas {
	schemas.AddMapperForType(types.Service{},
		riofilemapper.NewObject())
	return schemas
}
