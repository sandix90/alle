package cmd

import (
	"alle/internal"
	"alle/internal/kubeclient"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"time"
)

var (
	filepath    string
	environment string
	labels      []string
)

func syncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "schema",
		Short:   "sch",
		Long:    "Operations with schemas",
		Aliases: []string{"sch", "s"},
	}

	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "sync",
		Run: func(cmd *cobra.Command, args []string) {
			err := sync(args)
			if err != nil {
				log.Error(err)
			}
			return
		},
	}
	cmd.PersistentFlags().StringSliceVarP(&labels, "label", "l", []string{}, "specify label to select")
	cmd.PersistentFlags().StringVarP(&environment, "environment", "e", "", "environment")
	cmd.PersistentFlags().StringVarP(&filepath, "filepath", "f", "./allefile.yaml", "filepath to allefile.yaml")

	cmd.AddCommand(syncCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "Template",
		Short: "generate Template",
		Long:  "generate Template",
		Run: func(cmd *cobra.Command, args []string) {
			_, err := internal.GetStringTemplatesByLabels(filepath, environment, labels)
			if err != nil {
				log.Error(err)
			}
			return
		},
	})
	return cmd
}

func sync(args []string) error {
	fmt.Printf("from sync cli command %s\n", filepath)
	//tmpls, err := internal.GetStringTemplatesByLabels(filepath, environment, labels)
	//if err != nil {
	//	return err
	//}
	client, err := kubeclient.GetKubeClient()
	if err != nil {
		return err
	}

	tmpls, err := internal.GetStringTemplatesByLabels(filepath, environment, labels)
	if err != nil {
		return err
	}

	kubeclient.CreateDeployment(client, environment, tmpls)
	kubeclient.MonitorPods()
	//kubeclient.GetDepTest(tmpls)
	time.Sleep(time.Second*120)
	return nil
}

//func Template() ([]string, error) {
//	workDir, err := os.Getwd()
//	log.Debugf("Workdir: %s", workDir)
//	log.Debugf("Using alle file: %s", filepath)
//	err = internal.Exists(filepath)
//	if err != nil {
//		return nil, err
//	}
//	aleConfig := &internal.AlleConfig{}
//	aleConfig.Environment = environment
//
//	err = internal.UnmarshalAlleConfig(aleConfig, filepath)
//	if err != nil {
//		log.Error("Bad alle config")
//		return nil, err
//	}
//	var tmpls []string
//	for _, release := range aleConfig.Releases {
//		for _, pack := range release.Packages {
//
//			labelFound := internal.FindByLabel(pack, labels)
//			if labelFound {
//				tmpls, err = internal.GetPackageStringManifests(pack)
//				if err != nil {
//					return nil, err
//				}
//				fmt.Println(tmpls)
//			}
//		}
//	}
//	return tmpls, nil
//}
