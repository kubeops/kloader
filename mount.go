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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	clientset "k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type configMapMounter struct {
	source        *apiv1.ObjectReference
	mountLocation string
	cmdFile       string

	kubeConfig *rest.Config
	kubeClient clientset.Interface
}

func NewConfigMapMounter(kubeConfig *rest.Config, configMap, mountDir, cmd string) *configMapMounter {
	configMapParts := strings.Split(strings.TrimSpace(configMap), ".")
	source := &apiv1.ObjectReference{
		Name: configMapParts[0],
	}

	// If Namespace is not provided with configMap Name try the Pod Namespace
	// or default namespace.
	source.Namespace = os.Getenv("KUBE_NAMESPACE")
	if len(source.Namespace) == 0 {
		source.Namespace = apiv1.NamespaceDefault
	}
	if len(configMapParts) == 2 {
		source.Namespace = configMapParts[1]
	}

	return &configMapMounter{
		source:        source,
		mountLocation: strings.TrimSuffix(mountDir, "/"),
		cmdFile:       cmd,
		kubeConfig:    kubeConfig,
		kubeClient:    clientset.NewForConfigOrDie(kubeConfig),
	}
}

func (c *configMapMounter) Run() {
	c.Mount()
	c.Watch()
}

func (c *configMapMounter) Mount() {
	configMap, err := c.kubeClient.CoreV1().ConfigMaps(c.source.Namespace).Get(c.source.Name, metav1.GetOptions{})
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
		&apiv1.ConfigMap{},
		time.Minute*5,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				c.ReMount()
			},
			UpdateFunc: func(old, new interface{}) {
				if oldMap, oldOK := old.(*apiv1.ConfigMap); oldOK {
					if newMap, newOK := new.(*apiv1.ConfigMap); newOK {
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

func (c *configMapMounter) listFunc(client clientset.Interface) func(metav1.ListOptions) (runtime.Object, error) {
	return func(opts metav1.ListOptions) (runtime.Object, error) {
		return client.CoreV1().ConfigMaps(c.source.Namespace).List(metav1.ListOptions{
			FieldSelector: fields.OneTermEqualSelector("metadata.name", c.source.Name).String(),
		})
	}
}

func (c *configMapMounter) watchFunc(client clientset.Interface) func(options metav1.ListOptions) (watch.Interface, error) {
	return func(options metav1.ListOptions) (watch.Interface, error) {
		return client.CoreV1().ConfigMaps(c.source.Namespace).Watch(metav1.ListOptions{
			FieldSelector: fields.OneTermEqualSelector("metadata.name", c.source.Name).String(),
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
