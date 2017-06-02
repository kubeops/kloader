package main

import (
	"github.com/appscode/go/hold"
	"github.com/appscode/log"
	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
)

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "run kloader",
		Run: func(cmd *cobra.Command, args []string) {
			mounter := NewConfigMapMounter(getRestConfig(), configMap, mountDir, bashFile)
			mounter.Run()
			hold.Hold()
		},
	}

	return cmd
}

func getRestConfig() *restclient.Config {
	if configMap == "" || mountDir == "" {
		log.Fatal("ConfigMap/MountDir is required, but not provided")
	}

	config, err := clientcmd.BuildConfigFromFlags(kubeMaster, kubeConfig)
	if err != nil {
		log.Fatal("Failed to create KubeConfig")
	}
	return config
}
