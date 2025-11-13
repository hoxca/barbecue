/*
Package app implement barbecue application logic
*/
package app

import (
	"math"
	"strconv"
)

func RetrieveCameraStatus() Camstatus {
	var camstats = Camstatus{
		Ambient: GetFocuserTemperature(),
		Temp:    GetCameraTemperature(),
		Power:   GetCameraPower(),
		Status:  GetCameraStatus(),
		Cooling: GetCameraCooling(),
	}
	return camstats
}

func GetFocuserTemperature() int {
	var focusTemp int
	if ControlDataUpdated {
		focusTemp = int(math.Round(VoyagerStatus.AFTEMP))
	}
	return focusTemp
}

func GetCameraTemperature() int {
	var cameraTemp int
	if ControlDataUpdated {
		cameraTemp = int(math.Round(VoyagerStatus.CCDTEMP))
	}
	return cameraTemp
}

func GetCameraPower() string {
	var cameraPower string
	if ControlDataUpdated {
		if VoyagerStatus.CCDPOW == -123456789 {
			cameraPower = "OFF"
		} else {
			cameraPower = strconv.Itoa(VoyagerStatus.CCDPOW)
		}
	}
	return cameraPower
}

func GetCameraStatus() string {
	var cameraStatus string
	if ControlDataUpdated {
		cameraStatus = Ccdstat(VoyagerStatus.CCDSTAT).String()
	}
	return cameraStatus
}

func GetCameraCooling() bool {
	var cameraCooling bool
	if ControlDataUpdated {
		cameraCooling = VoyagerStatus.CCDCOOL
	}
	return cameraCooling
}
