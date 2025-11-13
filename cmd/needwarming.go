package cmd

import (
	"barbecue/internal/app"
	"fmt"
	"os"

	Log "github.com/apatters/go-conlog"
	"github.com/spf13/cobra"
)

// needWarmingCmd represents the needWarming command.
var needWarmingCmd = &cobra.Command{
	Use:   "needWarming",
	Short: "Check if camera needs warming",
	Long: `Check if the camera cooling temperature is below the specified warm temperature threshold.
Connects to Voyager server, retrieves camera status, and compares the temperature.`,
	Run: runNeedWarming,
}

var warmTemp int

//nolint:gochecknoinits // required for Cobra command registration
func init() {
	rootCmd.AddCommand(needWarmingCmd)
	needWarmingCmd.Flags().IntVar(&warmTemp, "warm-temp", 0, "Warm temperature threshold (required)")
	if err := needWarmingCmd.MarkFlagRequired("warm-temp"); err != nil {
		panic(err)
	}
}

func runNeedWarming(_ *cobra.Command, _ []string) {
	camera, err := app.GetCameraStatusWithConnection()
	if err != nil {
		Log.Debugf("Error retrieving camera status: %v\n", err)
		os.Exit(1)
	}

	if camera.Temp < warmTemp {
		fmt.Println("Ok, camera need warming")
	} else {
		fmt.Println("Camera temperature is adequate")
	}
}
