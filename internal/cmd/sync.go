package cmd

import (
	"alle/internal/kube"
	"alle/internal/services"
	"github.com/spf13/cobra"
)

func syncCmd() *cobra.Command {

	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "sync",
		RunE: func(cmd *cobra.Command, args []string) error {
			templator := services.NewTemplator()
			configurator, err := services.NewConfiguratorFromFile(templator, environment, filepath)

			deployController, err := kube.NewDeployControllerFromEnv(environment)
			if err != nil {
				return err
			}
			packagesToApply := configurator.GetPackagesByLabels(labels)
			return deployController.ApplyPackages(packagesToApply)
		},
	}
	//cli.rootCmd.PersistentFlags().StringSliceVarP(&labels, "label", "l", []string{}, "specify label to select")
	//cli.rootCmd.PersistentFlags().StringVarP(&environment, "environment", "e", "", "environment")
	//cli.rootCmd.PersistentFlags().StringVarP(&filepath, "filepath", "f", "./allefile.yaml", "filepath to allefile.yaml")
	//cli.rootCmd.AddCommand(syncCmd)
	//cli.rootCmd.AddCommand(&cobra.Command{
	//	Use:   "Template",
	//	Short: "generate Template",
	//	Long:  "generate Template",
	//	Run: func(cmd *cobra.Command, args []string) {
	//		_, err := internal.GetStringTemplatesByLabels(filepath, environment, labels)
	//		if err != nil {
	//			log.Error(err)
	//		}
	//		return
	//	},
	//})
	return syncCmd
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
