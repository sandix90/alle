package cmd

import (
	"alle/internal/kube"
	"alle/internal/services"
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func deleteCmd() *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "del",
		RunE: func(cmd *cobra.Command, args []string) error {

			templator := services.NewTemplator()
			configurator, err := services.NewConfiguratorFromFile(templator, environment, filepath)
			if err != nil {
				return err
			}

			kubeClient, err := kube.NewKubeClientFromKubeConfigEnv(environment)
			if err != nil {
				return err
			}

			packages := configurator.GetPackagesByLabels(labels)
			ctx := context.Background()
			for _, pack := range packages {
				for _, manifest := range pack.Manifests {
					err := kubeClient.DeleteManifest(ctx, manifest)
					if err != nil {
						return fmt.Errorf("cant delete manifest. Name: %s. OError: %s", manifest.GetFileName(), err.Error())
					}
					log.Debugf("Manifest \"%s\" terminating status: ok\n", manifest.GetFullName())
				}

			}
			return nil
		},
	}
	return deleteCmd
}
