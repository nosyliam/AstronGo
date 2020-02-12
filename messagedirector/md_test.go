package messagedirector

import (
	"astrongo/core"
	"astrongo/net"
	. "astrongo/util"
	"github.com/stretchr/testify/require"
	gonet "net"
	"src/github.com/tj/assert"
	"testing"
	"time"
)

var client *net.Client
var client2 *net.Client
var fakeParticipant *MDParticipantFake
var fakeParticipant2 *MDParticipantFake2
var msgQueue chan Datagram
var msgQueue2 chan Datagram

type MDParticipantFake2 struct{ MDParticipant }

func (m *MDParticipantFake2) ReceiveDatagram(datagram Datagram) {
	msgQueue2 <- datagram
}

func (m *MDParticipantFake2) HandleDatagram(datagram Datagram, dgi *DatagramIterator) {
	msgQueue2 <- datagram
}

func (m *MDParticipantFake2) Terminate(error) {}

func testSubscribeRoutine(ch Channel_t, t *testing.T) {
	timeoutWrapper(
		func() {
			t.Fatal("control subscribe timeout")
		},
		func() bool {
			if _, ok := channelMap.subscriptions.Load(ch); ok {
				return true
			}
			return false
		})
}

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

func createClient(p net.DatagramHandler, addr string) (client *net.Client, err error) {
	conn, err := gonet.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	socket := net.NewSocketTransport(conn, time.Second*60, 4096)
	client = net.NewClient(socket, p)
	return client, nil
}

func init() {
	msgQueue = make(chan Datagram)
	msgQueue2 = make(chan Datagram)
}

func TestMD_Start(t *testing.T) {
	channelMap = &ChannelMap{}
	channelMap.init()
	core.Config = &core.ServerConfig{MessageDirector: struct {
		Bind    string
		Connect string
	}{Bind: "127.0.0.1:7199", Connect: ""}}

	go Start()
	fakeParticipant = &MDParticipantFake{}

	if client1, err := createClient(fakeParticipant, ":7199"); err != nil {
		t.Fatal(err)
	} else {
		client = client1
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

	testSubscribeRoutine(Channel_t(50), t)
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
	assert.Equal(t, Channel_t(60), dgi.ReadChannel())
	assert.Equal(t, uint16(1337), dgi.ReadUint16())
}

func TestMD_ControlUnsubscribe(t *testing.T) {
	assert.Len(t, MD.participants, 1)
	dg := NewDatagram()
	dg.AddControlHeader(CONTROL_REMOVE_CHANNEL)
	dg.AddChannel(50)
	client.SendDatagram(dg)

	timeoutWrapper(
		func() {
			t.Fatal("control unsubscribe timeout")
		},
		func() bool {
			if _, ok := channelMap.subscriptions.Load(Channel_t(50)); !ok {
				return true
			}
			return false
		})
}

func TestMD_ControlSubscribeRange(t *testing.T) {
	assert.Len(t, MD.participants, 1)
	dg := NewDatagram()
	dg.AddControlHeader(CONTROL_ADD_RANGE)
	dg.AddChannel(1000)
	dg.AddChannel(2000)
	client.SendDatagram(dg)
	time.Sleep(50 * time.Millisecond)

	dg = NewDatagram()
	dg.AddServerHeader(1500, 70, 1337)
	MD.Queue <- struct {
		dg Datagram
		md MDParticipant
	}{dg, nil}
	<-msgQueue
}

func TestMD_ControlUnsubscribeRange(t *testing.T) {
	assert.Len(t, MD.participants, 1)
	dg := NewDatagram()
	dg.AddControlHeader(CONTROL_REMOVE_RANGE)
	dg.AddChannel(1400)
	dg.AddChannel(1600)
	client.SendDatagram(dg)
	timeoutWrapper(
		func() {
			t.Fatal("control unsubscribe range timeout")
		},
		func() bool {
			// Channelmap should splice two (three, in practice) new ranges
			if len(channelMap.ranges.intervals) > 1 {
				return true
			}
			return false
		})

	dg = NewDatagram()
	dg.AddServerHeader(1500, 80, 1337)
	MD.Queue <- struct {
		dg Datagram
		md MDParticipant
	}{dg, nil}
	time.Sleep(10 * time.Millisecond)
	require.Empty(t, msgQueue)
}

func TestMD_PostRemove(t *testing.T) {
	var recvDg *Datagram
	fakeParticipant2 = &MDParticipantFake2{}

	go func() {
		if client, err := createClient(fakeParticipant2, ":7199"); err != nil {
			t.Fatal(err)
		} else {
			client2 = client
		}
	}()
	timeoutWrapper(
		func() {
			t.Fatal("control multiple client connect timeout")
		},
		func() bool {
			if client2 != nil {
				return true
			}
			return false
		})

	time.Sleep(200 * time.Millisecond)

	dg := NewDatagram()
	dg.AddControlHeader(CONTROL_ADD_CHANNEL)
	dg.AddChannel(10000)
	client2.SendDatagram(dg)
	time.Sleep(100 * time.Millisecond)

	postRemove := NewDatagram()
	postRemove.AddServerHeader(10000, 1, 0)
	postRemove.AddUint32(0xDEADBEEF)

	dg = NewDatagram()
	dg.AddControlHeader(CONTROL_ADD_POST_REMOVE)
	dg.AddChannel(10000)
	dg.AddBlob(&postRemove)
	client.SendDatagram(dg)

	time.Sleep(10 * time.Millisecond)
	client.Close()
	time.Sleep(10 * time.Millisecond)
	go func() {
		dg = <-msgQueue2
		recvDg = &dg
	}()
	timeoutWrapper(
		func() {
			t.Fatal("post-remove timeout")
		},
		func() bool {
			if recvDg != nil {
				return true
			}
			return false
		})
	dgi := NewDatagramIterator(&dg)
	dgi.ReadChannel() // Sender
	dgi.ReadUint16()  // Message type
	assert.Equal(t, uint32(0xDEADBEEF), dgi.ReadUint32())
}
