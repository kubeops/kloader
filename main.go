package main

import (
	"github.com/appscode/go/flags"
	"github.com/appscode/go/hold"
	"github.com/appscode/log"
	logs "github.com/appscode/log/golog"
	"github.com/spf13/pflag"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
)

func main() {
	var configMap, mountDir, bashFile, kubeMaster, kubeConfig string
	pflag.StringVarP(&configMap, "config-map", "c", "", "Configmap name that needs to be mount")
	pflag.StringVarP(&mountDir, "mount-location", "m", "", "Volume location where the file will be mounted")
	pflag.StringVarP(&bashFile, "boot-cmd", "b", "", "Bashscript that will be run on every change of the file")
	pflag.StringVar(&kubeMaster, "k8s-master", "", "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	pflag.StringVar(&kubeConfig, "k8s-config", "", "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	flags.InitFlags()
	logs.InitLogs()
	flags.DumpAll()

	if configMap == "" || mountDir == "" {
		log.Fatal("ConfigMap/MountDir is required, but not provided")
	}

	config, err := clientcmd.BuildConfigFromFlags(kubeMaster, kubeConfig)
	if err != nil {
		log.Fatal("Failed to create KubeConfig")
	}

	log.Infoln("Running ConfigMap Mounter for configMap", configMap)
	mounter := NewConfigMapMounter(config, configMap, mountDir, bashFile)
	mounter.Run()

	hold.Hold()
}
