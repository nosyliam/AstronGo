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

const LATENCY = time.Millisecond * 1000

var client *net.Client
var fakeParticipant *MDParticipantFake
var msgQueue chan Datagram

func timeoutWrapper(timeout func(), tick func() bool) {
	timeoutChan := time.After(2 * time.Second)
	tickChan := time.Tick(1 * time.Millisecond)

	for {
		select {
		case <-timeoutChan:
			timeout()
			return
		case <-tickChan:
			if tick() == true {
				return
			}
		}
	}
}

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

	timeoutWrapper(
		func() {
			t.Fatal("MD start failure")
		},
		func() bool {
			if len(MD.participants) != 0 {
				return true
			}
			return false
		})
}

func TestMD_ControlMessages(t *testing.T) {
	assert.Len(t, MD.participants, 1)
	dg := NewDatagram()
	dg.AddControlHeader(CONTROL_SET_CON_NAME)
	dg.AddString("client")
	client.SendDatagram(dg)

	timeoutWrapper(
		func() {
			t.Fatal("control send timeout")
		},
		func() bool {
			if MD.participants[0].Name() == "client" {
				return true
			}
			return false
		})
}

func TestMD_ControlSubscribe(t *testing.T) {
	assert.Len(t, MD.participants, 1)
	dg := NewDatagram()
	dg.AddControlHeader(CONTROL_ADD_CHANNEL)
	dg.AddChannel(50)
	client.SendDatagram(dg)

	timeoutWrapper(
		func() {
			t.Fatal("control subscribe timeout")
		},
		func() bool {
			if _, ok := channelMap.subscriptions.Load(Channel_t(50)); ok {
				return true
			}
			return false
		})
}

func TestMD_MessageRoute(t *testing.T) {
	assert.Len(t, MD.participants, 1)
	dg := NewDatagram()
	dg.AddUint8(1)
	dg.AddChannel(50)
	dg.AddChannel(60)
	dg.AddUint16(1337)
	MD.Queue <- struct {
		dg Datagram
		md MDParticipant
	}{dg, nil}
	dgRecv := <-msgQueue
	dgi := NewDatagramIterator(&dgRecv)
	assert.Equal(t, dgi.ReadChannel(), Channel_t(60))
	assert.Equal(t, dgi.ReadUint16(), uint16(1337))
}
