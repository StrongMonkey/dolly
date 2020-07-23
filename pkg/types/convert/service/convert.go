package service

import (
	"fmt"
	"strings"

	"github.com/rancher/dolly/pkg/types"
	"github.com/rancher/dolly/pkg/types/convert/labels"
	"github.com/rancher/dolly/pkg/types/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ConvertService converts a rio service to k8s service
func Convert(service types.Service) *v1.Service {
	svc := newServiceSelector(service.Name, service.Namespace, v1.ServiceTypeClusterIP, service.Labels, labels.SelectorLabels(service))
	if len(serviceNamedPorts(service)) > 0 {
		svc.Spec.Ports = serviceNamedPorts(service)
	}

	return svc
}

func newServiceSelector(name, namespace string, serviceType v1.ServiceType, labels, selectorLabels map[string]string) *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:    labels,
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.ServiceSpec{
			Type:     serviceType,
			Selector: selectorLabels,
			Ports: []v1.ServicePort{
				{
					Name:       "default",
					Protocol:   v1.ProtocolTCP,
					TargetPort: intstr.FromInt(80),
					Port:       80,
				},
			},
		},
	}
}

func serviceNamedPorts(service types.Service) (servicePorts []v1.ServicePort) {
	for _, port := range utils.ContainerPorts(service) {
		servicePort := v1.ServicePort{
			Name:     port.Name,
			Port:     port.Port,
			Protocol: utils.Protocol(port.Protocol),
			TargetPort: intstr.IntOrString{
				IntVal: port.TargetPort,
			},
		}

		if servicePort.Name == "" {
			servicePort.Name = strings.ToLower(fmt.Sprintf("%s-%d", port.Protocol, port.Port))
		}

		servicePorts = append(servicePorts, servicePort)
	}

	return
}


