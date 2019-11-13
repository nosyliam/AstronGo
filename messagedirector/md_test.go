package messagedirector

import (
	"astrongo/core"
	"astrongo/net"
	. "astrongo/util"
	gonet "net"
	"src/github.com/tj/assert"
	"testing"
	"time"
)

var client *net.Client
var fakeParticipant *MDParticipantFake
var msgQueue chan Datagram

func createClient() error {
	conn, err := gonet.Dial("tcp", ":7199")
	if err != nil {
		return err
	}

	fakeParticipant = &MDParticipantFake{}
	socket := net.NewSocketTransport(conn, time.Second*60)
	client = net.NewClient(socket, fakeParticipant)
	return nil
}

func init() {
	core.Config = &core.ServerConfig{MessageDirector: struct{ Bind string }{Bind: ""}}
	msgQueue = make(chan Datagram)
}

func TestMD_Start(t *testing.T) {
	go Start()
	if err := createClient(); err != nil {
		t.Fatal(err)
	}
	for {
		if len(MD.participants) != 0 {
			break
		}
	}
}

func TestMD_ControlMessages(t *testing.T) {
	assert.Len(t, MD.participants, 1)
	dg := NewDatagram()
	dg.AddControlHeader(CONTROL_SET_CON_NAME)
	dg.AddString("client")
	client.SendDatagram(dg)
	for {
		if MD.participants[0].Name() == "client" {
			break
		}
	}
}
