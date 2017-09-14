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
	"k8s.io/client-go/util/workqueue"
)

const maxRetries = 5

type configMapMounter struct {
	source        *apiv1.ObjectReference
	mountLocation string
	cmdFile       string

	kubeConfig *rest.Config
	kubeClient clientset.Interface

	queue    workqueue.RateLimitingInterface
	informer cache.SharedIndexInformer
}

func NewConfigMapMounter(kubeConfig *rest.Config, configMap, mountDir, cmd string) *configMapMounter {
	configMapParts := strings.SplitN(strings.TrimSpace(configMap), ".", 2)
	source := &apiv1.ObjectReference{
		Name: configMapParts[0],
	}
	if len(configMapParts) == 2 {
		source.Namespace = configMapParts[1]
	} else {
		source.Namespace = namespace()
	}

	client := clientset.NewForConfigOrDie(kubeConfig)
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
				return client.CoreV1().ConfigMaps(source.Namespace).List(metav1.ListOptions{
					FieldSelector: fields.OneTermEqualSelector("metadata.name", source.Name).String(),
				})
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return client.CoreV1().ConfigMaps(source.Namespace).Watch(metav1.ListOptions{
					FieldSelector: fields.OneTermEqualSelector("metadata.name", source.Name).String(),
				})
			},
		},
		&apiv1.ConfigMap{},
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
			if oldMap, oldOK := old.(*apiv1.ConfigMap); oldOK {
				if newMap, newOK := new.(*apiv1.ConfigMap); newOK {
					if !reflect.DeepEqual(oldMap.Data, newMap.Data) {
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

	return &configMapMounter{
		source:        source,
		mountLocation: strings.TrimSuffix(mountDir, "/"),
		cmdFile:       cmd,
		kubeConfig:    kubeConfig,
		kubeClient:    client,
		queue:         queue,
		informer:      informer,
	}
}

func (c *configMapMounter) Run() {
	go c.informer.Run(wait.NeverStop)
	wait.Until(c.runWorker, time.Second, wait.NeverStop)
}

func (c *configMapMounter) runWorker() {
	for c.processNextItem() {
		// continue looping
	}
}

func (c *configMapMounter) processNextItem() bool {
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

func (c *configMapMounter) processItem(key string) error {
	log.Infoln("Processing change to ConfigMap %s\n", key)

	obj, exists, err := c.informer.GetIndexer().GetByKey(key)
	if err != nil {
		return fmt.Errorf("error fetching object with key %s from store: %v", key, err)
	}

	if !exists {
		log.Infoln("Not exists: ConfigMap %s\n", key)
		return nil
	}

	// handle the event
	c.Mount(obj.(*apiv1.ConfigMap))
	if len(c.cmdFile) > 0 {
		runCmd(c.cmdFile)
	}
	return nil
}

func (c *configMapMounter) Mount(configMap *apiv1.ConfigMap) {
	var err error
	if configMap == nil { // for initial call before caching
		configMap, err = c.kubeClient.CoreV1().ConfigMaps(c.source.Namespace).Get(c.source.Name, metav1.GetOptions{})
		if err != nil {
			log.Fatalln("Failed to get configMap, Cause", err)
			return
		}
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
