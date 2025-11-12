package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestMainWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	// Create a temporary directory for test configs
	tempDir := t.TempDir()

	// Create a test config file
	configContent := `
voyager:
  tcpserver:
    address: 127.0.0.1
    port: 5950
`
	confDir := filepath.Join(tempDir, "conf")
	err := os.MkdirAll(confDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create conf dir: %v", err)
	}

	configPath := filepath.Join(confDir, "barbecue.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create a mock Voyager server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, upgradeErr := upgrader.Upgrade(w, r, nil)
		if upgradeErr != nil {
			t.Errorf("Failed to upgrade connection: %v", upgradeErr)
			return
		}
		defer conn.Close()

		// Handle WebSocket messages with timeout
		messageCount := 0
		for messageCount < 5 { // Limit messages to prevent infinite loop
			_, message, readErr := conn.ReadMessage()
			if readErr != nil {
				break
			}
			messageCount++

			// Parse incoming message
			msg := string(message)
			msg = strings.TrimSuffix(msg, "\r\n")

			// Handle different message types
			if strings.Contains(msg, "RemoteSetDashboardMode") {
				// Send control data response
				controlData := controldata{
					Event:     "ControlData",
					Timestamp: float64(time.Now().Unix()),
					Host:      "test",
					Inst:      1,
					VOYSTAT:   1,
					CCDTEMP:   -15.0,
					CCDPOW:    50,
					CCDSTAT:   5, // COOLED
					AFTEMP:    20.0,
					CCDCOOL:   true,
					MNTCONN:   true,
					MNTPARK:   false,
					MNTTRACK:  true,
				}
				data, _ := json.Marshal(controlData)
				conn.WriteMessage(websocket.TextMessage, []byte(string(data)+"\r\n"))
			} else if strings.Contains(msg, "Polling") {
				// Send heartbeat response
				heartbeat := event{
					Event:     "Polling",
					Timestamp: float64(time.Now().Unix()),
					Inst:      1,
				}
				data, _ := json.Marshal(heartbeat)
				conn.WriteMessage(websocket.TextMessage, []byte(string(data)+"\r\n"))
			}
		}
	}))
	defer server.Close()

	// Extract just the host:port for connectVoyager
	hostPort := strings.TrimPrefix(server.URL, "http://")
	testAddr := &hostPort

	// Test connection and data retrieval
	t.Run("Complete workflow test", func(t *testing.T) {
		// Connect to Voyager
		conn, connectErr := connectVoyager(testAddr)
		if connectErr != nil {
			t.Errorf("Failed to connect to Voyager: %v", connectErr)
			return
		}
		defer conn.Close()

		// Wrap connection with SafeConnection
		sc := NewSafeConnection(conn)
		defer sc.Close()

		// Start receiving messages with proper cleanup
		testDone := make(chan bool, 1)
		go func() {
			recvFromVoyager(sc, testDone)
		}()

		// Set dashboard mode
		remoteSetDashboard(sc)

		// Wait a short time for messages to be processed
		time.Sleep(1 * time.Second)

		// Signal shutdown
		testDone <- true
		time.Sleep(100 * time.Millisecond) // Give goroutine time to exit

		// Check if control data was updated
		if !controlDataUpdated {
			t.Error("Expected controlDataUpdated to be true")
		}

		// Retrieve camera status
		camera := retrieveCameraStatus()
		if camera.Ambient == 0 && camera.Temp == 0 {
			t.Error("Expected camera status to be populated")
		}

		// Test camera status logic
		if camera.Temp >= camera.Ambient && camera.Power == "OFF" {
			t.Log("Camera is idle - OK")
		}
	})
}

func TestCameraStatusLogic(t *testing.T) {
	// Test camera status determination logic
	tests := []struct {
		name         string
		cameraTemp   int
		ambientTemp  int
		cameraPower  string
		expectedIdle bool
	}{
		{
			name:         "Camera idle - temp >= ambient, power OFF",
			cameraTemp:   20,
			ambientTemp:  20,
			cameraPower:  "OFF",
			expectedIdle: true,
		},
		{
			name:         "Camera not idle - temp < ambient, power OFF",
			cameraTemp:   15,
			ambientTemp:  20,
			cameraPower:  "OFF",
			expectedIdle: false,
		},
		{
			name:         "Camera not idle - temp >= ambient, power ON",
			cameraTemp:   20,
			ambientTemp:  20,
			cameraPower:  "50",
			expectedIdle: false,
		},
		{
			name:         "Camera not idle - temp < ambient, power ON",
			cameraTemp:   15,
			ambientTemp:  20,
			cameraPower:  "50",
			expectedIdle: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			camera := camstatus{
				Ambient: tt.ambientTemp,
				Temp:    tt.cameraTemp,
				Power:   tt.cameraPower,
			}

			isIdle := camera.Temp >= camera.Ambient && camera.Power == "OFF"
			if isIdle != tt.expectedIdle {
				t.Errorf("Camera idle status = %v, want %v", isIdle, tt.expectedIdle)
			}
		})
	}
}

func TestConfigurationIntegration(t *testing.T) {
	// Test configuration loading and parsing integration
	t.Run("Configuration loading", func(t *testing.T) {
		// Create a temporary directory for test configs
		tempDir := t.TempDir()

		// Create a test config file
		configContent := `
voyager:
  tcpserver:
    address: 192.168.1.100
    port: 5951
`
		confDir := filepath.Join(tempDir, "conf")
		err := os.MkdirAll(confDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create conf dir: %v", err)
		}

		configPath := filepath.Join(confDir, "barbecue.yaml")
		err = os.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		// Save original args and restore after test
		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()

		os.Args = []string{tempDir}

		// Test configuration reading
		v := readConfig()
		if v == nil {
			t.Error("readConfig() returned nil")
			return
		}

		// Test configuration parsing
		defaultAddr := "127.0.0.1:5950"
		addr = &defaultAddr
		parseConfig()

		// The address should be updated from config
		// Note: This test is limited by the current implementation
		// which doesn't easily allow for mocking of readConfig
	})
}

func TestErrorHandlingIntegration(t *testing.T) {
	// Test error handling in various scenarios
	t.Run("Connection failure handling", func(t *testing.T) {
		invalidAddr := "ws://127.0.0.1:99999" // Invalid port
		conn, err := connectVoyager(&invalidAddr)

		if err == nil {
			t.Error("Expected connection to fail")
		}

		if conn != nil {
			t.Error("Expected nil connection on failure")
		}
	})

	t.Run("Malformed JSON handling", func(t *testing.T) {
		// Test parsing malformed JSON
		malformedJSON := []byte(
			`{"Event":"ControlData","Timestamp":1234567890.123,"Host":"test","Inst":1,"RUNSEQ":"test_sequence","RUNDS":"","CCDSTAT":5,"VOYSTAT":1"`, // Missing closing brace
		)

		result := parseControlData(malformedJSON)

		// Should return default values for malformed JSON
		if result.SEQRUNNING || result.DRAGRUNNING {
			t.Error("Expected default values for malformed JSON")
		}
	})

	t.Run("Empty configuration handling", func(t *testing.T) {
		// Test with empty configuration
		tempDir := t.TempDir()

		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()

		os.Args = []string{tempDir}

		v := readConfig()
		if v == nil {
			t.Error("readConfig() should return viper instance even with no config file")
		}
	})
}

func TestMessageFlowIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	// Test complete message flow between client and server
	t.Run("Message flow", func(t *testing.T) {
		// Create a mock Voyager server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{}
			conn, upgradeErr := upgrader.Upgrade(w, r, nil)
			if upgradeErr != nil {
				t.Errorf("Failed to upgrade connection: %v", upgradeErr)
				return
			}
			defer conn.Close()

			// Track received messages
			receivedMessages := []string{}

			for {
				_, message, readErr := conn.ReadMessage()
				if readErr != nil {
					break
				}

				msg := string(message)
				msg = strings.TrimSuffix(msg, "\r\n")
				receivedMessages = append(receivedMessages, msg)

				// Respond to dashboard set
				if strings.Contains(msg, "RemoteSetDashboardMode") {
					controlData := controldata{
						Event:     "ControlData",
						Timestamp: float64(time.Now().Unix()),
						Host:      "test",
						Inst:      1,
						VOYSTAT:   1,
						CCDTEMP:   -15.0,
						CCDPOW:    50,
						CCDSTAT:   5,
						AFTEMP:    20.0,
					}
					data, _ := json.Marshal(controlData)
					conn.WriteMessage(websocket.TextMessage, []byte(string(data)+"\r\n"))
				}

				// Respond to polling
				if strings.Contains(msg, "Polling") {
					heartbeat := event{
						Event:     "Polling",
						Timestamp: float64(time.Now().Unix()),
						Inst:      1,
					}
					data, _ := json.Marshal(heartbeat)
					conn.WriteMessage(websocket.TextMessage, []byte(string(data)+"\r\n"))
				}
			}

			// Verify we received expected messages
			if len(receivedMessages) < 2 {
				t.Error("Expected at least 2 messages (dashboard set and polling)")
			}
		}))
		defer server.Close()

		// Convert HTTP server URL to WebSocket URL
		hostPort := strings.TrimPrefix(server.URL, "http://")
		testAddr2 := &hostPort

		// Connect and test message flow
		conn, err := connectVoyager(testAddr2)
		if err != nil {
			t.Errorf("Failed to connect: %v", err)
			return
		}
		defer conn.Close()

		// Wrap connection with SafeConnection
		sc := NewSafeConnection(conn)
		defer sc.Close()

		// Send dashboard set message
		remoteSetDashboard(sc)

		// Wait for response
		time.Sleep(500 * time.Millisecond)

		// Send polling message
		heartbeat := event{
			Event:     "Polling",
			Timestamp: float64(time.Now().Unix()),
			Inst:      1,
		}
		data, _ := json.Marshal(heartbeat)
		sendToVoyager(sc, data)

		// Wait for response
		time.Sleep(500 * time.Millisecond)
	})
}
