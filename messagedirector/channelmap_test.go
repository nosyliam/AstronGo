package messagedirector

import (
	"astrongo/util"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type MDParticipantFake struct{}

func (m *MDParticipantFake) RouteDatagram(datagram util.Datagram) {
	MD.Queue <- datagram
}

func (m *MDParticipantFake) HandleDatagram(datagram util.Datagram) {
	MD.Queue <- datagram
}

func (m *MDParticipantFake) Terminate() {}

func TestChannelMap_SubscribeRange(t *testing.T) {
	pt1 := &MDParticipantFake{}
	sub1 := &Subscriber{participant: MDParticipant(pt1), active: true}
	pt2 := &MDParticipantFake{}
	sub2 := &Subscriber{participant: MDParticipant(pt2), active: true}

	channelMap.SubscribeChannel(sub1, 532)
	channelMap.SubscribeRange(sub1, Range{500, 600})
	time.Sleep(time.Millisecond * 50)
	if _, ok := channelMap.subscriptions.Load(532); ok {
		t.Error("range subscription did not close single-channel subscription")
	}

	var recv chan interface{}
	recv = channelMap.Channel(util.Channel_t(555))
	dg := util.NewDatagram()
	dg.AddString("abc")
	recv <- dg
	out := <-MD.Queue
	require.EqualValues(t, out, dg)

	channelMap.SubscribeRange(sub2, Range{580, 700})
	recv = channelMap.Channel(util.Channel_t(585))
	recv <- dg
	<-MD.Queue
	<-MD.Queue

	channelMap.UnsubscribeRange(sub1, Range{590, 650})
	recv = channelMap.Channel(util.Channel_t(620))
	recv <- dg
	<-MD.Queue

	channelMap.SubscribeRange(sub1, Range{450, 600})
	recv = channelMap.Channel(util.Channel_t(611))
	recv <- dg
	<-MD.Queue

	channelMap.SubscribeRange(sub2, Range{300, 487})
	recv = channelMap.Channel(util.Channel_t(460))
	recv <- dg
	<-MD.Queue
	<-MD.Queue

	channelMap.UnsubscribeChannel(sub2, 460)
	recv = channelMap.Channel(util.Channel_t(460))
	recv <- dg
	<-MD.Queue

	channelMap.UnsubscribeRange(sub1, Range{0, 500})
	recv = channelMap.Channel(util.Channel_t(470))
	recv <- dg
	<-MD.Queue

	channelMap.UnsubscribeRange(sub1, Range{0, 1000})
	require.True(t, len(sub1.ranges) == 0)

}

func TestChannelMap_SubscribeChannel(t *testing.T) {
	pt := &MDParticipantFake{}
	sub := &Subscriber{participant: MDParticipant(pt), active: true}

	channelMap.SubscribeChannel(sub, 1000)
	require.Equal(t, int(sub.channels[0]), 1000)

	recv := channelMap.Channel(util.Channel_t(1000))
	dg := util.NewDatagram()
	dg.AddString("aaa")
	recv <- dg
	out := <-MD.Queue
	require.Empty(t, MD.Queue)
	require.EqualValues(t, out, dg)

	// Subscriber removal
	channelMap.UnsubscribeChannel(sub, 1000)
	time.Sleep(time.Millisecond * 5)
	if _, ok := channelMap.subscriptions.Load(1000); ok {
		t.Error("channel routine did not close when empty")
	}

}

func init() {
	MD = &MessageDirector{}
	MD.Queue = make(chan util.Datagram)
	channelMap = &ChannelMap{}
	channelMap.init()
}
