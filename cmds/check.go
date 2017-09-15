package cmds

import (
	"github.com/appscode/kloader/controller"
	"github.com/appscode/log"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Validate kloader configuration",
		Run: func(cmd *cobra.Command, args []string) {
			if configMap != "" {
				mounter := controller.NewConfigMapMounter(getRestConfig(), configMap, mountDir, bashFile)
				obj, err := mounter.KubeClient.CoreV1().ConfigMaps(mounter.Source.Namespace).
					Get(mounter.Source.Name, metav1.GetOptions{})
				if err != nil {
					log.Fatalln("Failed to get ConfigMap, Cause", err)
				}
				mounter.Mount(obj)
			} else if secret != "" {
				mounter := controller.NewSecretMounter(getRestConfig(), secret, mountDir, bashFile)
				obj, err := mounter.KubeClient.CoreV1().Secrets(mounter.Source.Namespace).
					Get(mounter.Source.Name, metav1.GetOptions{})
				if err != nil {
					log.Fatalln("Failed to get Secret, Cause", err)
				}
				mounter.Mount(obj)
			}
		},
	}
	addFlags(cmd)
	return cmd
}
