package stateserver

import (
	"astrongo/core"
	"astrongo/messagedirector"
	. "astrongo/util"
	"fmt"
	"github.com/apex/log"
)

type StateServer struct {
	messagedirector.MDParticipantBase

	config  core.Role
	log     *log.Entry
	objects map[Doid_t]*DistributedObject
}

func NewStateServer(config core.Role) *StateServer {
	ss := &StateServer{
		config:  config,
		objects: make(map[Doid_t]*DistributedObject),
		log: log.WithFields(log.Fields{
			"name": fmt.Sprintf("StateServer (%d)", config.Control),
		}),
	}
	ss.Init()

	if Channel_t(config.Control) != INVALID_CHANNEL {
		ss.SubscribeChannel(Channel_t(config.Control))
		ss.SubscribeChannel(BCHAN_STATESERVERS)
	}

	return ss
}

func (s *StateServer) handleGenerate(dgi *DatagramIterator, other bool) {
	do := dgi.ReadDoid()
	parent := dgi.ReadDoid()
	zone := dgi.ReadZone()
	dc := dgi.ReadUint16()

	if _, ok := s.objects[do]; ok {
		s.log.Warnf("Received generate for already-existing object ID=%d", do)
		return
	}

	dclass, ok := core.DC.Class(int(dc))
	if !ok {
		s.log.Errorf("Received create for unknown dclass id %d", dc)
		return
	}
}

func (s *StateServer) handleDelete(dgi *DatagramIterator, sender Channel_t) {

}

func (s *StateServer) ReceiveDatagram(dg Datagram) {
	dgi := NewDatagramIterator(&dg)
	sender := dgi.ReadChannel()
	msgType := dgi.ReadUint16()

	switch msgType {
	case STATESERVER_CREATE_OBJECT_WITH_REQUIRED:
		s.handleGenerate(dgi, false)
	case STATESERVER_CREATE_OBJECT_WITH_REQUIRED_OTHER:
		s.handleGenerate(dgi, true)
	case STATESERVER_DELETE_AI_OBJECTS:
		s.handleDelete(dgi, sender)
	default:
		s.log.Warnf("Received unknown msgtype=%d", msgType)
	}
}
