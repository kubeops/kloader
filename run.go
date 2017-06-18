package main

import (
	"github.com/appscode/go/hold"
	"github.com/appscode/log"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run and hold kloader",
		Run: func(cmd *cobra.Command, args []string) {
			mounter := NewConfigMapMounter(getRestConfig(), configMap, mountDir, bashFile)
			mounter.Run()
			hold.Hold()
		},
	}

	return cmd
}

func getRestConfig() *rest.Config {
	if configMap == "" || mountDir == "" {
		log.Fatal("ConfigMap/MountDir is required, but not provided")
	}

	config, err := clientcmd.BuildConfigFromFlags(kubeMaster, kubeConfig)
	if err != nil {
		log.Fatal("Failed to create KubeConfig")
	}
	return config
}
