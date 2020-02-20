package stateserver

import (
	"astrongo/dclass/dc"
	"astrongo/messagedirector"
	. "astrongo/util"
	"github.com/apex/log"
)

type FieldValues map[*dc.Field][]uint8

type DistributedObject struct {
	messagedirector.MDParticipantBase

	log *log.Entry

	stateserver *StateServer
	do          Doid_t
	parent      Doid_t
	zone        Zone_t
	dclass      *dc.Class

	requiredFields FieldValues
	ramFields      FieldValues

	aiChannel    Channel_t
	ownerChannel Channel_t

	context     int
	zoneObjects map[Zone_t][]Doid_t
}
