/*
Package cmd implement cobra command logic
*/
package cmd

import (
	"barbecue/internal/app"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "barbecue",
	Short: "Voyager camera control tool",
	Long: `Barbecue uses the Starkeeper Voyager Astronomy API to control
that the warmup of your camera is done before switching the power off.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default to report command
		reportCmd.Run(cmd, args)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.PersistentFlags().StringVar(&app.AddrFlag, "addr", "127.0.0.1:5950", "voyager tcp server address")
	rootCmd.PersistentFlags().StringVar(&app.VerbosityFlag, "level", "warn", "set log level of barbecue default warn")
	cobra.CheckErr(rootCmd.Execute())
}
