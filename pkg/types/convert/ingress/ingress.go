package ingress

import (
	"github.com/rancher/dolly/pkg/dollyfile"
	"github.com/rancher/dolly/pkg/types/utils"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Plugin struct{}

func (p Plugin) Convert(rf *dollyfile.DollyFile) (ret []runtime.Object) {
	for _, service := range rf.Services {
		app := service.Spec.App
		if app == "" {
			app = service.Name
		}

		var servicePort int32
		for _, port := range utils.ContainerPorts(service) {
			if port.IsExposed() && port.IsHTTP() {
				servicePort = port.Port
				continue
			}
		}
		if servicePort == 0 {
			return nil
		}

		hostnames := service.Spec.Hostnames
		if len(hostnames) == 0 {
			continue
		}

		ingress := &v1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      app,
				Namespace: service.Namespace,
			},
		}

		pathType := v1.PathTypeImplementationSpecific
		for _, hostname := range hostnames {
			ingress.Spec.Rules = append(ingress.Spec.Rules, v1.IngressRule{
				Host: hostname,
				IngressRuleValue: v1.IngressRuleValue{
					HTTP: &v1.HTTPIngressRuleValue{
						Paths: []v1.HTTPIngressPath{
							{
								PathType: &pathType,
								Backend: v1.IngressBackend{
									Service: &v1.IngressServiceBackend{
										Name: app,
										Port: v1.ServiceBackendPort{
											Number: servicePort,
										},
									},
								},
							},
						},
					},
				},
			})
			ingress.Spec.TLS = append(ingress.Spec.TLS, v1.IngressTLS{
				Hosts:      hostnames,
				SecretName: service.Spec.TLS,
			})
		}
		ret = append(ret, ingress)
	}

	for _, router := range rf.Routes {
		ingress := &v1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      router.Name,
				Namespace: router.Namespace,
			},
		}
		for _, routeSpec := range router.Spec.Routes {
			pathMatch := ""
			pathMatchType := v1.PathTypeExact
			if routeSpec.Match.Path != nil {
				if routeSpec.Match.Path.Prefix != "" {
					pathMatch = routeSpec.Match.Path.Prefix
					pathMatchType = v1.PathTypePrefix
				} else if routeSpec.Match.Path.Exact != "" {
					pathMatch = routeSpec.Match.Path.Exact
					pathMatchType = v1.PathTypeExact
				} else {
					pathMatch = routeSpec.Match.Path.Regexp
					pathMatchType = v1.PathTypeImplementationSpecific
				}
			}
			for _, hostname := range router.Spec.Hostnames {
				for _, to := range routeSpec.To {
					ingress.Spec.Rules = append(ingress.Spec.Rules, v1.IngressRule{
						Host: hostname,
						IngressRuleValue: v1.IngressRuleValue{
							HTTP: &v1.HTTPIngressRuleValue{
								Paths: []v1.HTTPIngressPath{
									{
										Path:     pathMatch,
										PathType: &pathMatchType,
										Backend: v1.IngressBackend{
											Service: &v1.IngressServiceBackend{
												Name: to.App,
												Port: v1.ServiceBackendPort{
													Number: int32(to.Port),
												},
											},
										},
									},
								},
							},
						},
					})
				}
			}
		}

		if router.Spec.Secret != "" {
			ingress.Spec.TLS = []v1.IngressTLS{
				{
					Hosts:      router.Spec.Hostnames,
					SecretName: router.Spec.Secret,
				},
			}
		}
		ret = append(ret, ingress)
	}

	return ret
}
