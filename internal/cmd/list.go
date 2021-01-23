package cmd

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func listCmd(handler Handler) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "list",
		Run: func(cmd *cobra.Command, args []string) {
			err := handler(args)
			if err != nil {
				log.Error(err)
			}
			return
		},
	}
	listCmd.PersistentFlags().StringVarP(&environment, "environment", "e", "", "environment")
	return listCmd
}

func (cli *cli) listEntities(args []string) error {

	log.Infof("Getting deployed entities...")
	ctx := context.Background()
	_, err := cli.kubeClient.GetManifestsList(ctx)
	if err != nil {
		return err
	}
	return nil
}
