package app

import (
	"testing"

	"github.com/gorilla/websocket"
)

func TestSafeConnection(t *testing.T) {
	// Test SafeConnection basic functionality
	t.Run("SafeConnection creation and basic operations", func(t *testing.T) {
		// Create a mock connection for testing the wrapper logic
		mockConn := &MockWebSocketConn{}
		sc := NewSafeConnection(mockConn)

		// Test initial state
		if sc.IsClosed() {
			t.Error("Expected SafeConnection to be open initially")
		}

		// Test close operation
		sc.Close()
		if !sc.IsClosed() {
			t.Error("Expected SafeConnection to be closed after Close()")
		}

		// Test write after close should return error
		err := sc.WriteMessage(websocket.TextMessage, []byte("test after close"))
		if err == nil {
			t.Error("Expected error when writing to closed connection")
		}
	})

	t.Run("Concurrent access safety", func(t *testing.T) {
		mockConn := &MockWebSocketConn{}
		sc := NewSafeConnection(mockConn)

		// Test concurrent close operations
		testDone := make(chan bool, 3)

		// Start multiple goroutines trying to close
		// connection
		for range 3 {
			go func() {
				sc.Close()
				testDone <- true
			}()
		}

		// Wait for all goroutines to complete
		for range 3 {
			<-testDone
		}

		// Connection should be closed
		if !sc.IsClosed() {
			t.Error("Expected SafeConnection to be closed")
		}

		// Multiple closes should not cause panic
		sc.Close()
		sc.Close()
	})
}
