package cmd

import (
	cli "github.com/rancher/wrangler-cli"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/kubeconfig"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	Apply        apply.Apply
	K8sInterface kubernetes.Interface
	RestConfig   *rest.Config
	Kubeconfig   string
	Debug        bool
	RdnsDomain   string
)

func New() *cobra.Command {
	root := cli.Command(&Dolly{}, cobra.Command{

		Short: "Create, manage kubernetes application using dollyfile",
	})
	root.AddCommand(
		NewUpCommand(),
		NewRenderCommand(),
		NewBuildCommand(),
		NewPushCommand(),
		NewPsCommand(),
		NewExecCommand(),
		NewKillCommand(),
		NewRmCommand(),
		NewLogCommand(),
		NewConfigCommand(),
		NewSecretCommand(),
	)
	return root
}

type Dolly struct {
	KubeConfig string `name:"kubeconfig" usage:"Path to the kubeconfig file to use for CLI requests."`
	Debug      bool   `name:"debug" usage:"Enable debug log"`
}

func (a *Dolly) Run(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

func (a *Dolly) PersistentPre(cmd *cobra.Command, args []string) error {
	Kubeconfig = a.KubeConfig
	Debug = a.Debug

	loader := kubeconfig.GetInteractiveClientConfig(a.KubeConfig)
	config, err := loader.ClientConfig()
	if err != nil {
		return err
	}
	k8s := kubernetes.NewForConfigOrDie(config)
	K8sInterface = k8s
	RestConfig = config
	Apply = apply.New(k8s.Discovery(), apply.NewClientFactory(config)).WithRateLimiting(20.0).WithSetID("dolly")
	return nil
}
