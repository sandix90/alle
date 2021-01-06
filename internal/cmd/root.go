package cmd

//var (
//	rootCmd = &cobra.Command{
//		Use: "alle",
//	}
//	//debug bool
//)

//func init() {
//	//rootCmd.AddCommand(syncCmd())
//	//rootCmd.AddCommand(deleteCmd())
//	rootCmd.AddCommand(listCmd())
//
//	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
//		level := "info"
//		if debug{
//			level = "debug"
//		}
//		if err := setUpLogs(os.Stdout, level); err != nil{
//			return err
//		}
//		return nil
//	}
//
//	rootCmd.PersistentFlags().BoolVarP(&debug,"debug", "d", false, "set debug flag")
//	rootCmd.SetVersionTemplate("0.0.1")
//}

//func setUpLogs(out io.Writer, level string) error{
//	log.SetOutput(out)
//
//	log.SetFormatter(&log.TextFormatter{
//		TimestampFormat: "2006-01-02 15:04:05",
//	})
//
//	lvl, err := log.ParseLevel(level)
//	if err != nil {
//		return err
//	}
//	log.SetLevel(lvl)
//	return nil
//}

//func Execute() {
//
//	if err := rootCmd.Execute(); err != nil {
//		log.Debug(err)
//		os.Exit(1)
//	}
//}
