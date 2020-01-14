package eventlogger

import (
	"astrongo/core"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
	"time"
)

func testLog(t *testing.T, key string, val string) {
	logfile.Seek(0, 0)
	buff := make([]byte, 1024)
	n, _ := logfile.Read(buff)
	out := make(map[string]interface{})
	json.Unmarshal(buff[0:n], &out)
	require.Equal(t, val, out[key])
	logfile.Truncate(0)
	logfile.Seek(0, 0)
}

func TestStartEventLogger(t *testing.T) {
	StartEventLogger()
	if server == nil {
		t.Fatal("could not start server")
	}
	logfile.Truncate(0)
	logfile.Seek(0, 0)
}

func TestEventLogger_Process(t *testing.T) {
	addr := &net.UDPAddr{IP: []byte{0, 0, 0, 0}, Port: 10001, Zone: ""}
	processPacket([]byte("if the ev reads this, the test fails"), addr)
	processPacket([]byte("\x82\xa3bar\xa3baz\xa4type\xa3foo"), addr)
	time.Sleep(time.Millisecond * 100)
	testLog(t, "bar", "baz")
}

func TestEventLogger_Listen(t *testing.T) {
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: []byte{0, 0, 0, 0}, Port: 7197, Zone: ""})
	if err != nil {
		t.Fatalf("failed to connect to eventlogger: %s", err)
	}

	conn.Write([]byte("if the ev reads this, the test fails"))
	conn.Write([]byte("\x82\xa3bar\xa3baz\xa4type\xa3foo"))
	time.Sleep(time.Millisecond * 100)
	testLog(t, "bar", "baz")

}

func init() {
	core.Config = &core.ServerConfig{}
}
