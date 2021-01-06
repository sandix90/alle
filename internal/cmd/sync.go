package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

//var (
//	filepath    string
//	environment string
//	labels      []string
//)

func syncCmd(handler CliHandler) *cobra.Command {
	//cmd := &cobra.Command{
	//	Use:     "schema",
	//	Short:   "sch",
	//	Long:    "Operations with schemas",
	//	Aliases: []string{"sch", "s"},
	//}

	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "sync",
		Run: func(cmd *cobra.Command, args []string) {
			err := handler(args)
			if err != nil {
				log.Error(err)
			}
			return
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

func (cli *cli) syncHandler(args []string) error {
	fmt.Printf("from syncHandler cli command %s\n", filepath)
	//tmpls, err := internal.GetStringTemplatesByLabels(filepath, environment, labels)
	//if err != nil {
	//	return err
	//}
	//kubeClient, err := kubeclient.GetKubeClient()
	//if err != nil {
	//	return err
	//}
	//
	//templator := models.NewTemplator()
	//configurator := services.NewConfigurator(templator)
	//
	//alleConfig := new(models.AlleConfig)
	//err := configurator.ParseConfig(alleConfig, environment, filepath)
	//if err != nil {
	//	return fmt.Errorf("error parse alle config. Origin error: %w", err)
	//}

	//tmpls, err := configurator.GetStringManifestsByLabels(alleConfig, labels)

	//tmpls, err := internal.GetStringTemplatesByLabels(filepath, environment, labels)
	//if err != nil {
	//	return err
	//}
	//log.Debugln(tmpls)

	//exist, err := kubeClient.IsManifestDeployed(alleConfig.Releases[0].Packages[0].Manifests[0])
	//log.Printf("Manifest exist: %v", exist)

	packages := cli.configurator.GetPackagesByLabels(cli.alleConfig, labels)
	for _, pack := range packages {
		for _, manifest := range pack.Manifests {
			err := cli.kubeClient.ApplyManifest(manifest)
			if err != nil {
				return fmt.Errorf("cant deploy manifest. Name: %s", manifest.GetName())
			}
			//log.Debugf("Manifest \"%s\" deployed status: ok\n", manifest.GetName())
		}
	}

	//kubeclient.MonitorPods()
	//kubeclient.GetDepTest(tmpls)
	//time.Sleep(time.Second*120)
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
