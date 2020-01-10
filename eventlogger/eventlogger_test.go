package eventlogger

import (
	"astrongo/core"
	"net"
	"testing"
)

func TestStartEventLogger(t *testing.T) {
	StartEventLogger()
	if server == nil {
		t.Fatal("could not start server")
	}
}

func TestEventLogger_Process(t *testing.T) {
	addr := &net.UDPAddr{IP: []byte{0, 0, 0, 0}, Port: 10001, Zone: ""}
	processPacket([]byte("\x82\xa3bar\xa3baz\xa4type\xa3foo"), addr)
}

func init() {
	core.Config = &core.ServerConfig{}
}
