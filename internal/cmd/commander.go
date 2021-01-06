package cmd

import (
	"alle/internal/kubeclient"
	"alle/internal/models"
	"alle/internal/services"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
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
type CliHandler func(args []string) error

func NewCommander() (Commander, error) {
	rootCmd := &cobra.Command{
		Use: "alle",
	}
	rootCmd.PersistentFlags().StringSliceVarP(&labels, "label", "l", []string{}, "specify label to select")
	rootCmd.PersistentFlags().StringVarP(&environment, "environment", "e", "", "environment")
	rootCmd.PersistentFlags().StringVarP(&filepath, "filepath", "f", "./allefile.yaml", "filepath to allefile.yaml")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "set debug flag")
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

	templator := models.NewTemplator()
	configurator := services.NewConfigurator(templator)
	alleConfig := new(models.AlleConfig)

	cli := cli{
		rootCmd:      rootCmd,
		templator:    templator,
		configurator: configurator,
		alleConfig:   alleConfig,
	}

	rootCmd.AddCommand(syncCmd(cli.syncHandler))
	rootCmd.AddCommand(deleteCmd(cli.deleteEntityHandler))

	return &cli, nil
}
func (cli *cli) init() {

	err := cli.configurator.ParseConfig(cli.alleConfig, environment, filepath)
	if err != nil {
		log.Errorf("Error parsing config. OError: %v", err)
		return
	}

	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	dynclient, err := dynamic.NewForConfig(config)
	kubeClient, err := kubeclient.NewKubeClient(dynclient, cli.alleConfig.Environment, config)
	if err != nil {
		log.Errorf("Error creating KubeClient. OError: %v", err)
		return
	}
	cli.kubeClient = kubeClient

}

type cli struct {
	rootCmd      *cobra.Command
	templator    models.Templator
	configurator services.Configurator
	alleConfig   *models.AlleConfig
	kubeClient   kubeclient.IKubeClient
}

func (cli *cli) Execute() error {
	cobra.OnInitialize(cli.init)
	return cli.rootCmd.Execute()
}

func setUpLogs(out io.Writer, level string) error {
	log.SetOutput(out)

	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	lvl, err := log.ParseLevel(level)
	if err != nil {
		return err
	}
	log.SetLevel(lvl)
	return nil
}
