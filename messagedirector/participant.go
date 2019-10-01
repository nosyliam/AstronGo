package messagedirector

import (
	"astrongo/net"
	"astrongo/util"
	gonet "net"
	"time"
)

type MDParticipant interface {
	net.DatagramHandler

	// RouteDatagram routes a datagram to the MD
	RouteDatagram(util.Datagram)
}

type MDParticipantBase struct{ MDParticipant }

func (m *MDParticipantBase) initialize() {
	MD.participants = append(MD.participants, m)
}

func (m *MDParticipantBase) routeDatagram(datagram util.Datagram) {
	MD.Queue <- datagram
}

func (m *MDParticipantBase) handleDatagram(datagram util.Datagram) {

}

func (m *MDParticipantBase) postRemove() {

}

func (m *MDParticipantBase) terminate() {
	for n, participant := range MD.participants {
		if participant == m {
			MD.participants = append(MD.participants[:n], MD.participants[n+1:]...)
		}
	}

}

// MDNetworkParticipant represents a downstream MD connection
type MDNetworkParticipant struct {
	MDParticipantBase
	transport *net.Transport
	client    *net.Client
}

func NewMDParticipant(conn gonet.Conn) *MDNetworkParticipant {
	participant := &MDNetworkParticipant{}
	socket := net.NewSocketTransport(conn, 60*time.Second)
	client := net.NewClient(socket, participant)

	participant.client = client
	return participant
}
