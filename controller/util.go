package controller

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"

	"github.com/appscode/log"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

var updateAcknowledged, updatePerformed uint64

func updateAcknowledgedCounter() {
	atomic.AddUint64(&updateAcknowledged, 1)
	log.Infoln("Update Acknowledged:", atomic.LoadUint64(&updateAcknowledged))
}

func updatePerformedCounter() {
	atomic.AddUint64(&updatePerformed, 1)
	log.Infoln("Update Performed:", atomic.LoadUint64(&updatePerformed))
}

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
