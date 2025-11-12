package main

import (
	"testing"
)

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		level    loglevel
		expected string
	}{
		{"Debug level", loglevel(1), "DEBUG"},
		{"Info level", loglevel(2), "INFO"},
		{"Warning level", loglevel(3), "WARNING"},
		{"Critical level", loglevel(4), "CRITICAL"},
		{"Title level", loglevel(5), "TITLE"},
		{"Subtitle level", loglevel(6), "SUBTITLE"},
		{"Event level", loglevel(7), "EVENT"},
		{"Request level", loglevel(8), "REQUEST"},
		{"Emergency level", loglevel(9), "EMERGENCY"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.level.String()
			if result != tt.expected {
				t.Errorf("loglevel.String() = %v, want %v", result, tt.expected)
			}
		})
	}

	// Test out of bounds (should panic or return empty)
	t.Run("Out of bounds", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic for out of bounds loglevel")
			}
		}()
		_ = loglevel(-1).String()
	})
}

func TestCcdStat_String(t *testing.T) {
	tests := []struct {
		name     string
		status   ccdstat
		expected string
	}{
		{"Init status", ccdstat(0), "INIT"},
		{"Undef status", ccdstat(1), "UNDEF"},
		{"No cooler status", ccdstat(2), "NO COOLER"},
		{"Off status", ccdstat(3), "OFF"},
		{"Cooling status", ccdstat(4), "COOLING"},
		{"Cooled status", ccdstat(5), "COOLED"},
		{"Timeout status", ccdstat(6), "TIMEOUT"},
		{"Warmup running status", ccdstat(7), "WARMUP RUNNING"},
		{"Warmup end status", ccdstat(8), "WARMUP END"},
		{"Error status", ccdstat(9), "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.String()
			if result != tt.expected {
				t.Errorf("ccdstat.String() = %v, want %v", result, tt.expected)
			}
		})
	}

	// Test out of bounds (should panic or return empty)
	t.Run("Out of bounds", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic for out of bounds ccdstat")
			}
		}()
		_ = ccdstat(-1).String()
	})
}

func TestGetCameraPower(t *testing.T) {
	// Reset control data flag
	controlDataUpdated = false

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
			controlDataUpdated = tt.controlUpdated
			voyagerStatus.CCDPOW = tt.ccdPow
			result := getCameraPower()
			if result != tt.expected {
				t.Errorf("getCameraPower() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetCameraStatus(t *testing.T) {
	// Reset control data flag
	controlDataUpdated = false

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
			controlDataUpdated = tt.controlUpdated
			voyagerStatus.CCDSTAT = tt.ccdStat
			result := getCameraStatus()
			if result != tt.expected {
				t.Errorf("getCameraStatus() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetFocuserTemperature(t *testing.T) {
	// Reset control data flag
	controlDataUpdated = false

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
			controlDataUpdated = tt.controlUpdated
			voyagerStatus.AFTEMP = tt.afTemp
			result := getFocuserTemperature()
			if result != tt.expected {
				t.Errorf("getFocuserTemperature() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetCameraTemperature(t *testing.T) {
	// Reset control data flag
	controlDataUpdated = false

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
			controlDataUpdated = tt.controlUpdated
			voyagerStatus.CCDTEMP = tt.ccdTemp
			result := getCameraTemperature()
			if result != tt.expected {
				t.Errorf("getCameraTemperature() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRetrieveCameraStatus(t *testing.T) {
	// Reset control data flag
	controlDataUpdated = false

	// Test when control data is not updated
	t.Run("Control data not updated", func(t *testing.T) {
		controlDataUpdated = false
		result := retrieveCameraStatus()
		expected := camstatus{Ambient: 0, Temp: 0, Power: "", Status: ""}
		if result != expected {
			t.Errorf("retrieveCameraStatus() = %v, want %v", result, expected)
		}
	})

	// Test when control data is updated
	t.Run("Control data updated", func(t *testing.T) {
		controlDataUpdated = true
		voyagerStatus.AFTEMP = 20.6
		voyagerStatus.CCDTEMP = -15.4
		voyagerStatus.CCDPOW = 50
		voyagerStatus.CCDSTAT = 5 // COOLED

		result := retrieveCameraStatus()
		expected := camstatus{
			Ambient: 21,  // 20.6 rounds to 21
			Temp:    -15, // -15.4 rounds to -15
			Power:   "50",
			Status:  "COOLED",
		}
		if result != expected {
			t.Errorf("retrieveCameraStatus() = %v, want %v", result, expected)
		}
	})

	// Test with camera OFF
	t.Run("Camera OFF", func(t *testing.T) {
		controlDataUpdated = true
		voyagerStatus.AFTEMP = 20.6
		voyagerStatus.CCDTEMP = -15.4
		voyagerStatus.CCDPOW = -123456789 // OFF
		voyagerStatus.CCDSTAT = 3         // OFF

		result := retrieveCameraStatus()
		expected := camstatus{
			Ambient: 21,
			Temp:    -15,
			Power:   "OFF",
			Status:  "OFF",
		}
		if result != expected {
			t.Errorf("retrieveCameraStatus() = %v, want %v", result, expected)
		}
	})
}
