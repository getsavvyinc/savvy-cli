package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "savvy",
	Short: "Create, share and discover runbooks from the command line",
	Long:  `Create, share and discover runbooks from the command line`,
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		logLevel := slog.LevelInfo
		if debugFlag {
			logLevel = slog.LevelDebug
		}
		textHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     logLevel,
		})
		logger := slog.New(textHandler)
		slog.SetDefault(logger)
		cmd.SetContext(ctxWithLogger(cmd.Context(), logger))
	},
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

var debugFlag bool

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.savvy.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&debugFlag, "debug", "d", false, "Enable debug mode")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
