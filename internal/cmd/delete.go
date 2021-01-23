package cmd

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func deleteCmd(handler Handler) *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "del",
		Run: func(cmd *cobra.Command, args []string) {
			err := handler(args)
			if err != nil {
				log.Error(err)
			}
			return
		},
	}
	return deleteCmd
}

func (cli *cli) deleteEntityHandler(args []string) error {

	packages := cli.configurator.GetPackagesByLabels(cli.alleConfig, labels)
	ctx := context.Background()
	for _, pack := range packages {
		for _, manifest := range pack.Manifests {
			err := cli.kubeClient.DeleteManifest(ctx, manifest)
			if err != nil {
				return fmt.Errorf("cant deploy manifest. Name: %s. OError: %s", manifest.GetFileName(), err.Error())
			}
			log.Debugf("Manifest \"%s\" terminating status: ok\n", manifest.GetFullName())
		}

	}
	return nil
}
