package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initCmd(handler Handler) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "init",
		Short: "init",
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

func (cli *cli) initAlleForKube(args []string) error {

	log.Infof("Initialize alle...")
	err := cli.initializer.Init()
	if err != nil {
		return err
	}
	log.Infof("Initialize alle done.")
	return nil
}
