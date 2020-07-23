package riofile

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rancher/dolly/pkg/types"
	"github.com/rancher/dolly/pkg/types/convert/deployment"
	"github.com/rancher/dolly/pkg/types/convert/rbac"
	"github.com/rancher/dolly/pkg/types/convert/service"
	"github.com/rancher/dolly/pkg/types/convert/volume"
	"github.com/rancher/dolly/pkg/types/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Riofile struct {
	Services   map[string]types.Service `json:"services,omitempty"`
	Configs    map[string]v1.ConfigMap  `json:"configs,omitempty"`
	Kubernetes []runtime.Object         `json:"kubernetes,omitempty"`
	Manifest   string                   `json:"manifest,omitempty"`
}

func (r *Riofile) Objects() []runtime.Object {
	var result []runtime.Object

	for _, cm := range r.Configs {
		result = append(result, &cm)
	}

	for _, svc := range r.Services {
		result = append(result, deployment.Convert(svc))
		result = append(result, rbac.Convert(svc)...)
		result = append(result, service.Convert(svc))
		result = append(result, volume.Convert(svc)...)
	}

	result = append(result, r.Kubernetes...)
	return result
}

func (r *Riofile) Build(push bool) error {
	for i, service := range r.Services {
		containers := utils.ToNamedContainers(service)
		for j, container := range containers {
			if container.Build == nil {
				continue
			}
			image := container.Image
			if image == "" {
				wd, err := os.Getwd()
				if err != nil {
					return err
				}
				image = filepath.Base(wd)
			}
			if j == 0 {
				service.Spec.Image = image
				r.Services[i] = service
			} else {
				service.Spec.Sidecars[j+1].Image = image
				r.Services[i] = service
			}
			// todo: should we run build in parallel
			if err := runBuild(image, *container.Build); err != nil {
				return err
			}

			if push {
				if err := runPush(container.Image); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func runBuild(image string, build types.ImageBuild) error {
	args := []string{"build"}
	for _, arg := range build.Args {
		args = append(args, "--build-arg", arg)
	}

	for _, label := range build.Labels {
		args = append(args, "--label", label)
	}

	if build.Dockerfile != "" {
		args = append(args, "--file", build.Dockerfile)
	}

	for _, cache := range build.CacheFrom {
		args = append(args, "--cache-from", cache)
	}

	if build.Network != "" {
		args = append(args, "--network", build.Network)
	}

	if build.ShmSize != "" {
		args = append(args, "--shm-size", build.ShmSize)
	}

	if build.Target != "" {
		args = append(args, "--target", build.Target)
	}

	args = append(args, "-t", image)

	context := "./"
	if build.Context != "" {
		context = build.Context
	}

	args = append(args, context)

	cmd := exec.Command("docker", args...)
	output := &strings.Builder{}
	cmd.Stdout = os.Stdout
	cmd.Stderr = output
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running docker: %v", output.String())
	}
	return nil
}

func runPush(image string) error {
	args := []string{"push", image}
	cmd := exec.Command("docker", args...)
	output := &strings.Builder{}
	cmd.Stdout = os.Stdout
	cmd.Stderr = output
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running docker: %v", output.String())
	}
	return nil
}
