package main

import (
	"flag"
	golog "log"
	"os"

	"github.com/appscode/go/version"
	"github.com/appscode/log"
	logs "github.com/appscode/log/golog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	configMap, mountDir, bashFile, kubeMaster, kubeConfig string
)

func main() {
	defer logs.FlushLogs()

	rootCmd := &cobra.Command{
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
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	// ref: https://github.com/kubernetes/kubernetes/issues/17162#issuecomment-225596212
	flag.CommandLine.Parse([]string{})
	logs.InitLogs()

	rootCmd.AddCommand(newCheckCmd())
	rootCmd.AddCommand(newRunCmd())
	rootCmd.AddCommand(version.NewCmdVersion())

	rootCmd.PersistentFlags().StringVarP(&configMap, "config-map", "c", "", "Configmap name that needs to be mount")
	rootCmd.PersistentFlags().StringVarP(&mountDir, "mount-location", "m", "", "Volume location where the file will be mounted")
	rootCmd.PersistentFlags().StringVarP(&bashFile, "boot-root", "b", "", "Bashscript that will be run on every change of the file")
	rootCmd.PersistentFlags().StringVar(&kubeMaster, "k8s-master", "", "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	rootCmd.PersistentFlags().StringVar(&kubeConfig, "k8s-config", "", "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
	os.Exit(0)
}
