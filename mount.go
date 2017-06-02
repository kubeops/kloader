package main

import (
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"time"

	"github.com/appscode/kloader/volume"
	"github.com/appscode/log"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/cache"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/wait"
	"k8s.io/kubernetes/pkg/watch"
)

type configMapMounter struct {
	source        *api.ObjectReference
	mountLocation string
	cmdFile       string

	kubeConfig *restclient.Config
	kubeClient internalclientset.Interface
}

func NewConfigMapMounter(kubeConfig *restclient.Config, configMap, mountDir, cmd string) *configMapMounter {
	configMapParts := strings.Split(strings.TrimSpace(configMap), ".")
	source := &api.ObjectReference{
		Name: configMapParts[0],
	}

	// If Namespace is not provided with configMap Name try the Pod Namespace
	// or default namespace.
	source.Namespace = os.Getenv("KUBE_NAMESPACE")
	if len(source.Namespace) == 0 {
		source.Namespace = api.NamespaceDefault
	}
	if len(configMapParts) == 2 {
		source.Namespace = configMapParts[1]
	}

	return &configMapMounter{
		source:        source,
		mountLocation: strings.TrimSuffix(mountDir, "/"),
		cmdFile:       cmd,
		kubeConfig:    kubeConfig,
		kubeClient:    internalclientset.NewForConfigOrDie(kubeConfig),
	}
}

func (c *configMapMounter) Run() {
	c.Mount()
	c.Watch()
}

func (c *configMapMounter) Mount() {
	configMap, err := c.kubeClient.Core().ConfigMaps(c.source.Namespace).Get(c.source.Name)
	if err != nil {
		log.Fatalln("Failed to get ConfigMap, Cause", err)
	}

	payload := make(map[string]volume.FileProjection)
	for k, v := range configMap.Data {
		payload[k] = volume.FileProjection{Mode: 0777, Data: []byte(v)}
	}

	writer, err := volume.NewAtomicWriter(c.mountLocation)
	if err != nil {
		log.Fatalln("Failed to Create atomic writer, Cause", err)
	}
	err = writer.Write(payload)
	if err != nil {
		log.Fatalln("Failed to Mount ConfigMap, Cause", err)
	}
}

func (c *configMapMounter) ReMount() {
	c.Mount()
	if len(c.cmdFile) > 0 {
		runCmd(c.cmdFile)
	}
}

func (c *configMapMounter) Watch() {
	lw := &cache.ListWatch{
		ListFunc:  c.listFunc(c.kubeClient),
		WatchFunc: c.watchFunc(c.kubeClient),
	}

	_, controller := cache.NewInformer(lw,
		&api.ConfigMap{},
		time.Minute*5,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				c.ReMount()
			},
			UpdateFunc: func(old, new interface{}) {
				if oldMap, oldOK := old.(*api.ConfigMap); oldOK {
					if newMap, newOK := new.(*api.ConfigMap); newOK {
						if !reflect.DeepEqual(oldMap.Data, newMap.Data) {
							c.ReMount()
						}
					}
				}
			},
		},
	)
	go controller.Run(wait.NeverStop)
}

func (c *configMapMounter) listFunc(client internalclientset.Interface) func(api.ListOptions) (runtime.Object, error) {
	return func(opts api.ListOptions) (runtime.Object, error) {
		return client.Core().ConfigMaps(c.source.Namespace).List(api.ListOptions{
			FieldSelector: fields.OneTermEqualSelector("metadata.name", c.source.Name),
		})
	}
}

func (c *configMapMounter) watchFunc(client internalclientset.Interface) func(options api.ListOptions) (watch.Interface, error) {
	return func(options api.ListOptions) (watch.Interface, error) {
		return client.Core().ConfigMaps(c.source.Namespace).Watch(api.ListOptions{
			FieldSelector: fields.OneTermEqualSelector("metadata.name", c.source.Name),
		})
	}
}

func runCmd(path string) error {
	log.Infoln("calling boot file to execute")
	output, err := exec.Command("sh", "-c", path).CombinedOutput()
	msg := fmt.Sprintf("%v", string(output))
	log.Infoln("Output:\n", msg)
	if err != nil {
		log.Errorln("failed to run cmd")
		return fmt.Errorf("error restarting %v: %v", msg, err)
	}
	log.Infoln("boot file executed")
	return nil
}
