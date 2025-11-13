package cmd

import (
	"barbecue/internal/app"
	"fmt"
	"os"

	Log "github.com/apatters/go-conlog"
	"github.com/spf13/cobra"
)

// needWarmingCmd implements the needWarming command.
var needWarmingCmd = &cobra.Command{
	Use:   "needWarming",
	Short: "Check if camera needs warming",
	Long: `Check if the camera cooling temperature is below the specified warm temperature threshold.
Connects to Voyager server, retrieves camera status, and compares the temperature.`,
	Run: runNeedWarming,
}

var warmTemp int

// required for Cobra command registration.
func init() {
	rootCmd.AddCommand(needWarmingCmd)
	needWarmingCmd.Flags().IntVar(&warmTemp, "warm-temp", 0, "Warm temperature threshold (required)")
	if err := needWarmingCmd.MarkFlagRequired("warm-temp"); err != nil {
		panic(err)
	}
}

func runNeedWarming(cmd *cobra.Command, args []string) {
	camera, err := app.GetCameraStatusWithConnection()
	if err != nil {
		Log.Debugf("Error retrieving camera status: %v\n", err)
		os.Exit(1)
	}

	if (camera.Status == "COOLING" || camera.Status == "COOLED" || camera.Status == "TIMEOUT") && camera.Temp < warmTemp {
		fmt.Println("Ok, camera need warming")
	} else {
		fmt.Println("Camera temperature is above target warming temp")
	}
}
