package go_eip

import (
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	tcpTimeout     = 10 * time.Second
	tcpIdleTimeout = 60 * time.Second
	tcpMaxLength   = 1024
	isoTCP         = 44818

	connectionTypeBasic = 3
)

type tcpPackager struct{}
type tcpTransporter struct {
	Address     string
	Timeout     time.Duration
	IdleTimeout time.Duration
	Logger      *log.Logger

	mu           sync.Mutex
	conn         net.Conn
	closeTimer   *time.Timer
	lastActivity time.Time
}
type tCPClientHandler struct {
	tcpPackager
	tcpTransporter
}

func NewTCPClientHandler(address string) *tCPClientHandler {
	h := &tCPClientHandler{}
	h.Address = address
	if !strings.Contains(address, ":") {
		h.Address = address + ":" + strconv.Itoa(isoTCP)
	}
	h.Timeout = tcpTimeout
	h.IdleTimeout = tcpIdleTimeout
	return h
}

func (t *tcpPackager) Verify(request []byte, response []byte) (err error) {
	return
}

func (t *tcpTransporter) Send(request []byte) ([]byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastActivity = time.Now()
	t.startCloseTimer()

	var timeout time.Time
	if t.Timeout > 0 {
		timeout = t.lastActivity.Add(t.Timeout)
	}
	if err := t.conn.SetDeadline(timeout); err != nil {
		return nil, err
	}

	//t.logf("logix: sending %02x", request)
	if _, err := t.conn.Write(request); err != nil {
		return nil, err
	}
	data := make([]byte, tcpMaxLength)

	l, e := t.conn.Read(data)
	if e != nil {
		return nil, e
	}
	return data[:l], nil
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
func (t *tcpTransporter) Connect() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.conn == nil {
		dialer := net.Dialer{Timeout: t.Timeout}
		conn, err := dialer.Dial("tcp", t.Address)
		if err != nil {
			t.logf("%v", err)
			return err
		}
		t.conn = conn
	}
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
