package portforward

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"k8s.io/client-go/kubernetes/scheme"

	v1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	portforwardtools "k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/kubectl/pkg/cmd/portforward"
)

type Option struct {
	Pod        v1.Pod
	Port       string
	TargetPort string
	Stdout     bool
	ReadyChan  chan struct{}
	StopChan   chan struct{}
}

func PortForward(restConfig *rest.Config, k8s *kubernetes.Clientset, option Option) error {
	if err := setConfigDefaults(restConfig); err != nil {
		return err
	}
	restClient, err := rest.RESTClientFor(restConfig)
	if err != nil {
		return err
	}
	ioStreams := genericclioptions.IOStreams{}
	if option.Stdout {
		ioStreams.Out = os.Stdout
		ioStreams.ErrOut = os.Stderr
	}

	portForwardOpt := portforward.PortForwardOptions{
		Namespace:    option.Pod.Namespace,
		PodName:      option.Pod.Name,
		RESTClient:   restClient,
		Config:       restConfig,
		PodClient:    k8s.CoreV1(),
		Address:      []string{"localhost"},
		Ports:        []string{fmt.Sprintf("%s:%s", option.Port, option.TargetPort)},
		StopChannel:  option.StopChan,
		ReadyChannel: option.ReadyChan,
		PortForwarder: &defaultPortForwarder{
			IOStreams: ioStreams,
		},
	}
	return portForwardOpt.RunPortForward()
}

type defaultPortForwarder struct {
	genericclioptions.IOStreams
}

func (f *defaultPortForwarder) ForwardPorts(method string, url *url.URL, opts portforward.PortForwardOptions) error {
	transport, upgrader, err := spdy.RoundTripperFor(opts.Config)
	if err != nil {
		return err
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, method, url)
	fw, err := portforwardtools.NewOnAddresses(dialer, opts.Address, opts.Ports, opts.StopChannel, opts.ReadyChannel, f.Out, f.ErrOut)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/api"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}
