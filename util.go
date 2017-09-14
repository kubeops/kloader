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
	"io/ioutil"
)

func namespace() string {
	if ns := os.Getenv("KUBE_NAMESPACE"); ns != "" {
		return ns
	}
	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns
		}
	}
	return apiv1.NamespaceDefault
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
