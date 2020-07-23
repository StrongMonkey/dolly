package cmd

import (
	"fmt"
	"strings"

	"github.com/rancher/dolly/pkg/kubectl"
	cli "github.com/rancher/wrangler-cli"
	"github.com/spf13/cobra"
)

func NewExecCommand() *cobra.Command {
	exec := cli.Command(&Exec{}, cobra.Command{
		Short: "Exec into pods",
	})
	return exec
}

type Exec struct {
	Container string `name:"container" usage:"specify container name to exec into" short:"c"`
	Namespace string `name:"namespace" usage:"specify namespace" default:"default"`
	Stdin     bool   `name:"stdin" usage:"open standard input" short:"i"`
	Tty       bool   `name:"tty" usage:"enable tty" short:"t"`
}

func (e *Exec) Run(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("at least two parameters are needed")
	}

	args[0] = strings.TrimPrefix(args[0], "pod/")
	if e.Stdin {
		args = append([]string{"-i"}, args...)
	}
	if e.Tty {
		args = append([]string{"-t"}, args...)
	}

	if e.Container != "" {
		args = append(args, "-c", e.Container, "--")
	}
	return kubectl.Run(e.Namespace, "exec", Kubeconfig, Debug, args...)
}
