package cmd

import (
	"alle/internal/kube"
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"time"
)

func listCmd() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeClient, err := kube.NewKubeClientFromKubeConfigEnv(environment)
			if err != nil {
				return err
			}

			log.Infof("Getting deployed entities...")
			ctx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancelFn()
			manifests, err := kubeClient.GetManifestsList(ctx)
			if err != nil {
				return err
			}
			log.Println(manifests)
			return nil
		},
	}
	listCmd.PersistentFlags().StringVarP(&environment, "environment", "e", "", "environment")
	return listCmd
}
