package tables

import (
	"fmt"
	"strings"

	"github.com/rancher/dolly/pkg/table"
	v1 "k8s.io/api/core/v1"
)

func NewPods(namespace, format string, quiet bool) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{id .Obj}}"},
		{"CREATED", "{{.Obj.CreationTimestamp | ago}}"},
		{"READY", "{{.Obj | podReady}}"},
		{"NODE", "{{.Obj.Spec.NodeName}}"},
		{"IP", "{{.Obj.Status.PodIP}}"},
		{"STATE", "{{.Obj.Status.Phase}}"},
		{"DETAIL", "{{.Obj | podDetail}}"},
	}, namespace, quiet, format)

	writer.AddFormatFunc("podReady", podReady)
	writer.AddFormatFunc("podDetail", podDetail)
	return &tableWriter{
		writer: writer,
	}
}

func podDetail(obj interface{}) string {
	pod, _ := obj.(*v1.Pod)
	return detail(pod)
}

func detail(pod *v1.Pod) string {
	output := strings.Builder{}
	if pod == nil {
		return ""
	}
	for _, con := range append(pod.Status.ContainerStatuses, pod.Status.InitContainerStatuses...) {
		if con.State.Waiting != nil && con.State.Waiting.Reason != "" {
			output.WriteString("; ")
			reason := con.State.Waiting.Reason
			if con.State.Waiting.Message != "" {
				reason = reason + "/" + con.State.Waiting.Message
			}
			output.WriteString(fmt.Sprintf("%s(%s)", con.Name, reason))
		}

		if con.State.Terminated != nil && con.State.Terminated.ExitCode != 0 {
			output.WriteString(";")
			if con.State.Terminated.Message == "" {
				con.State.Terminated.Message = "exit code not zero"
			}
			reason := con.State.Terminated.Reason
			if con.State.Terminated.Message != "" {
				reason = reason + "/" + con.State.Terminated.Message
			}
			output.WriteString(fmt.Sprintf("%s(%s), exit code: %v", con.Name, reason, con.State.Terminated.ExitCode))
		}
	}
	return strings.Trim(output.String(), "; ")
}

func podReady(obj interface{}) (string, error) {
	podData, _ := obj.(*v1.Pod)
	ready := 0
	total := 0
	for _, con := range podData.Status.ContainerStatuses {
		if con.Ready {
			ready++
		}
		total++
	}
	return fmt.Sprintf("%v/%v", ready, total), nil
}
