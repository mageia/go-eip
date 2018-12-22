package go_eip

import (
	"io"
	"log"
	"net"
	"sync"
	"time"
)

const (
	tcpTimeout     = 10 * time.Second
	tcpIdleTimeout = 60 * time.Second
	tcpMaxLength   = 4096
	isoTCP         = 44818

	connectionTypeBasic = 3
)

type tcpPackager struct {
}
type tcpTransporter struct {
	Address      string
	Timeout      time.Duration
	IdleTimeout  time.Duration
	Logger       *log.Logger
	mu           sync.Mutex
	conn         net.Conn
	closeTimer   *time.Timer
	lastActivity time.Time
}

type TCPClientHandler struct {
	tcpPackager
	tcpTransporter
}

func NewTCPClientHandler(address string) *TCPClientHandler {
	return nil
}
func TCPClient(address string) Client {
	handler := NewTCPClientHandler(address)
	return NewClient(handler)
}
func (t *tcpTransporter) Send(request []byte, response []byte) (err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastActivity = time.Now()
	t.startCloseTimer()

	var timeout time.Time
	if t.Timeout > 0 {
		timeout = t.lastActivity.Add(t.Timeout)
	}
	if err = t.conn.SetDeadline(timeout); err != nil {
		return
	}

	t.logf("logix: sending %x", request)
	if _, err = t.conn.Write(request); err != nil {
		return
	}
	done := false
	data := make([]byte, tcpMaxLength)
	//length := 0
	for !done && err == nil {
		if _, err = io.ReadFull(t.conn, data); err != nil {
			return
		}
		done = true
	}
	return
}

func (t *tcpTransporter) Connect() (err error) {
	return
}
func (t *tcpTransporter) tcpConnect() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.conn == nil {
		dialer := net.Dialer{Timeout: t.Timeout}
		conn, err := dialer.Dial("tcp", t.Address)
		if err != nil {
			return err
		}
		t.conn = conn
	}
	return nil
}
func (t *tcpTransporter) connect() error {
	//first stage: TCP connection
	err := t.tcpConnect()
	if err != nil {
		return err
	}
	//second stage: ISOTCP (ISO 8073) Connection
	//err = t.isoConnect()
	//if err != nil {
	//	return err
	//}
	// Third stage : S7 protocol data unit negotiation
	//return t.negotiatePduLength()
	return nil
}
func (t *tcpTransporter) startCloseTimer() {
	if t.IdleTimeout <= 0 {
		return
	}

	if t.closeTimer == nil {
		t.closeTimer = time.AfterFunc(t.IdleTimeout, t.closeIdle)
	} else {
		t.closeTimer.Reset(t.IdleTimeout)
	}
}
func (t *tcpTransporter) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.close()
}
func (t *tcpTransporter) logf(format string, v ...interface{}) {
	if t.Logger != nil {
		t.Logger.Printf(format, v...)
	}
}
func (t *tcpTransporter) close() (err error) {
	if t.conn != nil {
		err = t.conn.Close()
		t.conn = nil
	}
	return
}
func (t *tcpTransporter) closeIdle() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.IdleTimeout <= 0 {
		return
	}
	idle := time.Now().Sub(t.lastActivity)
	if idle >= t.IdleTimeout {
		t.logf("logix: closing connection due to idle timeout: %v", idle)
		t.close()
	}
}

func (t *tcpPackager) Verify(request []byte, response []byte) (err error) {
	return
}
