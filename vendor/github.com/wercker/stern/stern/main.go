//   Copyright 2016 Wercker Holding BV
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package stern

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/pkg/errors"
	"github.com/wercker/stern/kubernetes"
)

// Run starts the main run loop
func Run(ctx context.Context, config *Config) error {
	clientConfig := kubernetes.NewClientConfig(config.KubeConfig, config.ContextName)
	clientset, err := kubernetes.NewClientSet(clientConfig)
	if err != nil {
		return err
	}

	var namespace string
	// A specific namespace is ignored if all-namespaces is provided
	if config.AllNamespaces {
		namespace = ""
	} else {
		namespace = config.Namespace
		if namespace == "" {
			namespace, _, err = clientConfig.Namespace()
			if err != nil {
				return errors.Wrap(err, "unable to get default namespace")
			}
		}
	}

	added, removed, err := Watch(ctx,
		clientset.CoreV1().Pods(namespace),
		config.PodQuery,
		config.ContainerQuery,
		config.ExcludeContainerQuery,
		config.InitContainers,
		config.ContainerState,
		config.LabelSelector)
	if err != nil {
		return errors.Wrap(err, "failed to set up watch")
	}

	tails := make(map[string]*Tail)
	tailsMutex := sync.RWMutex{}
	logC := make(chan string, 1024)

	go func() {
		for {
			select {
			case str := <-logC:
				fmt.Fprintf(os.Stdout, str)
			case <-ctx.Done():
				break
			}
		}
	}()

	go func() {
		for p := range added {
			id := p.GetID()
			tailsMutex.RLock()
			existing := tails[id]
			tailsMutex.RUnlock()
			if existing != nil {
				if existing.Active == true {
					continue
				} else { // cleanup failed tail to restart
					tailsMutex.Lock()
					tails[id].Close()
					delete(tails, id)
					tailsMutex.Unlock()
				}
			}
			tail := NewTail(p.Namespace, p.Pod, p.Container, config.Template, &TailOptions{
				Timestamps:   config.Timestamps,
				SinceSeconds: int64(config.Since.Seconds()),
				Exclude:      config.Exclude,
				Include:      config.Include,
				Namespace:    config.AllNamespaces,
				TailLines:    config.TailLines,
			})
			tailsMutex.Lock()
			tails[id] = tail
			tailsMutex.Unlock()
			tail.Start(ctx, clientset.CoreV1().Pods(p.Namespace), logC)
		}
	}()

	go func() {
		for p := range removed {
			id := p.GetID()
			tailsMutex.RLock()
			existing := tails[id]
			tailsMutex.RUnlock()
			if existing == nil {
				continue
			}
			tailsMutex.Lock()
			tails[id].Close()
			delete(tails, id)
			tailsMutex.Unlock()
		}
	}()

	<-ctx.Done()

	return nil
}
