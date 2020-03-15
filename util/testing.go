package util

import (
	"astrongo/core"
	"astrongo/eventlogger"
	"astrongo/messagedirector"
	"fmt"
	"reflect"
)

func StartDaemon(config core.ServerConfig) {
	core.Config = &config
	eventlogger.StartEventLogger()
	messagedirector.Start()
}

func StopDameon() {
	core.StopChan <- true
}

func ReloadConfig(config core.ServerConfig) {
	StopDameon()
	StartDaemon(config)
}

type TestDatagram struct {
	*DatagramIterator
}

func (d *TestDatagram) Data() []uint8 {
	return d.dg.Bytes()
}

func (d *TestDatagram) Payload() []byte {
	return d.dg.Bytes()[1+Dgsize_t(d.RecipientCount())*Chansize:]
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
}

func (d *TestDatagram) Equals(other *TestDatagram) bool {
	return reflect.DeepEqual(d.Data(), other.Data())
}

func (d *TestDatagram) Create(recipients []Channel_t, sender Channel_t, msgType uint16) {
	dg := NewDatagram()
	dg.AddMultipleServerHeader(recipients, sender, msgType)
	d.DatagramIterator = NewDatagramIterator(&dg)
}

func (d *TestDatagram) CreateControl() {
	dg := NewDatagram()
	dg.AddUint8(1)
	dg.AddChannel(CONTROL_MESSAGE)
	d.DatagramIterator = NewDatagramIterator(&dg)
}

func (d *TestDatagram) CreateAddChannel(ch Channel_t) {
	dg := NewDatagram()
	dg.AddControlHeader(CONTROL_ADD_CHANNEL)
	dg.AddChannel(ch)
	d.DatagramIterator = NewDatagramIterator(&dg)
}

func (d *TestDatagram) CreateRemoveChannel(ch Channel_t) {
	dg := NewDatagram()
	dg.AddControlHeader(CONTROL_REMOVE_CHANNEL)
	dg.AddChannel(ch)
	d.DatagramIterator = NewDatagramIterator(&dg)
}

func (d *TestDatagram) CreateAddRange(upper Channel_t, lower Channel_t) {
	dg := NewDatagram()
	dg.AddControlHeader(CONTROL_ADD_RANGE)
	dg.AddChannel(upper)
	dg.AddChannel(lower)
	d.DatagramIterator = NewDatagramIterator(&dg)
}

func (d *TestDatagram) CreateRemoveRange(upper Channel_t, lower Channel_t) {
	dg := NewDatagram()
	dg.AddControlHeader(CONTROL_REMOVE_RANGE)
	dg.AddChannel(upper)
	dg.AddChannel(lower)
	d.DatagramIterator = NewDatagramIterator(&dg)
}

func (d *TestDatagram) CreateAddPostRemove(sender Channel_t, data Datagram) {
	dg := NewDatagram()
	dg.AddControlHeader(CONTROL_ADD_POST_REMOVE)
	dg.AddChannel(sender)
	dg.AddBlob(&data)
	d.DatagramIterator = NewDatagramIterator(&dg)
}

func (d *TestDatagram) CreateSetConName(name string) {
	dg := NewDatagram()
	dg.AddControlHeader(CONTROL_SET_CON_NAME)
	dg.AddString(name)
	d.DatagramIterator = NewDatagramIterator(&dg)
}

func (d *TestDatagram) CreateSetConUrl(name string) {
	dg := NewDatagram()
	dg.AddControlHeader(CONTROL_SET_CON_URL)
	dg.AddString(name)
	d.DatagramIterator = NewDatagramIterator(&dg)
}
