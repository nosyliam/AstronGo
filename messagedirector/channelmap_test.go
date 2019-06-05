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

func (m *MDParticipantFake) Terminate() {

}

func TestChannelMap_SubscribeChannel(t *testing.T) {
	pt := &MDParticipantFake{}
	sub := &Subscriber{participant: MDParticipant(pt), active: true}

	MD.SubscribeChannel(sub, 1000)
	require.Equal(t, int(sub.channels[0]), 1000)

	var recv interface{}
	var ok bool
	if recv, ok = MD.subscriptions.Load(util.Channel_t(1000)); !ok {
		t.Error("unable to load recv chan")
	}

	if ch, ok := recv.(chan interface{}); ok {
		dg := util.NewDatagram()
		dg.AddString("aaa")
		ch <- dg
		out := <-MD.Queue
		require.EqualValues(t, out, dg)

		// Subscriber removal
		sub.active = false
		ch <- sub
		ch <- dg
		time.Sleep(time.Millisecond * 5)
		require.Empty(t, MD.Queue)
	} else {
		t.Error("recv chan is invalid")
	}

}

func init() {
	start()
}
