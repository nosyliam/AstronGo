package test

import (
	"astrongo/core"
	"astrongo/eventlogger"
	"astrongo/net"
	. "astrongo/util"
	"encoding/hex"
	"fmt"
	gonet "net"
	"reflect"
	"strings"
	"testing"
	"time"
)

type LogSilencer struct {
}

type UpstreamHandler struct {
	Server *gonet.Conn
}

func (u *UpstreamHandler) HandleConnect(conn gonet.Conn) {
	u.Server = &conn
}

func StartDaemon(config core.ServerConfig) {
	core.Config = &config
	eventlogger.StartEventLogger()

}

func StopDameon() {
	core.StopChan <- true
}

func ReloadConfig(config core.ServerConfig) {
	StopDameon()
	StartDaemon(config)
}

func StartUpstream(bindAddr string) *UpstreamHandler {
	server := &net.NetworkServer{}
	handler := &UpstreamHandler{}
	server.Handler = handler
	errChan := make(chan error)
	go func() {
		err := <-errChan
		switch err {
		case nil:
		default:
			panic(fmt.Sprintf("Failed to open upstream: %s", err.Error()))
		}
	}()
	go server.Start(bindAddr, errChan)
	return handler
}

// Utility class for managing Datagrams in a test environment
type TestDatagram struct {
	*DatagramIterator
}

func (d *TestDatagram) Set(dg *Datagram) *TestDatagram {
	d.DatagramIterator = NewDatagramIterator(dg)
	return d
}

func (d *TestDatagram) Data() []uint8 {
	return d.Dg.Bytes()
}

func (d *TestDatagram) Payload() []byte {
	return d.Dg.Bytes()[1+Dgsize_t(d.RecipientCount())*Chansize:]
}

func (d *TestDatagram) Channels() []Channel_t {
	var channels []Channel_t
	d.Seek(1)
	for n := 0; n < int(d.RecipientCount()); n++ {
		channels = append(channels, d.ReadChannel())
	}
	return channels
}

func (d *TestDatagram) Matches(other *TestDatagram) bool {
	return reflect.DeepEqual(d.Payload(), other.Payload()) && reflect.DeepEqual(d.Channels(), other.Channels())
}

func (d *TestDatagram) MatchesHeader(recipients []Channel_t, sender Channel_t, msgType uint16, payloadSz int) (result bool, why string) {
	d.Seek(0)
	if !reflect.DeepEqual(d.Channels(), recipients) {
		return false, "Recipients do not match"
	}

	if d.Sender() != sender {
		return false, fmt.Sprintf("Sender doesn't match, %d = %d (expected, actual)", sender, d.Sender())
	}

	if d.MessageType() != msgType {
		return false, fmt.Sprintf("Message type doesn't match, %d = %d (expected, actual)", msgType, d.MessageType())
	}

	if payloadSz != -1 && len(d.Payload()) != payloadSz {
		return false, fmt.Sprintf("Payload size is %d; expecting %d", len(d.Payload()), payloadSz)
	}

	return true, ""
}

func (d *TestDatagram) Equals(other *TestDatagram) bool {
	return reflect.DeepEqual(d.Data(), other.Data())
}

func (d *TestDatagram) AssertEquals(other *TestDatagram, t *testing.T, client bool) {
	d.Seek(0)
	other.Seek(0)

	errorComp := func() {
		err := "Datagram assertion failed: payload" +
			"\n---EXPECTED DATAGRAM---\n%sRecipients=%d, Sender=%d, Message type=%d\n" +
			"\n---RECEIVED DATAGRAM---\n%sRecipients=%d, Sender=%d, Message type=%d"
		t.Errorf(err, hex.Dump(d.Dg.Bytes()), d.RecipientCount(), d.Sender(), d.MessageType(),
			hex.Dump(other.Dg.Bytes()), other.RecipientCount(), other.Sender(), other.MessageType())
		panic("")
	}

	if client {
		if msgTypeExpected, msgTypeReceived := d.ReadUint16(), other.ReadUint16(); msgTypeExpected != msgTypeReceived {
			t.Errorf("Datagram assertion failed: msg type, %d != %d", msgTypeExpected, msgTypeReceived)
			return
		}

		if !reflect.DeepEqual(other.ReadRemainder(), d.ReadRemainder()) {
			errorComp()
			return
		}
	} else {
		if channelsExpected, channelsReceived := d.ReadUint8(), other.ReadUint8(); channelsExpected != channelsReceived {
			t.Errorf("Datagram assertion failed: channels expected, %d != %d", channelsExpected, channelsReceived)
			//return
		}

		expectedRecipients, recievedRecipients := d.Channels(), other.Channels()
		if !reflect.DeepEqual(expectedRecipients, recievedRecipients) {
			t.Errorf("Datagram assertion failed: recipients expected, %s != %s", expectedRecipients, recievedRecipients)
			return
		}

		if expectedRecipients[0] != CONTROL_MESSAGE {
			if expectedSender, receivedSender := d.ReadChannel(), other.ReadChannel(); expectedSender != receivedSender {
				t.Errorf("Datagram assertion failed: sender, %d != %d", expectedSender, receivedSender)
			}
		}

		if msgTypeExpected, msgTypeReceived := d.ReadUint16(), other.ReadUint16(); msgTypeExpected != msgTypeReceived {
			t.Errorf("Datagram assertion failed: msg type, %d != %d", msgTypeExpected, msgTypeReceived)
		}

		if !reflect.DeepEqual(other.ReadRemainder(), d.ReadRemainder()) {
			errorComp()
			return
		}
	}
}

func (d *TestDatagram) Create(recipients []Channel_t, sender Channel_t, msgType uint16) *Datagram {
	dg := NewDatagram()
	dg.AddMultipleServerHeader(recipients, sender, msgType)
	d.DatagramIterator = NewDatagramIterator(&dg)
	return d.Dg
}

func (d *TestDatagram) CreateControl() *Datagram {
	dg := NewDatagram()
	dg.AddUint8(1)
	dg.AddChannel(CONTROL_MESSAGE)
	d.DatagramIterator = NewDatagramIterator(&dg)
	return d.Dg
}

func (d *TestDatagram) CreateAddChannel(ch Channel_t) *Datagram {
	dg := NewDatagram()
	dg.AddControlHeader(CONTROL_ADD_CHANNEL)
	dg.AddChannel(ch)
	d.DatagramIterator = NewDatagramIterator(&dg)
	return d.Dg
}

func (d *TestDatagram) CreateRemoveChannel(ch Channel_t) *Datagram {
	dg := NewDatagram()
	dg.AddControlHeader(CONTROL_REMOVE_CHANNEL)
	dg.AddChannel(ch)
	d.DatagramIterator = NewDatagramIterator(&dg)
	return d.Dg
}

func (d *TestDatagram) CreateAddRange(upper Channel_t, lower Channel_t) *Datagram {
	dg := NewDatagram()
	dg.AddControlHeader(CONTROL_ADD_RANGE)
	dg.AddChannel(upper)
	dg.AddChannel(lower)
	d.DatagramIterator = NewDatagramIterator(&dg)
	return d.Dg
}

func (d *TestDatagram) CreateRemoveRange(upper Channel_t, lower Channel_t) *Datagram {
	dg := NewDatagram()
	dg.AddControlHeader(CONTROL_REMOVE_RANGE)
	dg.AddChannel(upper)
	dg.AddChannel(lower)
	d.DatagramIterator = NewDatagramIterator(&dg)
	return d.Dg
}

func (d *TestDatagram) CreateAddPostRemove(sender Channel_t, data Datagram) *Datagram {
	dg := NewDatagram()
	dg.AddControlHeader(CONTROL_ADD_POST_REMOVE)
	dg.AddChannel(sender)
	dg.AddBlob(&data)
	d.DatagramIterator = NewDatagramIterator(&dg)
	return d.Dg
}

func (d *TestDatagram) CreateClearPostRemove(sender Channel_t) *Datagram {
	dg := NewDatagram()
	dg.AddControlHeader(CONTROL_CLEAR_POST_REMOVES)
	dg.AddChannel(sender)
	d.DatagramIterator = NewDatagramIterator(&dg)
	return d.Dg
}

func (d *TestDatagram) CreateSetConName(name string) *Datagram {
	dg := NewDatagram()
	dg.AddControlHeader(CONTROL_SET_CON_NAME)
	dg.AddString(name)
	d.DatagramIterator = NewDatagramIterator(&dg)
	return d.Dg
}

func (d *TestDatagram) CreateSetConUrl(name string) *Datagram {
	dg := NewDatagram()
	dg.AddControlHeader(CONTROL_SET_CON_URL)
	dg.AddString(name)
	d.DatagramIterator = NewDatagramIterator(&dg)
	return d.Dg
}

// Utility class for managing MD connections in a test environment
type TestMDConnection struct {
	*net.Client
	messages chan Datagram
	name     string
	timeout  int
}

func (c *TestMDConnection) Set(conn gonet.Conn, name string) *TestMDConnection {
	c.timeout = 201
	c.messages = make(chan Datagram, 200)
	c.name = name
	socket := net.NewSocketTransport(conn, 60*time.Second, 4096)
	c.Client = net.NewClient(socket, c, 200*time.Millisecond)
	return c
}

func (c *TestMDConnection) Connect(addr string, name string) *TestMDConnection {
	c.timeout = 201
	c.messages = make(chan Datagram, 200)
	c.name = name
	conn, err := gonet.Dial("tcp", addr)
	if err != nil {
		panic(fmt.Sprintf("Testing client failed to connect to %d: %s", addr, err))
	}

	socket := net.NewSocketTransport(conn, 60*time.Second, 4096)
	c.Client = net.NewClient(socket, c, 200*time.Millisecond)
	return c
}

func (c *TestMDConnection) HandleDatagram(dg Datagram, dgi *DatagramIterator) {
	c.messages <- dg
}

func (c *TestMDConnection) ReceiveDatagram(dg Datagram) {
	c.messages <- dg
}

func (c *TestMDConnection) Terminate(err error) { /* not needed */ }

func (c *TestMDConnection) Receive() *Datagram {
	select {
	case dg := <-c.messages:
		return &dg
	case <-time.After(time.Duration(c.timeout) * time.Millisecond):
		panic("No message received!")
	}
}

func (c *TestMDConnection) Expect(t *testing.T, dg Datagram, client bool) {
	recv := c.ReceiveMaybe()
	if recv == nil {
		t.Errorf("No datagram received for connection %s", c.name)
		panic("")
		return
	}
	(&TestDatagram{}).Set(&dg).AssertEquals((&TestDatagram{}).Set(recv), t, client)
}

func (c *TestMDConnection) ExpectMany(t *testing.T, datagrams []Datagram, client bool, allowTime bool) {
	var recvs []Datagram
	if allowTime {
		c.timeout = 1000
	}
	received, matched, expected := 0, 0, len(datagrams)
	for len(datagrams) != matched {
		recv := c.ReceiveMaybe()
		if recv == nil {
			if matched == 0 {
				t.Errorf("Expected %d datagrams, but received nothing", expected)
			} else {
				var msgTypes []string
				for _, dg := range recvs {
					msgTypes = append(msgTypes, fmt.Sprintf("%d", NewDatagramIterator(&dg).MessageType()))
				}
				t.Errorf("Recieved %d datagrams, of which %d matched, but expected %d\n"+
					"Received message types: %s", received, matched, expected, strings.Join(msgTypes, ", "))
			}
			for n, dg := range recvs {
				fmt.Printf("Datagram #%d:\n%s\n", n, hex.Dump(dg.Bytes()))
			}
			break
		} else {
			received++
		}

		testRecv := (&TestDatagram{}).Set(recv)
		found := false
		for _, dg := range datagrams {
			testDg := (&TestDatagram{}).Set(&dg)
			if (client && testRecv.Equals(testDg)) || testRecv.Matches(testDg) {
				recvs = append(recvs, dg)
				found = true
			}
		}

		if found {
			matched++
		}
	}

	c.timeout = 201
}

func (c *TestMDConnection) ExpectNone(t *testing.T) {
	recv := c.ReceiveMaybe()
	if recv != nil {
		t.Errorf("Received unexpected datagram:\n%s", hex.Dump(recv.Bytes()))
		panic("")
	}
}

func (c *TestMDConnection) ReceiveMaybe() *Datagram {
	select {
	case dg := <-c.messages:
		return &dg
	case <-time.After(time.Duration(c.timeout) * time.Millisecond):
		return nil
	}
}

func (c *TestMDConnection) Flush() {
	for len(c.messages) > 0 {
		<-c.messages
	}
}

// Utility class for addressing multiple channels in a test environment
type TestChannelConnection struct {
	TestMDConnection
	channels map[Channel_t]bool
}

func (c *TestChannelConnection) Create(addr string, name string, ch Channel_t) {
	c.TestMDConnection.Connect(addr, name)
	c.channels = make(map[Channel_t]bool)

	if ch != 0 {
		c.channels[ch] = true
		c.SendDatagram(*(&TestDatagram{}).CreateAddChannel(ch))
	}
}

func (c *TestChannelConnection) AddChannel(ch Channel_t) {
	if _, ok := c.channels[ch]; !ok {
		c.channels[ch] = true
		c.SendDatagram(*(&TestDatagram{}).CreateAddChannel(ch))
	}
}

func (c *TestChannelConnection) RemoveChannel(ch Channel_t) {
	if _, ok := c.channels[ch]; !ok {
		delete(c.channels, ch)
		c.SendDatagram(*(&TestDatagram{}).CreateRemoveChannel(ch))
	}
}

func (c *TestChannelConnection) ClearChannels() {
	for ch, _ := range c.channels {
		delete(c.channels, ch)
		c.SendDatagram(*(&TestDatagram{}).CreateRemoveChannel(ch))
	}
}

func (c *TestChannelConnection) Close() {
	c.ClearChannels()
	c.Close()
}
