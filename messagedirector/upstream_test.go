package messagedirector

import (
	"astrongo/core"
	"astrongo/net"
	. "astrongo/util"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	gonet "net"
	"src/github.com/tj/assert"
	"testing"
)

var downstream net.NetworkServer
var dsConn gonet.Conn
var recv chan []byte

type FakeDownstream struct {
	net.Server
}

func (s FakeDownstream) HandleConnect(conn gonet.Conn) {
	dsConn = conn
	go func() {
		for {
			buf, err := ioutil.ReadAll(conn)
			if err != nil {
				return
			} else {
				recv <- buf
			}
		}
	}()
}

func TestMDUpstream_Start(t *testing.T) {
	if MD != nil {
		MD.Shutdown()
	}

	downstream.Handler = FakeDownstream{}
	core.Config = &core.ServerConfig{MessageDirector: struct {
		Bind    string
		Connect string
	}{Bind: "127.0.0.1:7200", Connect: "127.0.0.1"}}

	recv = make(chan []byte)
	errChan := make(chan error)
	go Start()
	go downstream.Start(":7199", errChan)
	<-errChan

	fakeParticipant = &MDParticipantFake{}
	if client1, err := createClient(fakeParticipant, ":7200"); err != nil {
		t.Fatal(err)
	} else {
		client = client1
	}

	timeoutWrapper(
		func() {
			t.Fatal("downstream client connect timeout")
		},
		func() bool {
			if dsConn != nil {
				return true
			}
			return false
		})
}

func TestMDUpstream_ReceiveDatagram(t *testing.T) {
	socket := net.NewSocketTransport(dsConn, 0)

	// Client subscribe to channel 1000
	dg := NewDatagram()
	dg.AddControlHeader(CONTROL_ADD_CHANNEL)
	dg.AddChannel(1000)
	client.SendDatagram(dg)
	testSubscribeRoutine(Channel_t(1000), t)

	// Downstream sends update to channel 100
	dg = NewDatagram()
	dg.AddServerHeader(1000, 0, 100)
	dg.AddUint32(0xDEADBEEF)
	// Create a datagram with a size header. too lazy to manually calculate it.
	dgs := NewDatagram()
	dgs.AddSize(Dgsize_t(dg.Len()))
	dgs.Write(dg.Bytes())
	if _, err := socket.WriteDatagram(dgs); err != nil {
		t.Fatal(err)
	}
	assert.Empty(t, msgQueue)
	<-socket.Flush()

	// Client receives update
	dg = <-msgQueue
	dgi := NewDatagramIterator(&dg)
	dgi.ReadUint32() // receiver
	dgi.ReadUint32() // ???
	dgi.ReadUint16() // msgtype
	require.Equal(t, uint32(0xDEADBEEF), dgi.ReadUint32())
}

func init() {
	msgQueue = make(chan Datagram)
	downstream.Handler = FakeDownstream{}
}
