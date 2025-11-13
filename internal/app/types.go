package app

import (
	"errors"
	"sync"
)

type Event struct {
	Event     string  `json:"Event"`
	Timestamp float64 `json:"Timestamp"`
	Host      string  `json:"Host,omitempty"`
	Inst      int     `json:"Inst"`
}

type Controldata struct {
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

// WebSocketConn interface for WebSocket operations.
type WebSocketConn interface {
	WriteMessage(messageType int, data []byte) error
	ReadMessage() (messageType int, p []byte, err error)
	Close() error
}

// SafeConnection provides thread-safe access to WebSocket connection.
type SafeConnection struct {
	conn   WebSocketConn
	mu     sync.RWMutex
	closed bool
}

// NewSafeConnection creates a new safe connection wrapper.
func NewSafeConnection(conn WebSocketConn) *SafeConnection {
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

type Method struct {
	Method string `json:"method"`
	Params Params `json:"params"`
	ID     int    `json:"id"`
}

type Params struct {
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

type Loglevel int

func (l Loglevel) String() string {
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

type Ccdstat int

func (cmos Ccdstat) String() string {
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

type Camstatus struct {
	Ambient int
	Temp    int
	Power   string
	Status  string
}

var (
	AddrFlag           string
	VerbosityFlag      string
	Quit               chan bool
	Done               chan bool
	VoyagerStatus      Controldata
	ControlDataUpdated = false
)
