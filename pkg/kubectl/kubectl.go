package kubectl

import (
	"fmt"
	"os"
	"os/exec"
)

func Run(namespace, command, kubeconfig string, debug bool, args ...string) error {
	var execArgs []string
	if debug {
		execArgs = append(execArgs, "--v=9")
	}
	if namespace != "" {
		execArgs = append(execArgs, "-n", namespace)
	}
	if command != "" {
		execArgs = append(execArgs, command)
	}
	execArgs = append(execArgs, args...)

	cmd := exec.Command("kubectl", execArgs...)
	if kubeconfig != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", kubeconfig))
	}
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
