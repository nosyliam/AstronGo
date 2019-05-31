package messagedirector

import "astrongo/util"

type MDParticipant interface {
	HandleDatagram()
	RouteDatagram(util.Datagram)

	Terminate()
}

type MDParticipantBase struct{}

func (m *MDParticipantBase) RouteDatagram(datagram util.Datagram) {
	MD.Queue <- datagram
}
