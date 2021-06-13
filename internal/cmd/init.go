package cmd

import (
	"alle/internal/kube"
	"github.com/spf13/cobra"
)

func initCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "init",
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeInitializer, err := kube.NewKubeInitializer(environment)
			if err != nil {
				return err
			}

			err = kubeInitializer.Init()
			if err != nil {
				return err
			}
			return nil
		},
	}
	return initCmd
}
