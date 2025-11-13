package cmd

import (
	"barbecue/internal/app"
	"fmt"
	"os"

	Log "github.com/apatters/go-conlog"
	"github.com/spf13/cobra"
)

// reportCmd represents the report command.
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Report camera status and check if idle",
	Long: `Connect to Voyager server, retrieve camera status,
and determine if the camera is idle (temperature >= ambient and power off).`,
	Run: runReport,
}

//nolint:gochecknoinits // required for Cobra command registration
func init() {
	rootCmd.AddCommand(reportCmd)
}

func runReport(_ *cobra.Command, _ []string) {
	camera, err := app.GetCameraStatusWithConnection()
	if err != nil {
		Log.Debugf("Error retrieving camera status: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Ambient Temperature: %d\n", camera.Ambient)
	fmt.Printf("Camera Temperature: %d\n", camera.Temp)
	fmt.Printf("Camera Status: %s\n", camera.Status)
	fmt.Printf("Camera Power: %s\n", camera.Power)

	if camera.Temp >= camera.Ambient && camera.Power == "OFF" {
		fmt.Print("OK CAMERA IDLE!\n")
	}
}
