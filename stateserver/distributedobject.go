package stateserver

import (
	"astrongo/dclass/dc"
	"astrongo/messagedirector"
	. "astrongo/util"
	"fmt"
	"github.com/apex/log"
)

type FieldValues map[dc.Field][]uint8

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
	explicitAi   bool

	context     int
	zoneObjects map[Zone_t][]Doid_t
}

// TODO for database support
func NewDistributedObjectWithData() {}

func NewDistributedObject(ss *StateServer, doid Doid_t, parent Doid_t,
	zone Zone_t, dclass *dc.Class, dgi *DatagramIterator, hasOther bool) *DistributedObject {
	do := &DistributedObject{
		stateserver:    ss,
		do:             doid,
		zone:           zone,
		dclass:         dclass,
		requiredFields: make(map[dc.Field][]uint8),
		ramFields:      make(map[dc.Field][]uint8),
		log: log.WithFields(log.Fields{
			"name": fmt.Sprintf("%s (%d)", dclass.Name(), doid),
		}),
	}

	for i := 0; i < dclass.GetNumFields(); i++ {
		field := dclass.GetField(i)
		if field.Keywords().HasKeyword("required") {
			if _, ok := field.(*dc.MolecularField); ok {
				continue
			}

			do.requiredFields[field] = dgi.UnpackFieldtoUint8(field)
		}
	}

	if hasOther {
		count := dgi.ReadUint16()
		for i := 0; i < int(count); i++ {
			id := dgi.ReadUint16()
			field, ok := dclass.GetFieldById(uint(id))
			if !ok {
				do.log.Errorf("Receieved unknown field with ID %d within an OTHER section!", id)
				break
			}

			if field.Keywords().HasKeyword("ram") {
				do.ramFields[field] = dgi.UnpackFieldtoUint8(field)
			} else {
				do.log.Errorf("Received non-RAM field %s within an OTHER section!", field.Name())
				dgi.SkipField(field)
			}

		}
	}

	do.SubscribeChannel(Channel_t(doid))
	do.log.Debug("Object instantiated ...")

	dgi.SeekPayload()

	return do
}

func (d *DistributedObject) appendRequiredData(dg Datagram, client bool, owner bool) {
	dg.AddDoid(d.do)
	dg.AddLocation(d.parent, Doid_t(d.zone))
	dg.AddUint16(uint16(d.dclass.ClassId()))
	count := d.dclass.GetNumFields()
	for i := 0; i < int(count); i++ {
		field := d.dclass.GetField(i)
		if _, ok := field.(*dc.MolecularField); ok {
			continue
		}

		if field.Keywords().HasKeyword("required") && (!client || field.Keywords().HasKeyword("broadcast") ||
			field.Keywords().HasKeyword("clrecv") || (owner && field.Keywords().HasKeyword("ownrecv"))) {
			dg.AddData(d.requiredFields[field])
		}
	}
}

func (d *DistributedObject) appendOtherData(dg Datagram, client bool, owner bool) {
	if client {
		var broadcastFields []dc.Field
		for field, _ := range d.ramFields {
			if field.Keywords().HasKeyword("broadcast") || field.Keywords().HasKeyword("clrecv") ||
				(owner && field.Keywords().HasKeyword("ownrecv")) {
				broadcastFields = append(broadcastFields, field)
			}
		}

		dg.AddUint16(uint16(len(broadcastFields)))
		for _, field := range broadcastFields {
			dg.AddUint16(uint16(field.Id()))
			dg.AddData(d.ramFields[field])
		}
	} else {
		dg.AddUint16(uint16(len(d.ramFields)))
		for field, data := range d.ramFields {
			dg.AddUint16(uint16(field.Id()))
			dg.AddData(data)
		}
	}
}

func (d *DistributedObject) HandleDatagram(dg Datagram, dgi *DatagramIterator) {
	sender := dgi.ReadChannel()
	msgType := dgi.ReadUint16()

	switch msgType {
	case STATESERVER_DELETE_AI_OBJECTS:
	case STATESERVER_OBJECT_DELETE_RAM:
	case STATESERVER_OBJECT_SET_FIELD:
	case STATESERVER_OBJECT_SET_FIELDS:
	case STATESERVER_OBJECT_CHANGING_AI:
	case STATESERVER_OBJECT_SET_AI:
	case STATESERVER_OBJECT_GET_AI:
	case STATESERVER_OBJECT_CHANGING_LOCATION:
	case STATESERVER_OBJECT_LOCATION_ACK:
	case STATESERVER_OBJECT_SET_LOCATION:
	case STATESERVER_OBJECT_GET_LOCATION:
	case STATESERVER_OBJECT_GET_LOCATION_RESP:
	case STATESERVER_OBJECT_GET_ALL:
	case STATESERVER_OBJECT_GET_FIELD:
	case STATESERVER_OBJECT_GET_FIELDS:
	case STATESERVER_OBJECT_SET_OWNER:
	case STATESERVER_OBJECT_GET_ZONE_OBJECTS:
		fallthrough
	case STATESERVER_OBJECT_GET_ZONES_OBJECTS:
	case STATESERVER_GET_ACTIVE_ZONES:
	default:
		if msgType < STATESERVER_MSGTYPE_MIN || msgType > STATESERVER_MSGTYPE_MAX {
			d.log.Warnf("Recieved unknown message of type %d.", msgType)
		} else {
			d.log.Warnf("Ignoring message of type %d.", msgType)
		}

	}
}
