package volume

import (
	"github.com/rancher/dolly/pkg/dollyfile"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	RegistryStorageSize = "10G"
)

type Plugin struct{}

func (p Plugin) Convert(rf *dollyfile.DollyFile) (ret []runtime.Object) {
	for _, service := range rf.Services {
		volumes := service.Spec.Volumes

		for _, c := range service.Spec.Sidecars {
			for _, volume := range c.Volumes {
				volumes = append(volumes, volume)
			}
		}

		for _, v := range volumes {
			if v.Persistent {
				size := v.Size
				if size == "" {
					size = RegistryStorageSize
				}

				pv := &v1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      v.Name,
						Namespace: service.Namespace,
					},
					Spec: v1.PersistentVolumeClaimSpec{
						AccessModes: []v1.PersistentVolumeAccessMode{
							v1.ReadWriteOnce,
						},
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: resource.MustParse(size),
							},
						},
					},
				}
				ret = append(ret, pv)
			}
		}
	}

	return ret
}
