package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"os"
)

var (
	debug       bool
	filepath    string
	environment string
	labels      []string
)

type Commander interface {
	Execute() error
}
type Handler func(args []string) error

func RootCmd() (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use: "alle",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("hello")
		},
	}
	rootCmd.PersistentFlags().StringSliceVarP(&labels, "label", "l", []string{}, "specify label to select")
	rootCmd.PersistentFlags().StringVarP(&environment, "environment", "e", "", "environment")
	rootCmd.PersistentFlags().StringVarP(&filepath, "filepath", "f", "./allefile.yaml", "filepath to allefile.yaml")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "set debug flag")

	//templator := services.NewTemplator()
	//configurator := services.NewConfigurator(templator)
	//alleConfig := new(models.AlleConfig)
	//
	//cliInst := cli{
	//	rootCmd:      rootCmd,
	//	templator:    templator,
	//	configurator: configurator,
	//	alleConfig:   alleConfig,
	//}
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		level := "info"
		if debug {
			level = "debug"
		}
		if err := setUpLogs(os.Stdout, level); err != nil {
			return err
		}
		return nil
	}

	rootCmd.AddCommand(syncCmd())
	rootCmd.AddCommand(deleteCmd())
	rootCmd.AddCommand(listCmd())
	rootCmd.AddCommand(initCmd())

	return rootCmd, nil
}

func setUpLogs(out io.Writer, level string) error {
	log.SetOutput(out)

	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})

	lvl, err := log.ParseLevel(level)
	if err != nil {
		return err
	}
	log.SetLevel(lvl)

	return nil
}

//func (cli *cli) init() error {
//
//	level := "info"
//	if debug {
//		level = "debug"
//	}
//	if err := setUpLogs(os.Stdout, level); err != nil {
//		return err
//	}
//
//	workDir, err := os.Getwd()
//	log.Debugf("Workdir: %s", workDir)
//	log.Debugf("Using alle file: %s", filepath)
//
//	err = internal.Exists(filepath)
//	if err != nil {
//		log.Errorf("alle file is not found")
//		return err
//	}
//
//	file, err := os.Open(filepath)
//	if err != nil {
//		log.Errorf("Error open file: %s. OErr: %v", filepath, err)
//		return err
//	}
//
//	err = cli.configurator.parseConfig(cli.alleConfig, environment, file)
//	if err != nil {
//		log.Errorf("Error parsing config. OError: %v", err)
//		return err
//	}
//
//	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
//	if err != nil {
//		return err
//	}
//	dynclient, err := dynamic.NewForConfig(config)
//	kubeClient, err := kube.NewKubeClient(dynclient, cli.alleConfig.Environment, config)
//	if err != nil {
//		log.Errorf("Error creating KubeClient. OError: %v", err)
//		return err
//	}
//	cli.kubeClient = kubeClient
//	cli.deployController = kube.NewDeployController(kubeClient)
//	cli.initializer = kube.NewKubeInitializer(dynclient, cli.alleConfig.Environment, config)
//
//	return nil
//}
//
//type cli struct {
//	rootCmd          *cobra.Command
//	templator        services.Templator
//	configurator     services.Configurator
//	alleConfig       *models.AlleConfig
//	kubeClient       kube.IKubeClient
//	deployController kube.DeployController
//	initializer      kube.AlleInitializer
//}
//
//func (cli *cli) Execute() error {
//	return cli.rootCmd.Execute()
//}
