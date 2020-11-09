package utils

import (
	"fmt"

	"github.com/rancher/dolly/pkg/dollyfile/stringers"
	"github.com/rancher/dolly/pkg/types"
	v1 "k8s.io/api/core/v1"
)

func ToNamedContainers(service types.Service) (result []types.NamedContainer) {
	if containerIsValid(service.Spec.Container) {
		result = append(result, types.NamedContainer{
			Name:      rootContainerName(service),
			Container: service.Spec.Container,
		})
	}

	result = append(result, service.Spec.Sidecars...)
	return
}

func rootContainerName(service types.Service) string {
	return service.Name
}

func containerIsValid(container types.Container) bool {
	return container.Image != "" || container.Build != nil
}

func Protocol(proto types.Protocol) (protocol v1.Protocol) {
	switch proto {
	case types.ProtocolUDP:
		protocol = v1.ProtocolUDP
	case types.ProtocolSCTP:
		protocol = v1.ProtocolSCTP
	default:
		protocol = v1.ProtocolTCP
	}
	return
}

func ServiceAccountName(service types.Service) string {
	if len(service.Spec.Permissions) == 0 && len(service.Spec.GlobalPermissions) == 0 {
		return ""
	}
	return service.Name
}

func ContainerPorts(service types.Service) []types.ContainerPort {
	var (
		ports   []types.ContainerPort
		portMap = map[string]bool{}
	)

	for _, container := range ToNamedContainers(service) {
		for _, port := range container.Ports {
			port = stringers.NormalizeContainerPort(port)

			if port.Port == 0 {
				continue
			}

			key := fmt.Sprintf("%v/%v", port.Port, port.Protocol)
			if portMap[key] {
				continue
			}
			portMap[key] = true

			ports = append(ports, port)
		}
	}

	return ports
}
