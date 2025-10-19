package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	Log "github.com/apatters/go-conlog"
	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/websocket"
)

type event struct {
	Event     string  `json:"Event"`
	Timestamp float64 `json:"Timestamp"`
	Host      string  `json:"Host,omitempty"`
	Inst      int     `json:"Inst"`
}

type controldata struct {
	Event       string  `json:"Event"`
	Timestamp   float64 `json:"Timestamp"`
	Host        string  `json:"Host"`
	Inst        int     `json:"Inst"`
	TI          string  `json:"TI"`
	VOYSTAT     int     `json:"VOYSTAT"`
	SETUPCONN   bool    `json:"SETUPCONN"`
	CCDCONN     bool    `json:"CCDCONN"`
	CCDTEMP     float64 `json:"CCDTEMP"`
	CCDPOW      int     `json:"CCDPOW"`
	CCDSETP     int     `json:"CCDSETP"`
	CCDCOOL     bool    `json:"CCDCOOL"`
	CCDSTAT     int     `json:"CCDSTAT"`
	MNTCONN     bool    `json:"MNTCONN"`
	MNTPARK     bool    `json:"MNTPARK"`
	MNTRA       string  `json:"MNTRA"`
	MNTDEC      string  `json:"MNTDEC"`
	MNTRAJ2000  string  `json:"MNTRAJ2000"`
	MNTDECJ2000 string  `json:"MNTDECJ2000"`
	MNTAZ       string  `json:"MNTAZ"`
	MNTALT      string  `json:"MNTALT"`
	MNTPIER     string  `json:"MNTPIER"`
	MNTTFLIP    string  `json:"MNTTFLIP"`
	MNTSFLIP    int     `json:"MNTSFLIP"`
	MNTTRACK    bool    `json:"MNTTRACK"`
	MNTSLEW     bool    `json:"MNTSLEW"`
	AFCONN      bool    `json:"AFCONN"`
	AFTEMP      float64 `json:"AFTEMP"`
	AFPOS       int     `json:"AFPOS"`
	SEQTOT      int     `json:"SEQTOT"`
	SEQPARZ     int     `json:"SEQPARZ"`
	GUIDECONN   bool    `json:"GUIDECONN"`
	GUIDESTAT   int     `json:"GUIDESTAT"`
	DITHSTAT    int     `json:"DITHSTAT"`
	GUIDEX      float64 `json:"GUIDEX"`
	GUIDEY      float64 `json:"GUIDEY"`
	PLACONN     bool    `json:"PLACONN"`
	PSCONN      bool    `json:"PSCONN"`
	SEQNAME     string  `json:"SEQNAME"`
	SEQSTART    string  `json:"SEQSTART"`
	SEQREMAIN   string  `json:"SEQREMAIN"`
	SEQEND      string  `json:"SEQEND"`
	RUNSEQ      string  `json:"RUNSEQ"`
	RUNDS       string  `json:"RUNDS"`
	ROTCONN     bool    `json:"ROTCONN"`
	ROTPA       float64 `json:"ROTPA"`
	ROTSKYPA    float64 `json:"ROTSKYPA"`
	ROTISROT    bool    `json:"ROTISROT"`
	DRAGRUNNING bool    `json:"DRAGRUNNING"`
	SEQRUNNING  bool    `json:"SEQRUNNING"`
	CAMSTATUS   string  `json:"CAMSTATUS"`
}

var controlDataUpdated = false

type method struct {
	Method string `json:"method"`
	Params params `json:"params"`
	ID     int    `json:"id"`
}

type params struct {
	UID         string `json:"UID"`
	IsOn        bool   `json:"IsOn"`
	Level       *int   `json:"Level,omitempty"`
	IsHalt      bool   `json:"IsHalt,omitempty"`
	CommandType int    `json:"CommandType,omitempty"`
	IsSetPoint  bool   `json:"IsSetPoint,omitempty"`
	IsCoolDown  bool   `json:"IsCoolDown,omitempty"`
	IsASync     bool   `json:"IsASync,omitempty"`
	IsWarmup    bool   `json:"IsWarmup,omitempty"`
	IsCoolerOFF bool   `json:"IsCoolerOFF,omitempty"`
	Temperature int    `json:"Temperature,omitempty"`
}

type loglevel int

func (l loglevel) String() string {
	return [...]string{
		"DEBUG",
		"INFO",
		"WARNING",
		"CRITICAL",
		"TITLE",
		"SUBTITLE",
		"EVENT",
		"REQUEST",
		"EMERGENCY",
	}[l-1]
}

var voyagerStatus controldata

type ccdstat int

func (cmos ccdstat) String() string {
	return [...]string{
		"INIT",
		"UNDEF",
		"NO COOLER",
		"OFF",
		"COOLING",
		"COOLED",
		"TIMEOUT",
		"WARMUP RUNNING",
		"WARMUP END",
		"ERROR",
	}[cmos]
}

type camstatus struct {
	Ambient int
	Temp    int
	Power   string
	Status  string
}

/*
var quit chan bool
var done chan bool

func startVoyagerClient() *websocket.Conn {
	c, errcon := connectVoyager(addr)
	if errcon != nil {
		Log.Debugf("Voyager is not running or is not responding !\n")
		os.Exit(0)
	}
	defer c.Close()

	return c

	quit = make(chan bool)
	done = make(chan bool)

	go recvFromVoyager(c, done)
	remoteSetDashboard(c)
	go heartbeatVoyager(c, quit)
	time.Sleep(2 * time.Second)

	voyagerStatusDebug()
	return c
}
*/

func connectVoyager(addr *string) (*websocket.Conn, error) {
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/"}
	Log.Debugf("connecting to %s", u.String())

	websocket.DefaultDialer.HandshakeTimeout = 1 * time.Second
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		Log.Printf("Can't connect, verify Voyager address or tcp port in the Voyager configuration\n")
		// Log.Fatal("Critical: ", err)
	}
	return c, err
}

func remoteSetDashboard(c *websocket.Conn) {
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
	sendToVoyager(c, data)
}

func sendToVoyager(c *websocket.Conn, data []byte) {
	lastpoll = time.Now()
	err := c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%s\r\n", data)))
	if err != nil {
		Log.Println("write:", err)
		return
	}
	Log.Debugf("send: %s", data)
	time.Sleep(1 * time.Second)
}

func retrieveCameraStatus() camstatus {
	var camstats = camstatus{
		Ambient: getFocuserTemperature(),
		Temp:    getCameraTemperature(),
		Power:   getCameraPower(),
		Status:  getCameraStatus(),
	}
	return camstats
}

func getFocuserTemperature() int {
	var focusTemp int
	if controlDataUpdated {
		focusTemp = int(math.Round(voyagerStatus.AFTEMP))
	}
	return focusTemp
}

func getCameraTemperature() int {
	var cameraTemp int
	if controlDataUpdated {
		cameraTemp = int(math.Round(voyagerStatus.CCDTEMP))
	}
	return cameraTemp
}

func getCameraPower() string {
	var cameraPower string
	if controlDataUpdated {
		if voyagerStatus.CCDPOW == -123456789 {
			cameraPower = "OFF"
		} else {
			cameraPower = strconv.Itoa(voyagerStatus.CCDPOW)
		}
	}
	return cameraPower
}

func getCameraStatus() string {
	var cameraStatus string
	if controlDataUpdated {
		cameraStatus = ccdstat(voyagerStatus.CCDSTAT).String()
	}
	return cameraStatus
}

func voyagerStatusDebug() {
	if controlDataUpdated {
		Log.Info("Voyager Status:")
		Log.Infof("  Voyager    status: %d", voyagerStatus.VOYSTAT)
		Log.Infof("  Camera     status: %s", ccdstat(voyagerStatus.CCDSTAT).String())
		Log.Infof("  Camera    cooling: %s", strconv.FormatBool(voyagerStatus.CCDCOOL))
		Log.Infof("  Camera   ccd temp: %f", voyagerStatus.CCDTEMP)
		Log.Infof("  Camera  ccd power: %d", voyagerStatus.CCDPOW)
		Log.Infof("  Focuser      temp: %f\n", voyagerStatus.AFTEMP)
	}
}

func recvFromVoyager(c *websocket.Conn, done chan bool) {
	for {
		select {
		case <-done:
			Log.Debugf("Quit recv loop!")
			quit <- true
			return
		default:
			_, message, err := c.ReadMessage()
			if err != nil {
				Log.Warn("read:", err)
				quit <- true
				return
			}
			// parse incoming message
			msg := string(message)
			switch {
			case strings.Contains(msg, `"Event":"ControlData"`):
				//              if !controlDataUpdated {
				Log.Debugf("recv msg: %s", strings.TrimRight(msg, "\r\n"))
				voyagerStatus = parseControlData(message)
			//      }
			case strings.Contains(msg, `"Event":"LogEvent"`):
				ts, level, logline := parseLogEvent(message)
				Log.Debugf("recv log: %.5f %s %s", ts, level, logline)
			case strings.Contains(msg, `"Event":"RemoteActionResult"`):
				Log.Debugf("recv result: %s", strings.TrimRight(msg, "\r\n"))
			case strings.Contains(msg, `"Event":"Version"`):
				Log.Debugf("recv version: %s", strings.TrimRight(msg, "\r\n"))
			case strings.Contains(msg, `"Event":"VikingManaged"`):
				Log.Debugf("recv viking: %s", strings.TrimRight(msg, "\r\n"))
			default:
				Log.Debugf("recv not managed: %s", strings.TrimRight(msg, "\r\n"))
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func parseLogEvent(message []byte) (float64, string, string) {
	type logEvent struct {
		Event     string   `json:"Event"`
		Timestamp float64  `json:"Timestamp"`
		Host      string   `json:"Host"`
		Inst      int      `json:"Inst"`
		TimeInfo  float64  `json:"TimeInfo"`
		Type      loglevel `json:"Type"`
		Text      string   `json:"Text"`
	}

	var e logEvent
	err := json.Unmarshal(message, &e)
	if err != nil {
		Log.Warn("Cannot parse logEvent: %s", err)
	}

	return e.TimeInfo, e.Type.String(), e.Text
}

func parseControlData(message []byte) controldata {
	var cdata controldata
	err := json.Unmarshal(message, &cdata)
	if err != nil {
		Log.Warn("Cannot parse controlData: %s", err)
	}
	if cdata.RUNSEQ == "" {
		Log.Debugln("Sequence   running: false")
		cdata.SEQRUNNING = false
	} else {
		Log.Debugf("Sequence   running: true; sequence: %s", cdata.RUNSEQ)
		cdata.SEQRUNNING = true
	}
	if cdata.RUNDS == "" {
		Log.Debugln("Dragscript running: false")
		cdata.DRAGRUNNING = false
	} else {
		Log.Debugf("Dragscript running: true; dragscript: %s", cdata.RUNDS)
		cdata.DRAGRUNNING = true
	}
	cdata.CAMSTATUS = ccdstat(cdata.CCDSTAT).String()

	controlDataUpdated = true
	Log.Debugf("Mount status: %s", strconv.FormatBool(cdata.MNTPARK))
	return cdata
}

var lastpoll time.Time

func heartbeatVoyager(c *websocket.Conn, quit chan bool) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	lastpoll = time.Now()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			now := t
			elapsed := now.Sub(lastpoll)

			// manage heartbeat
			if elapsed.Seconds() > 10 {
				lastpoll = now
				secs := now.Unix()
				heartbeat := &event{
					Event:     "Polling",
					Timestamp: float64(secs),
					Inst:      1,
				}
				data, _ := json.Marshal(heartbeat)
				sendToVoyager(c, data)

				Log.Debugf("Heartbeat Sent")
			}
		case <-quit:
			Log.Debugf("Quit heartbeat loop!")
			err := c.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			)
			if err != nil {
				Log.Println("write close:", err)
				return
			}
			return
		case <-interrupt:
			Log.Debugf("Want interrupt!")
			// Close the read goroutine
			done <- true
			// Cleanly close the websocket connection by sending a close message
			err := c.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			)
			if err != nil {
				Log.Warn("write close:", err)
				return
			}
			Log.Println("Shutdown barbecue")

			os.Exit(0)
		}
	}
}
