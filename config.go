package main

import (
	"fmt"
	"os"
	"path/filepath"

	Log "github.com/apatters/go-conlog"
	"github.com/spf13/viper"
)

func parseConfig() {
	viper := readConfig()
	if viper == nil {
		fmt.Println("null config")
	}

	var viperVoyAddr string
	var viperVoyPort int

	if viper.IsSet("voyager.tcpserver.address") {
		viperVoyAddr = viper.GetString("voyager.tcpserver.address")
	} else {
		viperVoyAddr = "127.0.0.1"
	}

	if viper.IsSet("voyager.tcpserver.port") {
		viperVoyPort = viper.GetInt("voyager.tcpserver.port")
	} else {
		viperVoyPort = 5950
	}

	if *addr == "127.0.0.1:5950" &&
		(viper.IsSet("voyager.tcpserver.address") || viper.IsSet("voyager.tcpserver.port")) {
		*addr = fmt.Sprintf("%s:%d", viperVoyAddr, viperVoyPort)
	}

	Log.Debugf("voyager addr: %s", *addr)
}

func setUpLogs() {
	formatter := Log.NewStdFormatter()
	formatter.Options.LogLevelFmt = Log.LogLevelFormatLongTitle
	Log.SetFormatter(formatter)
	switch *verbosity {
	case "debug":
		Log.SetLevel(Log.DebugLevel)
	case "info":
		Log.SetLevel(Log.InfoLevel)
	case "warn":
		Log.SetLevel(Log.WarnLevel)
	case "error":
		Log.SetLevel(Log.ErrorLevel)
	default:
		Log.SetLevel(Log.WarnLevel)
	}
}

func readConfig() *viper.Viper {
	var err error
	v := viper.New()
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	confdir := fmt.Sprintf("%s/conf", dir)

	// if we came from bin directory
	confdir1 := fmt.Sprintf("%s/../conf", dir)
	confdir2 := "./conf"
	// Search yaml config file in program path with name "barbecue.yaml".
	v.AddConfigPath(confdir)
	v.AddConfigPath(confdir1)
	v.AddConfigPath(confdir2)
	v.SetConfigType("yaml")
	v.SetConfigName("barbecue.yml")

	viper.AutomaticEnv() // read in environment variables that match

	// Find and read the config file
	err = v.ReadInConfig()
	if err != nil {
		Log.Warn(fmt.Errorf("%w", err))
		Log.Warn("Will use localhost default configuration")
	}
	Log.Debugf("Using config file: %s", v.ConfigFileUsed())
	return v
}
