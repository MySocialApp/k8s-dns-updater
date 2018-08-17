package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/MySocialApp/k8s-dns-updater/core"
	"github.com/sirupsen/logrus"
)

var cfgFile string
var Verbose string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "k8s-dns-updater",
	Short: "Update your public round robin DNS according your available kubernetes nodes",
	Long: `Kubernetes DNS updater is a tool watching Kubernetes nodes status changes and update the Round Robin DNS accordingly.
This is useful when running an on premise cluster with a simple DNS load balancing.
This to avoid manual intervention when a node fails down or is going into maintenance.

More info: https://github.com/MySocialApp/k8s-dns-updater`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		// Set log level
		setLogLevel()
		// Launch controller
		core.Main()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of k8s-dns-updater",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("v0.2")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.test.yaml)")
	rootCmd.PersistentFlags().StringVar(&Verbose, "log-level", "info", "debug output")

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func setLogLevel() {
	switch Verbose {
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "fatal":
		logrus.SetLevel(logrus.FatalLevel)
	case "panic":
		logrus.SetLevel(logrus.PanicLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}