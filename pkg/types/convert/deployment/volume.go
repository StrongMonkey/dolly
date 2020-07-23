package deployment

import (
    "fmt"
    "github.com/rancher/dolly/pkg/types"
    "github.com/rancher/dolly/pkg/types/utils"
    v1 "k8s.io/api/core/v1"
    "sort"
)

func secretVolumes(containers []types.NamedContainer) (result []v1.Volume) {
    var names []string
    for _, container := range containers {
        for _, mount := range container.Secrets {
            if mount.Name != "" {
                names = append(names, mount.Name)
            }
        }
    }

    for _, secret := range removeDuplicateAndSort(names) {
        result = append(result, v1.Volume{
            Name: fmt.Sprintf("secret-%s", secret),
            VolumeSource: v1.VolumeSource{
                Secret: &v1.SecretVolumeSource{
                    SecretName: secret,
                    Optional:   &[]bool{true}[0],
                },
            },
        })
    }

    return
}

func configVolumes(containers []types.NamedContainer) (result []v1.Volume) {
    var names []string
    for _, container := range containers {
        for _, mount := range container.Configs {
            if mount.Name != "" {
                names = append(names, mount.Name)
            }
        }
    }

    for _, config := range removeDuplicateAndSort(names) {
        result = append(result, v1.Volume{
            Name: fmt.Sprintf("config-%s", config),
            VolumeSource: v1.VolumeSource{
                ConfigMap: &v1.ConfigMapVolumeSource{
                    LocalObjectReference: v1.LocalObjectReference{
                        Name: config,
                    },
                },
            },
        })
    }

    return
}

type sortedVolumes struct {
    All      map[string]bool
    EmptyDir map[string]types.Volume
    HostPath map[string]types.Volume
    PVC      map[string]types.Volume
}

func sortVolumes(containers []types.NamedContainer, volumeTemplates map[string]types.VolumeTemplate) (result sortedVolumes) {
    result.All = map[string]bool{}
    result.EmptyDir = map[string]types.Volume{}
    result.HostPath = map[string]types.Volume{}
    result.PVC = map[string]types.Volume{}

    for _, container := range containers {
        for _, volume := range normalizeVolumes(container.Name, container.Volumes) {
            if result.All[volume.Name] {
                continue
            }
            result.All[volume.Name] = true
            if volume.HostPath != "" {
                result.HostPath[volume.Name] = volume
            } else if _, ok := volumeTemplates[volume.Name]; !ok && volume.Persistent {
                result.PVC[volume.Name] = volume
            } else if !volume.Persistent {
                result.EmptyDir[volume.Name] = volume
            }
        }
    }

    return
}

func emptyDirVolumes(emptyDir map[string]types.Volume) (result []v1.Volume) {
    var names []string
    for name := range emptyDir {
        names = append(names, name)
    }
    for _, name := range removeDuplicateAndSort(names) {
        result = append(result, v1.Volume{
            Name: fmt.Sprintf("vol-%s", name),
            VolumeSource: v1.VolumeSource{
                EmptyDir: &v1.EmptyDirVolumeSource{},
            },
        })
    }
    return
}

func hostPathVolumes(hostPath map[string]types.Volume) (result []v1.Volume) {
    var names []string
    for name := range hostPath {
        names = append(names, name)
    }
    for _, name := range removeDuplicateAndSort(names) {
        result = append(result, v1.Volume{
            Name: fmt.Sprintf("vol-%s", name),
            VolumeSource: v1.VolumeSource{
                HostPath: &v1.HostPathVolumeSource{
                    Path: hostPath[name].HostPath,
                    Type: hostPath[name].HostPathType,
                },
            },
        })
    }
    return
}

func pvcVolumes(pvcs map[string]types.Volume) (result []v1.Volume) {
    var names []string
    for name := range pvcs {
        names = append(names, name)
    }
    for _, name := range removeDuplicateAndSort(names) {
        result = append(result, v1.Volume{
            Name: fmt.Sprintf("vol-%s", name),
            VolumeSource: v1.VolumeSource{
                PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
                    ClaimName: name,
                },
            },
        })
    }
    return
}

func NormalizeVolumeTemplates(service types.Service) map[string]types.VolumeTemplate {
    templates := map[string]types.VolumeTemplate{}
    for _, template := range service.Spec.VolumeTemplates {
        if _, ok := templates[template.Name]; ok || template.Name == "" {
            continue
        }
        templates[template.Name] = template
    }
    return templates
}

func diskVolumes(containers []types.NamedContainer, service types.Service) (result []v1.Volume) {
    templates := NormalizeVolumeTemplates(service)
    sorted := sortVolumes(containers, templates)

    result = append(result, emptyDirVolumes(sorted.EmptyDir)...)
    result = append(result, hostPathVolumes(sorted.HostPath)...)
    result = append(result, pvcVolumes(sorted.PVC)...)
    return
}

func volumes(service types.Service) (result []v1.Volume) {
    containers := utils.ToNamedContainers(service)
    result = append(result, secretVolumes(containers)...)
    result = append(result, configVolumes(containers)...)
    result = append(result, diskVolumes(containers, service)...)
    return
}

func removeDuplicateAndSort(array []string) (result []string) {
    set := map[string]struct{}{}
    for _, s := range array {
        set[s] = struct{}{}
    }
    for k := range set {
        result = append(result, k)
    }
    sort.Strings(result)
    return result
}

