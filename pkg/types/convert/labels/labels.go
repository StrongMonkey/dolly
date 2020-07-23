package labels

import "github.com/rancher/dolly/pkg/types"

func SelectorLabels(service types.Service) map[string]string {
	app := service.Spec.App
	if app == "" {
		app = service.Name
	}
	return map[string]string{
		"app": app,
	}
}

func Merge(base map[string]string, overlay ...map[string]string) map[string]string {
	result := map[string]string{}
	for k, v := range base {
		result[k] = v
	}

	i := len(overlay)
	switch {
	case i == 1:
		for k, v := range overlay[0] {
			result[k] = v
		}
	case i > 1:
		result = Merge(Merge(base, overlay[0]), overlay[1:]...)
	}

	return result
}
