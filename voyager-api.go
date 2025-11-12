package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"sync"
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

// SafeConnection provides thread-safe access to WebSocket connection.
type SafeConnection struct {
	conn   *websocket.Conn
	mu     sync.RWMutex
	closed bool
}

// NewSafeConnection creates a new safe connection wrapper.
func NewSafeConnection(conn *websocket.Conn) *SafeConnection {
	return &SafeConnection{
		conn:   conn,
		closed: false,
	}
}

// WriteMessage safely writes to the connection if not closed.
func (sc *SafeConnection) WriteMessage(messageType int, data []byte) error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if sc.closed {
		return errors.New("connection is closed")
	}

	if sc.conn == nil {
		return errors.New("no underlying connection")
	}

	return sc.conn.WriteMessage(messageType, data)
}

// Close safely closes the connection and marks it as closed.
func (sc *SafeConnection) Close() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.closed {
		return nil // Already closed
	}

	sc.closed = true
	if sc.conn != nil {
		return sc.conn.Close()
	}
	return nil // No connection to close
}

// IsClosed returns whether the connection is closed.
func (sc *SafeConnection) IsClosed() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.closed
}

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
	c, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		Log.Printf("Can't connect, verify Voyager address or tcp port in the Voyager configuration\n")
		return c, err
	}
	defer resp.Body.Close()
	return c, err
}

func remoteSetDashboard(sc *SafeConnection) {
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
	sendToVoyager(sc, data)
}

func sendToVoyager(sc *SafeConnection, data []byte) {
	message := fmt.Appendf(nil, "%s\r\n", data)
	err := sc.WriteMessage(websocket.TextMessage, message)
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

func recvFromVoyager(sc *SafeConnection, done chan bool) {
	for {
		select {
		case <-done:
			Log.Debugf("Quit recv loop!")
			return
		default:
			_, message, err := sc.conn.ReadMessage()
			if err != nil {
				Log.Warn("read:", err)
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
		return 0, "", ""
	}

	// Check if Type is within valid range to avoid panic
	if e.Type < 1 || e.Type > 9 {
		return e.TimeInfo, "", e.Text
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

func heartbeatVoyager(sc *SafeConnection, quit chan bool) {
	ticker := time.NewTicker(45 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Send heartbeat
			if err := sc.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				Log.Error("Failed to send heartbeat: %v", err)
				return
			}
		case <-quit:
			// Received signal to quit
			Log.Info("Heartbeat goroutine exiting...")
			return
		}
	}
}
