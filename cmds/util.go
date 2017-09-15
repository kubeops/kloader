package cmds

import (
	"github.com/spf13/cobra"
)

var (
	configMap, secret, mountDir, bashFile, kubeMaster, kubeConfig string
)

func addFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&configMap, "config-map", "c", "", "Configmap name that needs to be mount")
	cmd.PersistentFlags().StringVarP(&secret, "secret", "s", "", "Secret name that needs to be mount")
	cmd.PersistentFlags().StringVarP(&mountDir, "mount-location", "m", "", "Volume location where the file will be mounted")
	cmd.PersistentFlags().StringVarP(&bashFile, "boot-cmd", "b", "", "Bash script that will be run on every change of the file")
	cmd.PersistentFlags().StringVar(&kubeMaster, "k8s-master", "", "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	cmd.PersistentFlags().StringVar(&kubeConfig, "k8s-config", "", "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
}
