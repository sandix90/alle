package cmd

import (
	"alle/internal/kubeclient"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

func listCmd() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "list",
		Run: func(cmd *cobra.Command, args []string) {
			err := listEntities(args)
			if err != nil {
				log.Error(err)
			}
			return
		},
	}
	listCmd.PersistentFlags().StringVarP(&environment, "environment", "e", "", "environment")
	return listCmd
}

func listEntities(args []string) error {

	client, err := kubeclient.GetKubeClient()
	if err != nil {
		return err
	}

	log.Infof("Getting deployed entities...")
	_, err = kubeclient.ListDeployments(client, environment, os.Stdout)
	if err != nil {
		return err
	}
	return nil
}
