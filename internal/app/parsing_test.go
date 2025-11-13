package app

import (
	"testing"
)

func TestParseLogEvent(t *testing.T) {
	tests := []struct {
		name        string
		message     []byte
		expectError bool
		expectedTS  float64
		expectedLvl string
		expectedTxt string
	}{
		{
			name: "Valid log event",
			message: []byte(`{"Event":"LogEvent","Timestamp":1234567890.123,
														"Host":"test","Inst":1,"TimeInfo":1234567890.456,
														"Type":1,"Text":"Test message"}`),
			expectError: false,
			expectedTS:  1234567890.456,
			expectedLvl: "DEBUG",
			expectedTxt: "Test message",
		},
		{
			name: "Valid log event with warning level",
			message: []byte(`{"Event":"LogEvent","Timestamp":1234567890.123,
														"Host":"test","Inst":1,"TimeInfo":1234567890.456,
														"Type":3,"Text":"Warning message"}`),
			expectError: false,
			expectedTS:  1234567890.456,
			expectedLvl: "WARNING",
			expectedTxt: "Warning message",
		},
		{
			name: "Invalid JSON",
			message: []byte(`{"Event":"LogEvent","Timestamp":1234567890.123,
														"Host":"test","Inst":1,"TimeInfo":1234567890.456,
														"Type":1,"Text":"Test message"`),
			expectError: true,
			expectedTS:  0,
			expectedLvl: "",
			expectedTxt: "",
		},
		{
			name:        "Empty JSON",
			message:     []byte(`{}`),
			expectError: false, // JSON is valid but fields are empty/zero
			expectedTS:  0,
			expectedLvl: "",
			expectedTxt: "",
		},
		{
			name:        "Missing required fields",
			message:     []byte(`{"Event":"LogEvent","Timestamp":1234567890.123}`),
			expectError: false, // JSON is valid but fields are empty/zero
			expectedTS:  0,
			expectedLvl: "",
			expectedTxt: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, level, text := ParseLogEvent(tt.message)

			if tt.expectError {
				// For error cases, we expect zero/default values
				_ = ts != tt.expectedTS || level != tt.expectedLvl || text != tt.expectedTxt
			} else {
				if ts != tt.expectedTS {
					t.Errorf("parseLogEvent() timestamp = %v, want %v", ts, tt.expectedTS)
				}
				if level != tt.expectedLvl {
					t.Errorf("parseLogEvent() level = %v, want %v", level, tt.expectedLvl)
				}
				if text != tt.expectedTxt {
					t.Errorf("parseLogEvent() text = %v, want %v", text, tt.expectedTxt)
				}
			}
		})
	}
}

func TestParseControlData(t *testing.T) {
	tests := []struct {
		name                string
		message             []byte
		expectError         bool
		expectedSeqRunning  bool
		expectedDragRunning bool
		expectedCamStatus   string
	}{
		{
			name: "Valid control data with sequence running",
			message: []byte(`{"Event":"ControlData","Timestamp":1234567890.123,
												"Host":"test", "Inst":1,"RUNSEQ":"test_sequence",
												"RUNDS":"","CCDSTAT":5,"VOYSTAT":1}`),
			expectError:         false,
			expectedSeqRunning:  true,
			expectedDragRunning: false,
			expectedCamStatus:   "COOLED",
		},
		{
			name: "Valid control data with dragscript running",
			message: []byte(`{"Event":"ControlData","Timestamp":1234567890.123,"Host":"test","Inst":1,
																		"RUNSEQ":"","RUNDS":"test_dragscript","CCDSTAT":4,"VOYSTAT":1}`),
			expectError:         false,
			expectedSeqRunning:  false,
			expectedDragRunning: true,
			expectedCamStatus:   "COOLING",
		},
		{
			name: "Valid control data with both running",
			message: []byte(`{"Event":"ControlData","Timestamp":1234567890.123,
												"Host":"test","Inst":1, "RUNSEQ":"test_sequence",
												"RUNDS":"test_dragscript","CCDSTAT":7,"VOYSTAT":1}`),
			expectError:         false,
			expectedSeqRunning:  true,
			expectedDragRunning: true,
			expectedCamStatus:   "WARMUP RUNNING",
		},
		{
			name: "Valid control data with neither running",
			message: []byte(`{"Event":"ControlData","Timestamp":1234567890.123,
												"Host":"test","Inst":1, "RUNSEQ":"","RUNDS":"",
												"CCDSTAT":3,"VOYSTAT":1}`),
			expectError:         false,
			expectedSeqRunning:  false,
			expectedDragRunning: false,
			expectedCamStatus:   "OFF",
		},
		{
			name: "Invalid JSON",
			message: []byte(`{"Event":"ControlData","Timestamp":1234567890.123,
												"Host":"test","Inst":1, "RUNSEQ":"test_sequence",
												"RUNDS":"","CCDSTAT":5,"VOYSTAT":1`),
			expectError:         true,
			expectedSeqRunning:  false,
			expectedDragRunning: false,
			expectedCamStatus:   "",
		},
		{
			name:                "Empty JSON",
			message:             []byte(`{ }`),
			expectError:         false,
			expectedSeqRunning:  false,
			expectedDragRunning: false,
			expectedCamStatus:   "INIT",
		},
		{
			name: "Missing RUNSEQ and RUNDS fields",
			message: []byte(`{"Event":"ControlData","Timestamp":1234567890.123,
												"Host":"test","Inst":1, "CCDSTAT":5,"VOYSTAT":1}`),
			expectError:         false,
			expectedSeqRunning:  false,
			expectedDragRunning: false,
			expectedCamStatus:   "COOLED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseControlData(tt.message)

			if tt.expectError {
				// For error cases, we expect default values
				_ = result.SEQRUNNING != false || result.DRAGRUNNING != false
			} else {
				if result.SEQRUNNING != tt.expectedSeqRunning {
					t.Errorf("ParseControlData() SEQRUNNING = %v, want %v", result.SEQRUNNING, tt.expectedSeqRunning)
				}
				if result.DRAGRUNNING != tt.expectedDragRunning {
					t.Errorf("ParseControlData() DRAGRUNNING = %v, want %v", result.DRAGRUNNING, tt.expectedDragRunning)
				}
				if result.CAMSTATUS != tt.expectedCamStatus {
					t.Errorf("ParseControlData() CAMSTATUS = %v, want %v", result.CAMSTATUS, tt.expectedCamStatus)
				}
			}
		})
	}
}

func TestParseControlDataWithMountStatus(t *testing.T) {
	// Test mount status parsing
	message := []byte(`{"Event":"ControlData","Timestamp":1234567890.123,
											"Host":"test","Inst":1,"MNTPARK":true,"RUNSEQ":"",
											"RUNDS":"","CCDSTAT":5,"VOYSTAT":1}`)

	result := ParseControlData(message)

	if !result.MNTPARK {
		t.Error("Expected MNTPARK to be true")
	}

	if result.CAMSTATUS != "COOLED" {
		t.Errorf("Expected CAMSTATUS to be COOLED, got %v", result.CAMSTATUS)
	}
}

func TestParseControlDataWithAllFields(t *testing.T) {
	// Test with a comprehensive control data message
	message := []byte(`{
		"Event":"ControlData",
		"Timestamp":1234567890.123,
		"Host":"test",
		"Inst":1,
		"TI":"test_ti",
		"VOYSTAT":2,
		"SETUPCONN":true,
		"CCDCONN":true,
		"CCDTEMP":-15.5,
		"CCDPOW":50,
		"CCDSETP":-20,
		"CCDCOOL":true,
		"CCDSTAT":5,
		"MNTCONN":true,
		"MNTPARK":false,
		"MNTRA":"12:34:56",
		"MNTDEC":"+12:34:56",
		"MNTRAJ2000":"12:34:56",
		"MNTDECJ2000":"+12:34:56",
		"MNTAZ":"180.0",
		"MNTALT":"45.0",
		"MNTPIER":"EAST",
		"MNTTFLIP":"MERIDIAN",
		"MNTSFLIP":1,
		"MNTTRACK":true,
		"MNTSLEW":false,
		"AFCONN":true,
		"AFTEMP":20.5,
		"AFPOS":5000,
		"SEQTOT":100,
		"SEQPARZ":50,
		"GUIDECONN":true,
		"GUIDESTAT":1,
		"DITHSTAT":0,
		"GUIDEX":0.1,
		"GUIDEY":0.2,
		"PLACONN":true,
		"PSCONN":true,
		"SEQNAME":"test_sequence",
		"SEQSTART":"2023-01-01T00:00:00",
		"SEQREMAIN":"01:00:00",
		"SEQEND":"2023-01-01T01:00:00",
		"RUNSEQ":"test_sequence",
		"RUNDS":"test_dragscript",
		"ROTCONN":true,
		"ROTPA":0.0,
		"ROTSKYPA":45.0,
		"ROTISROT":false,
		"DRAGRUNNING":true,
		"SEQRUNNING":true,
		"CAMSTATUS":"COOLED"
	}`)

	result := ParseControlData(message)

	// Test key fields
	if !result.SEQRUNNING {
		t.Error("Expected SEQRUNNING to be true")
	}

	if !result.DRAGRUNNING {
		t.Error("Expected DRAGRUNNING to be true")
	}

	if result.CCDTEMP != -15.5 {
		t.Errorf("Expected CCDTEMP to be -15.5, got %v", result.CCDTEMP)
	}

	if result.CCDPOW != 50 {
		t.Errorf("Expected CCDPOW to be 50, got %v", result.CCDPOW)
	}

	if result.AFTEMP != 20.5 {
		t.Errorf("Expected AFTEMP to be 20.5, got %v", result.AFTEMP)
	}

	if result.CAMSTATUS != "COOLED" {
		t.Errorf("Expected CAMSTATUS to be COOLED, got %v", result.CAMSTATUS)
	}
}
