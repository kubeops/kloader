package main

import (
	"github.com/spf13/cobra"
)

func newCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Validate kloader configuration",
		Run: func(cmd *cobra.Command, args []string) {
			mounter := NewConfigMapMounter(getRestConfig(), configMap, mountDir, bashFile)
			mounter.Mount()
		},
	}
	return cmd
}
