package tables

import (
	"fmt"
	"strconv"

	"github.com/rancher/dolly/pkg/table"
	appsv1 "k8s.io/api/apps/v1"
)

func NewService(namespace, format string, quiet bool) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{.Obj | id}}"},
		{"IMAGE", "{{.Obj | image}}"},
		{"SCALE", "{{.Obj | scale}}"},
		{"CREATED", "{{.Obj.CreationTimestamp | ago}}"},
	}, namespace, quiet, format)

	writer.AddFormatFunc("image", FormatImage)
	writer.AddFormatFunc("scale", formatRevisionScale)

	return &tableWriter{
		writer: writer,
	}
}

func formatRevisionScale(data interface{}) (string, error) {
	switch v := data.(type) {
	case *appsv1.DaemonSet:
		scale := int(v.Status.DesiredNumberScheduled)
		return FormatScale(&scale,
			int(v.Status.NumberAvailable),
			int(v.Status.NumberUnavailable))
	case *appsv1.Deployment:
		var scale *int
		if v.Spec.Replicas != nil {
			iScale := int(*v.Spec.Replicas)
			scale = &iScale
		}
		return FormatScale(scale,
			int(v.Status.AvailableReplicas),
			int(v.Status.UnavailableReplicas))
	}
	return "", nil
}

func FormatScale(scale *int, available, unavailable int) (string, error) {
	scaleNum := 1
	if scale != nil {
		scaleNum = *scale
	}

	scaleStr := strconv.Itoa(scaleNum)

	if scaleNum == -1 {
		return strconv.Itoa(available), nil
	}

	if unavailable == 0 {
		return scaleStr, nil
	}

	var prefix string
	percentage := ""
	ready := available
	if scaleNum > 0 {
		percentage = fmt.Sprintf(" %d%%", (ready*100)/scaleNum)
	}

	prefix = fmt.Sprintf("%d/", ready)

	return fmt.Sprintf("%s%d%s", prefix, scaleNum, percentage), nil
}

func getDaemonSetImage(d *appsv1.DaemonSet) string {
	if len(d.Spec.Template.Spec.Containers) > 0 {
		return d.Spec.Template.Spec.Containers[0].Image
	}
	return ""
}

func getDeploymentImage(d *appsv1.Deployment) string {
	if len(d.Spec.Template.Spec.Containers) > 0 {
		return d.Spec.Template.Spec.Containers[0].Image
	}
	return ""
}

func FormatImage(data interface{}) string {
	image := ""
	switch v := data.(type) {
	case *appsv1.Deployment:
		image = getDeploymentImage(v)
	case *appsv1.DaemonSet:
		image = getDaemonSetImage(v)
	}

	return image
}
