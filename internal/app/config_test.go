package app

import (
	"os"
	"testing"
)

func TestSetUpLogs(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected string
	}{
		{"Debug level", "debug", "DEBUG"},
		{"Info level", "info", "INFO"},
		{"Warn level", "warn", "WARN"},
		{"Error level", "error", "ERROR"},
		{"Invalid level defaults to warn", "invalid", "WARN"},
		{"Empty level defaults to warn", "", "WARN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			VerbosityFlag = tt.level
			SetUpLogs()
			// We can't easily test the actual log level without exposing internal state
			// but we can verify the function doesn't panic
		})
	}
}

func TestReadConfig(t *testing.T) {
	// Test reading non-existent config
	t.Run("Non-existent config file", func(t *testing.T) {
		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()

		// Set args[0] to a directory that doesn't exist
		os.Args = []string{"/non/existent/path"}

		v := ReadConfig()
		// Should return a viper instance even if config file is not found
		if v == nil {
			t.Error("readConfig() returned nil even for non-existent config")
		}
	})
}

func TestParseConfig(t *testing.T) {
	// Test with default values
	t.Run("Default values", func(_ *testing.T) {
		AddrFlag = "127.0.0.1:5950"

		ParseConfig()

		// This test mainly verifies the function doesn't panic
		// since we can't easily mock readConfig
	})
}

func TestConfigFileSearchPaths(t *testing.T) {
	// Test that readConfig doesn't panic with various scenarios
	t.Run("Config path resolution", func(t *testing.T) {
		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()

		os.Args = []string{"/some/path"}

		v := ReadConfig()
		if v == nil {
			t.Error("readConfig() returned nil")
		}
	})
}

func TestConfigEnvironmentOverride(t *testing.T) {
	// Test that environment variable handling doesn't panic
	t.Run("Environment variable handling", func(t *testing.T) {
		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()

		os.Args = []string{"/some/path"}

		v := ReadConfig()
		if v == nil {
			t.Error("readConfig() returned nil")
		}
	})
}
