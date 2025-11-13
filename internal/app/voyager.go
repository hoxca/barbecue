/*
Package app implement barbecue application logic
*/
package app

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	Log "github.com/apatters/go-conlog"
	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/websocket"
)

func ConnectVoyager(addr *string) (*websocket.Conn, error) {
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

func RemoteSetDashboard(sc *SafeConnection) {
	p := &Params{
		UID:  fmt.Sprintf("%s", uuid.Must(uuid.NewV4())),
		IsOn: true,
	}

	setDashboard := &Method{
		Method: "RemoteSetDashboardMode",
		Params: *p,
		ID:     1,
	}

	data, _ := json.Marshal(setDashboard)
	SendToVoyager(sc, data)
}

func SendToVoyager(sc *SafeConnection, data []byte) {
	message := fmt.Appendf(nil, "%s\r\n", data)
	err := sc.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		Log.Println("write:", err)
		return
	}
	Log.Debugf("send: %s", data)
	time.Sleep(1 * time.Second)
}

func VoyagerStatusDebug() {
	if ControlDataUpdated {
		Log.Info("Voyager Status:")
		Log.Infof("  Voyager    status: %d", VoyagerStatus.VOYSTAT)
		Log.Infof("  Camera     status: %s", Ccdstat(VoyagerStatus.CCDSTAT).String())
		Log.Infof("  Camera    cooling: %s", strconv.FormatBool(VoyagerStatus.CCDCOOL))
		Log.Infof("  Camera   ccd temp: %f", VoyagerStatus.CCDTEMP)
		Log.Infof("  Camera  ccd power: %d", VoyagerStatus.CCDPOW)
		Log.Infof("  Focuser      temp: %f\n", VoyagerStatus.AFTEMP)
	}
}

func RecvFromVoyager(sc *SafeConnection, done chan bool) {
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
				//              if !ControlDataUpdated {
				Log.Debugf("recv msg: %s", strings.TrimRight(msg, "\r\n"))
				VoyagerStatus = ParseControlData(message)
			//      }
			case strings.Contains(msg, `"Event":"LogEvent"`):
				ts, level, logline := ParseLogEvent(message)
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

func ParseLogEvent(message []byte) (float64, string, string) {
	type logEvent struct {
		Event     string   `json:"Event"`
		Timestamp float64  `json:"Timestamp"`
		Host      string   `json:"Host"`
		Inst      int      `json:"Inst"`
		TimeInfo  float64  `json:"TimeInfo"`
		Type      Loglevel `json:"Type"`
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

func ParseControlData(message []byte) Controldata {
	var cdata Controldata
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
	cdata.CAMSTATUS = Ccdstat(cdata.CCDSTAT).String()

	ControlDataUpdated = true
	Log.Debugf("Mount status: %s", strconv.FormatBool(cdata.MNTPARK))
	return cdata
}

// HeartbeatVoyager Perform reccurent heartbeat signal to voyager to maintain the websocket alive.
func HeartbeatVoyager(sc *SafeConnection, quit chan bool, ticker *time.Ticker) {
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

// GetCameraStatusWithConnection establishes connection to Voyager and retrieves camera status.
func GetCameraStatusWithConnection() (Camstatus, error) {
	SetUpLogs()
	ParseConfig()

	c, err := ConnectVoyager(&AddrFlag)
	if err != nil {
		return Camstatus{}, fmt.Errorf("failed to connect to Voyager: %w", err)
	}
	defer c.Close()

	sc := NewSafeConnection(c)
	defer sc.Close()

	Quit = make(chan bool)
	Done = make(chan bool)

	go RecvFromVoyager(sc, Done)
	RemoteSetDashboard(sc)

	ticker := time.NewTicker(45 * time.Second)
	defer ticker.Stop()
	go HeartbeatVoyager(sc, Quit, ticker)
	time.Sleep(1 * time.Second)

	VoyagerStatusDebug()

	camera := RetrieveCameraStatus()

	Done <- true

	return camera, nil
}
