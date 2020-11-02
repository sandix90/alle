package cmd

import (
	"alle/internal"
	"alle/internal/kubeclient"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func deleteCmd() *cobra.Command{
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "del",
		Run: func(cmd *cobra.Command, args []string) {
			err := deleteEntity(args)
			if err != nil {
				log.Error(err)
			}
			return
		},
	}
	deleteCmd.PersistentFlags().StringSliceVarP(&labels, "label", "l", []string{}, "specify label to select")
	deleteCmd.PersistentFlags().StringVarP(&environment, "environment", "e", "", "environment")
	deleteCmd.PersistentFlags().StringVarP(&filepath, "filepath", "f", "./allefile.yaml", "filepath to allefile.yaml")
	return deleteCmd
}

func deleteEntity(args []string) error{
	client, err := kubeclient.GetKubeClient()
	if err != nil {
		return err
	}

	tmpls, err := internal.GetStringTemplatesByLabels(filepath, environment, labels)
	if err != nil {
		return err
	}

	if len(tmpls) == 0 {
		return errors.New("No templates with given labels were found")
	}

	err = kubeclient.DeleteDeployment(client, environment, tmpls)
	if err != nil {
		return err
	}
	return nil
}
