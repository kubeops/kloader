package main

import (
	"flag"
	golog "log"

	"github.com/appscode/go/hold"
	"github.com/appscode/go/version"
	"github.com/appscode/log"
	logs "github.com/appscode/log/golog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
)

var (
	configMap, mountDir, bashFile, kubeMaster, kubeConfig string
)

func newKloaderCmd() *cobra.Command {
	root := &cobra.Command{
		Use: "kloader",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmd.Flags().VisitAll(func(flag *pflag.Flag) {
				golog.Printf("FLAG: --%s=%q", flag.Name, flag.Value)
			})
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	logs.InitLogs()

	root.AddCommand(&cobra.Command{
		Use:   "check",
		Short: "check validation of required configmap",
		Run: func(cmd *cobra.Command, args []string) {
			check()
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "run",
		Short: "run kloader",
		Run: func(cmd *cobra.Command, args []string) {
			run()
		},
	})
	root.AddCommand(version.NewCmdVersion())

	root.PersistentFlags().StringVarP(&configMap, "config-map", "c", "", "Configmap name that needs to be mount")
	root.PersistentFlags().StringVarP(&mountDir, "mount-location", "m", "", "Volume location where the file will be mounted")
	root.PersistentFlags().StringVarP(&bashFile, "boot-root", "b", "", "Bashscript that will be run on every change of the file")
	root.PersistentFlags().StringVar(&kubeMaster, "k8s-master", "", "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	root.PersistentFlags().StringVar(&kubeConfig, "k8s-config", "", "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	root.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	return root
}

func check() {
	mounter := NewConfigMapMounter(getRestConfig(), configMap, mountDir, bashFile)
	mounter.Mount()
}

func run() {
	mounter := NewConfigMapMounter(getRestConfig(), configMap, mountDir, bashFile)
	mounter.Run()
	hold.Hold()
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
