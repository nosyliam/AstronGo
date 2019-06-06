package messagedirector

import (
	"astrongo/net"
	"astrongo/util"
)

type MDParticipant interface {
	net.DatagramHandler

	// RouteDatagram routes a datagram to the MD
	RouteDatagram(util.Datagram)

	// Terminate mutually unsubscribes the participant from all channels and calls any post-removes it may have
	Terminate()
}

type MDParticipantBase struct{}

func (m *MDParticipantBase) RouteDatagram(datagram util.Datagram) {
	MD.Queue <- datagram
}

func (m *MDParticipantBase) HandleDatagram(datagram util.Datagram) {

}

func (m *MDParticipantBase) Terminate() {

}

// MDNetworkParticipant represents a downstream MD connection
type MDNetworkParticipant struct {
	transport net.Transport
}
