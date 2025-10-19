package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	Log "github.com/apatters/go-conlog"
	"github.com/gorilla/websocket"
)

var (
	addr      = flag.String("addr", "127.0.0.1:5950", "voyager tcp server address")
	verbosity = flag.String("level", "warn", "set log level of clandestine default warn")
)

var quit chan bool
var done chan bool

func main() {
	var c *websocket.Conn

	flag.Parse()
	setUpLogs()
	parseConfig()

	c, errcon := connectVoyager(addr)
	if errcon != nil {
		Log.Debugf("Voyager is not running or is not responding !\n")
		os.Exit(0)
	}
	defer c.Close()

	quit = make(chan bool)
	done = make(chan bool)

	go recvFromVoyager(c, done)
	remoteSetDashboard(c)
	go heartbeatVoyager(c, quit)
	time.Sleep(1 * time.Second)

	voyagerStatusDebug()

	camera := retrieveCameraStatus()

	fmt.Printf("Ambient Temperature: %d\n", camera.Ambient)
	fmt.Printf("Camera Temperature: %d\n", camera.Temp)
	fmt.Printf("Camera Status: %s\n", camera.Status)
	fmt.Printf("Camera Power: %s\n", camera.Power)

	if camera.Temp >= camera.Ambient && camera.Power == "OFF" {
		fmt.Print("OK CAMERA IDLE!\n")
	}

	done <- true
}
