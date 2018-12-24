package test

import (
	"go-eip"
	"log"
	"os"
	"testing"
	"time"
)

const (
	tcpDevice = "192.168.1.164"
)

func TestTCPClient(t *testing.T) {
	handler := go_eip.NewTCPClientHandler(tcpDevice, 0)
	handler.Timeout = 5 * time.Second
	handler.IdleTimeout = 180 * time.Second
	handler.Logger = log.New(os.Stdout, "tcp", log.LstdFlags)
	handler.Connect()
	defer handler.Close()

	client := go_eip.NewClient(handler)
	ClientTestAll(t, client)
}
