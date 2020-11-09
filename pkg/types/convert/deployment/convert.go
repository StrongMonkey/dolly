package deployment

import (
	"sort"

	"github.com/rancher/dolly/pkg/dollyfile"
	"github.com/rancher/dolly/pkg/types"
	"github.com/rancher/dolly/pkg/types/convert/labels"
	"github.com/rancher/dolly/pkg/types/utils"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Plugin struct{}

func (p Plugin) Convert(rf *dollyfile.DollyFile) (ret []runtime.Object) {
	for _, svc := range rf.Services {
		podTemplateSpec := populatePodTemplate(svc)

		cp := newControllerParams(svc, podTemplateSpec)
		if svc.Spec.Global {
			ret = append(ret, daemonset(svc, cp))
		} else if len(cp.VolumeTemplates) > 0 {
			ret = append(ret, statefulset(svc, cp))
		} else {
			ret = append(ret, deployment(svc, cp))
		}
	}
	return ret
}

func statefulset(service types.Service, cp *controllerParams) runtime.Object {
	ss := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   service.Namespace,
			Name:        service.Name,
			Labels:      cp.Labels,
			Annotations: cp.Annotations,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: nil,
			Selector: &metav1.LabelSelector{
				MatchLabels: cp.SelectorLabels,
			},
			Template:             cp.PodTemplateSpec,
			VolumeClaimTemplates: volumeClaimTemplates(cp.VolumeTemplates),
			ServiceName:          service.Name,
			PodManagementPolicy:  appsv1.ParallelPodManagement,
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
		},
	}
	return ss
}

func deployment(service types.Service, cp *controllerParams) runtime.Object {
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        service.Name,
			Namespace:   service.Namespace,
			Labels:      cp.Labels,
			Annotations: cp.Annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: cp.Scale.Scale,
			Selector: &metav1.LabelSelector{
				MatchLabels: cp.SelectorLabels,
			},
			Template: cp.PodTemplateSpec,
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: cp.Scale.MaxUnavailable,
					MaxSurge:       cp.Scale.MaxSurge,
				},
			},
		},
	}

	return dep
}

func daemonset(service types.Service, cp *controllerParams) runtime.Object {
	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   service.Namespace,
			Name:        service.Name,
			Labels:      cp.Labels,
			Annotations: cp.Annotations,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: cp.SelectorLabels,
			},
			Template: cp.PodTemplateSpec,
			UpdateStrategy: appsv1.DaemonSetUpdateStrategy{
				Type: appsv1.RollingUpdateDaemonSetStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDaemonSet{
					MaxUnavailable: cp.Scale.MaxUnavailable,
				},
			},
		},
	}
	return ds
}

func volumeClaimTemplates(templates map[string]types.VolumeTemplate) (result []v1.PersistentVolumeClaim) {
	var names []string
	for name := range templates {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		template := templates[name]
		q := resource.NewQuantity(template.StorageRequest, resource.BinarySI)
		result = append(result, v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "vol-" + name,
				Labels:      template.Labels,
				Annotations: template.Annotations,
			},
			Spec: v1.PersistentVolumeClaimSpec{
				AccessModes: template.AccessModes,
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceStorage: *q,
					},
				},
				StorageClassName: &template.StorageClassName,
				VolumeMode:       template.VolumeMode,
			},
		})
	}

	return
}

func newControllerParams(service types.Service, podTemplateSpec v1.PodTemplateSpec) *controllerParams {
	scaleParams := parseScaleParams(service)
	selectorLabels := labels.SelectorLabels(service)
	volumeTemplates := NormalizeVolumeTemplates(service)

	// Selector labels must be on the podTemplateSpec
	podTemplateSpec.Labels = labels.Merge(podTemplateSpec.Labels, selectorLabels)

	return &controllerParams{
		Scale:           scaleParams,
		Labels:          service.Labels,
		Annotations:     service.Annotations,
		SelectorLabels:  selectorLabels,
		PodTemplateSpec: podTemplateSpec,
		VolumeTemplates: volumeTemplates,
	}
}

type controllerParams struct {
	Scale           scaleParams
	Labels          map[string]string
	Annotations     map[string]string
	SelectorLabels  map[string]string
	VolumeTemplates map[string]types.VolumeTemplate
	PodTemplateSpec v1.PodTemplateSpec
}

func populatePodTemplate(service types.Service) v1.PodTemplateSpec {
	pts := v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      service.Labels,
			Annotations: service.Annotations,
		},
	}

	podSpec := podSpec(service)

	pts.Spec = podSpec
	return pts
}

var (
	f             = false
	t             = true
	defaultCPU    = resource.MustParse("50m")
	defaultMemory = resource.MustParse("64Mi")
)

func podSpec(service types.Service) v1.PodSpec {
	podSpec := v1.PodSpec{
		DNSConfig:          podDNS(service),
		HostAliases:        service.Spec.HostAliases,
		Hostname:           service.Spec.Hostname,
		HostNetwork:        service.Spec.HostNetwork,
		EnableServiceLinks: &f,
		Containers:         containers(service, false),
		InitContainers:     containers(service, true),
		Volumes:            volumes(service),
		Affinity:           service.Spec.Affinity,
		ImagePullSecrets:   pullSecrets(service.Spec.ImagePullSecrets),
	}

	serviceAccountName := utils.ServiceAccountName(service)
	if serviceAccountName != "" {
		podSpec.ServiceAccountName = serviceAccountName
		podSpec.AutomountServiceAccountToken = nil
	}

	if service.Spec.DNS != nil {
		podSpec.DNSPolicy = service.Spec.DNS.Policy
	}

	if podSpec.ServiceAccountName == "" {
		podSpec.AutomountServiceAccountToken = &f
	} else {
		podSpec.AutomountServiceAccountToken = &t
	}

	return podSpec
}

func pullSecrets(names []string) (result []v1.LocalObjectReference) {
	for _, name := range names {
		result = append(result, v1.LocalObjectReference{
			Name: name,
		})
	}
	return
}

func podDNS(service types.Service) *v1.PodDNSConfig {
	if service.Spec.DNS == nil {
		return nil
	}

	if len(service.Spec.DNS.Options) == 0 &&
		len(service.Spec.DNS.Nameservers) == 0 &&
		len(service.Spec.DNS.Searches) == 0 {
		return nil
	}

	var options []v1.PodDNSConfigOption
	for _, opt := range service.Spec.DNS.Options {
		options = append(options, v1.PodDNSConfigOption{
			Name:  opt.Name,
			Value: opt.Value,
		})
	}
	return &v1.PodDNSConfig{
		Options:     options,
		Nameservers: service.Spec.DNS.Nameservers,
		Searches:    service.Spec.DNS.Searches,
	}
}
