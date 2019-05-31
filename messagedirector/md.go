package messagedirector

import "astrongo/util"

var MD *MessageDirector

type MessageDirector struct {
	ChannelMap

	// Connections within the context of the MessageDirector are represented as
	// participants; however, clients and objects on the SS may function as participants
	// as well. The MD will keep track of them and what channels they subscribe and route data to them.
	participants []MDParticipant

	// MD participants may directly queue datagarams to be routed by inserting it into the
	// queue channel, where they will be processed asynchronously
	Queue chan<- util.Datagram
}

func start() {
	MD = &MessageDirector{}
	MD.Queue = make(chan<- util.Datagram)
	MD.subscriptions = make(map[util.Channel_t][]Subscriber)
	MD.participants = make([]MDParticipant, 0)
}
