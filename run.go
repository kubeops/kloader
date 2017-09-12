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
			if configMap != "" {
				mounter := NewConfigMapMounter(getRestConfig(), configMap, mountDir, bashFile)
				mounter.Run()
			} else if secret != "" {
				mounter := NewSecretMounter(getRestConfig(), secret, mountDir, bashFile)
				mounter.Run()
			}
			hold.Hold()
		},
	}

	return cmd
}

func getRestConfig() *rest.Config {
	if configMap == "" && secret == "" {
		log.Fatal("ConfigMap/Secret is required, but not provided")
	}

	if configMap != "" && secret != "" {
		log.Fatal("Either ConfigMap or Secret is required, but both are provided")
	}

	if mountDir == "" {
		log.Fatal("MountDir is required, but not provided")
	}

	config, err := clientcmd.BuildConfigFromFlags(kubeMaster, kubeConfig)
	if err != nil {
		log.Fatal("Failed to create KubeConfig")
	}
	return config
}
