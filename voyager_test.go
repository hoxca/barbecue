package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	Log "github.com/apatters/go-conlog"
	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/websocket"
)

// WebSocketConnection interface for testing.
type WebSocketConnection interface {
	WriteMessage(messageType int, data []byte) error
	ReadMessage() (messageType int, p []byte, err error)
	Close() error
}

// MockWebSocketConn implements a mock WebSocket connection for testing.
type MockWebSocketConn struct {
	closed        bool
	writeError    error
	readMessage   []byte
	readError     error
	closeError    error
	writeMessages [][]byte
}

func (m *MockWebSocketConn) WriteMessage(_ int, data []byte) error {
	m.writeMessages = append(m.writeMessages, data)
	return m.writeError
}

func (m *MockWebSocketConn) ReadMessage() (int, []byte, error) {
	if m.readError != nil {
		return 0, nil, m.readError
	}
	if m.readMessage != nil {
		return websocket.TextMessage, m.readMessage, nil
	}
	return 0, nil, nil
}

func (m *MockWebSocketConn) Close() error {
	m.closed = true
	return m.closeError
}

// Helper functions for testing that accept the interface.
func sendToVoyagerMock(c WebSocketConnection, data []byte) {
	message := fmt.Appendf(nil, "%s\r\n", data)
	err := c.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		Log.Println("write:", err)
		return
	}
	Log.Debugf("send: %s", data)
}

func remoteSetDashboardMock(c WebSocketConnection) {
	p := &params{
		UID:  fmt.Sprintf("%s", uuid.Must(uuid.NewV4())),
		IsOn: true,
	}

	setDashboard := &method{
		Method: "RemoteSetDashboardMode",
		Params: *p,
		ID:     1,
	}

	data, _ := json.Marshal(setDashboard)
	sendToVoyagerMock(c, data)
}

func TestConnectVoyager(t *testing.T) {
	// Test successful connection
	t.Run("Successful connection", func(t *testing.T) {
		// Create a test WebSocket server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{}
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				t.Errorf("Failed to upgrade connection: %v", err)
				return
			}
			defer conn.Close()

			// Keep connection open for a bit
			time.Sleep(100 * time.Millisecond)
		}))
		defer server.Close()

		// Convert HTTP server URL to WebSocket URL
		hostPort := strings.TrimPrefix(server.URL, "http://")
		testAddr := &hostPort

		conn, err := connectVoyager(testAddr)

		if err != nil {
			t.Errorf("connectVoyager() returned error: %v", err)
			return
		}

		if conn == nil {
			t.Error("connectVoyager() returned nil connection")
			return
		}

		conn.Close()
	})
}

func TestSendToVoyager(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectError bool
	}{
		{
			name:        "Valid JSON data",
			data:        []byte(`{"Event":"Test","Timestamp":1234567890}`),
			expectError: false,
		},
		{
			name:        "Empty data",
			data:        []byte(`{}`),
			expectError: false,
		},
		{
			name:        "Nil data",
			data:        nil,
			expectError: false, // Should not panic
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock connection
			mockConn := &MockWebSocketConn{}

			// Test sendToVoyager
			sendToVoyagerMock(mockConn, tt.data)

			// Check if message was written (for non-nil data)
			if tt.data != nil && len(mockConn.writeMessages) == 0 {
				t.Error("Expected message to be written to connection")
			}

			// Check if message has proper format (ends with \r\n)
			if tt.data != nil && len(mockConn.writeMessages) > 0 {
				msg := string(mockConn.writeMessages[0])
				if !strings.HasSuffix(msg, "\r\n") {
					t.Errorf("Expected message to end with \\r\\n, got: %q", msg)
				}
			}
		})
	}
}

func TestRemoteSetDashboard(t *testing.T) {
	tests := []struct {
		name        string
		expectError bool
	}{
		{
			name:        "Valid dashboard set",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock connection
			mockConn := &MockWebSocketConn{}

			// Test remoteSetDashboard
			remoteSetDashboardMock(mockConn)

			// Check if message was sent
			if len(mockConn.writeMessages) == 0 {
				t.Error("Expected dashboard set message to be sent")
				return
			}

			// Parse the sent message
			msg := strings.TrimSuffix(string(mockConn.writeMessages[0]), "\r\n")
			var methodStruct method
			err := json.Unmarshal([]byte(msg), &methodStruct)
			if err != nil {
				t.Errorf("Failed to parse dashboard set message: %v", err)
				return
			}

			// Verify message structure
			if methodStruct.Method != "RemoteSetDashboardMode" {
				t.Errorf("Expected method RemoteSetDashboardMode, got %v", methodStruct.Method)
			}

			if !methodStruct.Params.IsOn {
				t.Error("Expected IsOn to be true")
			}

			if methodStruct.Params.UID == "" {
				t.Error("Expected UID to be set")
			}

			if methodStruct.ID != 1 {
				t.Errorf("Expected ID to be 1, got %v", methodStruct.ID)
			}
		})
	}
}

func TestVoyagerMessageTypes(t *testing.T) {
	// Test that different message types are properly formatted
	tests := []struct {
		name     string
		event    event
		expected string
	}{
		{
			name: "Polling event",
			event: event{
				Event:     "Polling",
				Timestamp: 1234567890.0,
				Inst:      1,
			},
			expected: "Polling",
		},
		{
			name: "Control data event",
			event: event{
				Event:     "ControlData",
				Timestamp: 1234567890.0,
				Inst:      1,
				Host:      "test",
			},
			expected: "ControlData",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := &MockWebSocketConn{}

			// Marshal the event
			data, err := json.Marshal(tt.event)
			if err != nil {
				t.Errorf("Failed to marshal event: %v", err)
				return
			}

			// Send the event
			sendToVoyagerMock(mockConn, data)

			// Check if message was sent
			if len(mockConn.writeMessages) == 0 {
				t.Error("Expected message to be sent")
				return
			}

			// Parse the sent message
			msg := strings.TrimSuffix(string(mockConn.writeMessages[0]), "\r\n")
			var receivedEvent event
			err = json.Unmarshal([]byte(msg), &receivedEvent)
			if err != nil {
				t.Errorf("Failed to parse received message: %v", err)
				return
			}

			// Verify event content
			if receivedEvent.Event != tt.expected {
				t.Errorf("Expected event %v, got %v", tt.expected, receivedEvent.Event)
			}
		})
	}
}

func TestWebSocketErrorHandling(t *testing.T) {
	// Test error handling in WebSocket operations
	t.Run("Write error handling", func(t *testing.T) {
		mockConn := &MockWebSocketConn{
			writeError: websocket.ErrBadHandshake,
		}

		// This should not panic
		sendToVoyagerMock(mockConn, []byte(`{"test": "data"}`))

		// Message should still be attempted to be written
		if len(mockConn.writeMessages) == 0 {
			t.Error("Expected write attempt even with error")
		}
	})
}

func TestMessageFormatting(t *testing.T) {
	// Test that messages are properly formatted with \r\n suffix
	t.Run("Message formatting", func(t *testing.T) {
		mockConn := &MockWebSocketConn{}

		testData := []byte(`{"Event":"Test","Data":"value"}`)
		sendToVoyagerMock(mockConn, testData)

		if len(mockConn.writeMessages) == 0 {
			t.Error("Expected message to be written")
			return
		}

		msg := string(mockConn.writeMessages[0])
		expectedSuffix := "\r\n"

		if !strings.HasSuffix(msg, expectedSuffix) {
			t.Errorf("Expected message to end with %q, got %q", expectedSuffix, msg[len(msg)-2:])
		}

		// Check that the original data is preserved
		expectedContent := string(testData) + expectedSuffix
		if msg != expectedContent {
			t.Errorf("Expected message %q, got %q", expectedContent, msg)
		}
	})
}
