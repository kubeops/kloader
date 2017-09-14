package main

import (
	"fmt"
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
	"k8s.io/client-go/util/workqueue"
)

type secretMounter struct {
	source        *apiv1.ObjectReference
	mountLocation string
	cmdFile       string

	kubeConfig *rest.Config
	kubeClient clientset.Interface

	queue    workqueue.RateLimitingInterface
	informer cache.SharedIndexInformer
}

func NewSecretMounter(kubeConfig *rest.Config, secret, mountDir, cmd string) *secretMounter {
	secretParts := strings.SplitN(strings.TrimSpace(secret), ".", 2)
	source := &apiv1.ObjectReference{
		Name: secretParts[0],
	}
	if len(secretParts) == 2 {
		source.Namespace = secretParts[1]
	} else {
		source.Namespace = namespace()
	}

	client := clientset.NewForConfigOrDie(kubeConfig)
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
				return client.CoreV1().Secrets(source.Namespace).List(metav1.ListOptions{
					FieldSelector: fields.OneTermEqualSelector("metadata.name", source.Name).String(),
				})
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return client.CoreV1().Secrets(source.Namespace).Watch(metav1.ListOptions{
					FieldSelector: fields.OneTermEqualSelector("metadata.name", source.Name).String(),
				})
			},
		},
		&apiv1.Secret{},
		time.Minute*5,
		cache.Indexers{},
	)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if key, err := cache.MetaNamespaceKeyFunc(obj); err == nil {
				log.Infoln("Queued Add event")
				queue.Add(key)
			} else {
				log.Infoln(err)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			if oldSecret, oldOK := old.(*apiv1.Secret); oldOK {
				if newSecret, newOK := new.(*apiv1.Secret); newOK {
					if !reflect.DeepEqual(oldSecret.Data, newSecret.Data) {
						if key, err := cache.MetaNamespaceKeyFunc(new); err == nil {
							log.Infoln("Queued Update event", key)
							queue.Add(key)
						} else {
							log.Infoln(err)
						}
					}
				}
			}
		},
	})

	return &secretMounter{
		source:        source,
		mountLocation: strings.TrimSuffix(mountDir, "/"),
		cmdFile:       cmd,
		kubeConfig:    kubeConfig,
		kubeClient:    client,
		queue:         queue,
		informer:      informer,
	}
}

func (c *secretMounter) Run() {
	go c.informer.Run(wait.NeverStop)
	wait.Until(c.runWorker, time.Second, wait.NeverStop)
}

func (c *secretMounter) runWorker() {
	for c.processNextItem() {
		// continue looping
	}
}

func (c *secretMounter) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.processItem(key.(string))
	if err == nil {
		c.queue.Forget(key)
	} else if c.queue.NumRequeues(key) < maxRetries {
		log.Infoln("Error processing %s (will retry): %v\n", key, err)
		c.queue.AddRateLimited(key)
	} else {
		log.Infoln("Error processing %s (giving up): %v\n", key, err)
		c.queue.Forget(key)
	}

	return true
}

func (c *secretMounter) processItem(key string) error {
	log.Infoln("Processing change to secret %s\n", key)

	obj, exists, err := c.informer.GetIndexer().GetByKey(key)
	if err != nil {
		return fmt.Errorf("Error fetching object with key %s from store: %v", key, err)
	}

	if !exists {
		log.Infoln("Not exists: secret %s\n", key)
		return nil
	}

	// handle the event
	if obj.(*apiv1.Secret) != nil {
		c.Mount(obj.(*apiv1.Secret))
	}
	if len(c.cmdFile) > 0 {
		runCmd(c.cmdFile)
	}
	return nil
}

func (c *secretMounter) Mount(secret *apiv1.Secret) {
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
		log.Fatalln("Failed to Mount secret, Cause", err)
	}
}
