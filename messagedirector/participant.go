package messagedirector

import (
	"astrongo/net"
	. "astrongo/util"
	gonet "net"
	"time"
)

type MDParticipant interface {
	net.DatagramHandler

	// RouteDatagram routes a datagram through the MD
	RouteDatagram(Datagram)

	SubscribeChannel(Channel_t)
	UnsubscribeChannel(Channel_t)

	SubscribeRange(Range)
	UnsubscribeRange(Range)
	UnsubscribeAll()
}

type MDParticipantBase struct {
	MDParticipant

	subscriber  *Subscriber
	postRemoves map[Channel_t][]Datagram
}

func (m *MDParticipantBase) initialize() {
	m.postRemoves = make(map[Channel_t][]Datagram)
	MD.participants = append(MD.participants, m)
}

func (m *MDParticipantBase) RouteDatagram(datagram Datagram) {
	MD.Queue <- datagram
}

func (m *MDParticipantBase) PostRemove() {
	for sender, dgt := range m.postRemoves {
		for _, dg := range dgt {
			m.RouteDatagram(dg)
		}

		MD.RecallPostRemoves(sender)
	}
}

func (m *MDParticipantBase) AddPostRemove(ch Channel_t, dg Datagram) {
	m.postRemoves[ch] = append(m.postRemoves[ch], dg)
	MD.PreroutePostRemove(ch, dg)
}

func (m *MDParticipantBase) ClearPostRemoves(ch Channel_t) {
	delete(m.postRemoves, ch)
	MD.RecallPostRemoves(ch)
}

func (m *MDParticipantBase) SubscribeChannel(ch Channel_t) {
	channelMap.SubscribeChannel(m.subscriber, ch)
}

func (m *MDParticipantBase) UnsubscribeChannel(ch Channel_t) {
	channelMap.UnsubscribeChannel(m.subscriber, ch)
}

func (m *MDParticipantBase) SubscribeRange(rng Range) {
	channelMap.SubscribeRange(m.subscriber, rng)
}

func (m *MDParticipantBase) UnsubscribeRange(rng Range) {
	channelMap.UnsubscribeRange(m.subscriber, rng)
}

func (m *MDParticipantBase) terminate() {
	m.UnsubscribeAll()
	m.PostRemove()

	for n, participant := range MD.participants {
		if participant == m {
			MD.participants = append(MD.participants[:n], MD.participants[n+1:]...)
		}
	}

}

// MDNetworkParticipant represents a downstream MD connection
type MDNetworkParticipant struct {
	MDParticipantBase

	client *net.Client
	conn   gonet.Conn
}

func NewMDParticipant(conn gonet.Conn) *MDNetworkParticipant {
	participant := &MDNetworkParticipant{conn: conn}
	socket := net.NewSocketTransport(conn, 60*time.Second)

	participant.client = net.NewClient(socket, participant)
	participant.subscriber = &Subscriber{participant: participant, active: true}
	return participant
}

func (m *MDNetworkParticipant) HandleDatagram(dg Datagram) {

}

func (m *MDNetworkParticipant) Terminate(err error) {
	MDLog.Infof("Lost connection from %s: %s", m.conn.RemoteAddr(), err.Error())
	m.terminate()
}
