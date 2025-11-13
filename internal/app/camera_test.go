package app

import (
	"testing"
)

func TestCcdStat_String(t *testing.T) {
	tests := []struct {
		name     string
		status   Ccdstat
		expected string
	}{
		{"Init status", Ccdstat(0), "INIT"},
		{"Undef status", Ccdstat(1), "UNDEF"},
		{"No cooler status", Ccdstat(2), "NO COOLER"},
		{"Off status", Ccdstat(3), "OFF"},
		{"Cooling status", Ccdstat(4), "COOLING"},
		{"Cooled status", Ccdstat(5), "COOLED"},
		{"Timeout status", Ccdstat(6), "TIMEOUT"},
		{"Warmup running status", Ccdstat(7), "WARMUP RUNNING"},
		{"Warmup end status", Ccdstat(8), "WARMUP END"},
		{"Error status", Ccdstat(9), "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.String()
			if result != tt.expected {
				t.Errorf("Ccdstat.String() = %v, want %v", result, tt.expected)
			}
		})
	}

	// Test out of bounds (should panic or return empty)
	t.Run("Out of bounds", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic for out of bounds Ccdstat")
			}
		}()
		_ = Ccdstat(-1).String()
	})
}

func TestGetCameraPower(t *testing.T) {
	// Reset control data flag
	ControlDataUpdated = false

	tests := []struct {
		name           string
		ccdPow         int
		controlUpdated bool
		expected       string
	}{
		{"Control data not updated", 100, false, ""},
		{"Camera OFF", -123456789, true, "OFF"},
		{"Camera ON with power 50", 50, true, "50"},
		{"Camera ON with power 0", 0, true, "0"},
		{"Camera ON with power 100", 100, true, "100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ControlDataUpdated = tt.controlUpdated
			VoyagerStatus.CCDPOW = tt.ccdPow
			result := GetCameraPower()
			if result != tt.expected {
				t.Errorf("GetCameraPower() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetCameraStatus(t *testing.T) {
	// Reset control data flag
	ControlDataUpdated = false

	tests := []struct {
		name           string
		ccdStat        int
		controlUpdated bool
		expected       string
	}{
		{"Control data not updated", 0, false, ""},
		{"Init status", 0, true, "INIT"},
		{"Undef status", 1, true, "UNDEF"},
		{"No cooler status", 2, true, "NO COOLER"},
		{"Off status", 3, true, "OFF"},
		{"Cooling status", 4, true, "COOLING"},
		{"Cooled status", 5, true, "COOLED"},
		{"Timeout status", 6, true, "TIMEOUT"},
		{"Warmup running status", 7, true, "WARMUP RUNNING"},
		{"Warmup end status", 8, true, "WARMUP END"},
		{"Error status", 9, true, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ControlDataUpdated = tt.controlUpdated
			VoyagerStatus.CCDSTAT = tt.ccdStat
			result := GetCameraStatus()
			if result != tt.expected {
				t.Errorf("GetCameraStatus() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetFocuserTemperature(t *testing.T) {
	// Reset control data flag
	ControlDataUpdated = false

	tests := []struct {
		name           string
		afTemp         float64
		controlUpdated bool
		expected       int
	}{
		{"Control data not updated", 20.5, false, 0},
		{"Temperature 20.4 rounds to 20", 20.4, true, 20},
		{"Temperature 20.5 rounds to 21", 20.5, true, 21},
		{"Temperature 20.6 rounds to 21", 20.6, true, 21},
		{"Temperature -10.2 rounds to -10", -10.2, true, -10},
		{"Temperature -10.8 rounds to -11", -10.8, true, -11},
		{"Temperature 0.0 rounds to 0", 0.0, true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ControlDataUpdated = tt.controlUpdated
			VoyagerStatus.AFTEMP = tt.afTemp
			result := GetFocuserTemperature()
			if result != tt.expected {
				t.Errorf("GetFocuserTemperature() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetCameraTemperature(t *testing.T) {
	// Reset control data flag
	ControlDataUpdated = false

	tests := []struct {
		name           string
		ccdTemp        float64
		controlUpdated bool
		expected       int
	}{
		{"Control data not updated", -15.4, false, 0},
		{"Temperature -15.4 rounds to -15", -15.4, true, -15},
		{"Temperature -15.5 rounds to -16", -15.5, true, -16},
		{"Temperature -15.6 rounds to -16", -15.6, true, -16},
		{"Temperature 0.2 rounds to 0", 0.2, true, 0},
		{"Temperature 0.8 rounds to 1", 0.8, true, 1},
		{"Temperature 10.0 rounds to 10", 10.0, true, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ControlDataUpdated = tt.controlUpdated
			VoyagerStatus.CCDTEMP = tt.ccdTemp
			result := GetCameraTemperature()
			if result != tt.expected {
				t.Errorf("GetCameraTemperature() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRetrieveCameraStatus(t *testing.T) {
	// Reset control data flag
	ControlDataUpdated = false

	// Test when control data is not updated
	t.Run("Control data not updated", func(t *testing.T) {
		ControlDataUpdated = false
		result := RetrieveCameraStatus()
		expected := Camstatus{Ambient: 0, Temp: 0, Power: "", Status: ""}
		if result != expected {
			t.Errorf("RetrieveCameraStatus() = %v, want %v", result, expected)
		}
	})

	// Test when control data is updated
	t.Run("Control data updated", func(t *testing.T) {
		ControlDataUpdated = true
		VoyagerStatus.AFTEMP = 20.6
		VoyagerStatus.CCDTEMP = -15.4
		VoyagerStatus.CCDPOW = 50
		VoyagerStatus.CCDSTAT = 5 // COOLED

		result := RetrieveCameraStatus()
		expected := Camstatus{
			Ambient: 21,  // 20.6 rounds to 21
			Temp:    -15, // -15.4 rounds to -15
			Power:   "50",
			Status:  "COOLED",
		}
		if result != expected {
			t.Errorf("RetrieveCameraStatus() = %v, want %v", result, expected)
		}
	})

	// Test with camera OFF
	t.Run("Camera OFF", func(t *testing.T) {
		ControlDataUpdated = true
		VoyagerStatus.AFTEMP = 20.6
		VoyagerStatus.CCDTEMP = -15.4
		VoyagerStatus.CCDPOW = -123456789 // OFF
		VoyagerStatus.CCDSTAT = 3         // OFF

		result := RetrieveCameraStatus()
		expected := Camstatus{
			Ambient: 21,
			Temp:    -15,
			Power:   "OFF",
			Status:  "OFF",
		}
		if result != expected {
			t.Errorf("RetrieveCameraStatus() = %v, want %v", result, expected)
		}
	})
}
