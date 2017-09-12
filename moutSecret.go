package main

import (
	"os"
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

type secretMounter struct {
	source        *apiv1.ObjectReference
	mountLocation string
	cmdFile       string

	kubeConfig *rest.Config
	kubeClient clientset.Interface
}

func NewSecretMounter(kubeConfig *rest.Config, secret, mountDir, cmd string) *secretMounter {
	secretParts := strings.Split(strings.TrimSpace(secret), ".")
	source := &apiv1.ObjectReference{
		Name: secretParts[0],
	}

	// If Namespace is not provided with secret Name try the Pod Namespace
	// or default namespace.
	source.Namespace = os.Getenv("KUBE_NAMESPACE")
	if len(source.Namespace) == 0 {
		source.Namespace = apiv1.NamespaceDefault
	}
	if len(secretParts) == 2 {
		source.Namespace = secretParts[1]
	}

	return &secretMounter{
		source:        source,
		mountLocation: strings.TrimSuffix(mountDir, "/"),
		cmdFile:       cmd,
		kubeConfig:    kubeConfig,
		kubeClient:    clientset.NewForConfigOrDie(kubeConfig),
	}
}

func (c *secretMounter) Run() {
	c.Mount()
	c.Watch()
}

func (c *secretMounter) Mount() {
	secret, err := c.kubeClient.CoreV1().Secrets(c.source.Namespace).Get(c.source.Name, metav1.GetOptions{})
	if err != nil {
		log.Fatalln("Failed to get Secret, Cause", err)
	}

	payload := make(map[string]volume.FileProjection)
	for k, v := range secret.Data {
		payload[k] = volume.FileProjection{Mode: 0777, Data: []byte(v)}
	}

	writer, err := volume.NewAtomicWriter(c.mountLocation)
	if err != nil {
		log.Fatalln("Failed to Create atomic writer, Cause", err)
	}
	err = writer.Write(payload)
	if err != nil {
		log.Fatalln("Failed to Mount Secret, Cause", err)
	}
}

func (c *secretMounter) ReMount() {
	c.Mount()
	if len(c.cmdFile) > 0 {
		runCmd(c.cmdFile)
	}
}

func (c *secretMounter) Watch() {
	lw := &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return c.kubeClient.CoreV1().Secrets(c.source.Namespace).List(metav1.ListOptions{
				FieldSelector: fields.OneTermEqualSelector("metadata.name", c.source.Name).String(),
			})
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return c.kubeClient.CoreV1().Secrets(c.source.Namespace).Watch(metav1.ListOptions{
				FieldSelector: fields.OneTermEqualSelector("metadata.name", c.source.Name).String(),
			})
		},
	}

	_, controller := cache.NewInformer(lw,
		&apiv1.Secret{},
		time.Minute*5,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				c.ReMount()
			},
			UpdateFunc: func(old, new interface{}) {
				if oldSecret, oldOK := old.(*apiv1.Secret); oldOK {
					if newSecret, newOK := new.(*apiv1.Secret); newOK {
						if !reflect.DeepEqual(oldSecret.Data, newSecret.Data) {
							c.ReMount()
						}
					}
				}
			},
		},
	)
	go controller.Run(wait.NeverStop)
}
