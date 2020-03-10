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

	aiChannel          Channel_t
	ownerChannel       Channel_t
	explicitAi         bool
	parentSynchronized bool

	context     uint32
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
	dg.AddLocation(d.parent, d.zone)
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

func (d *DistributedObject) sendInterestEntry(location Channel_t, context uint32) {
	msgType := STATESERVER_OBJECT_ENTER_INTEREST_WITH_REQUIRED
	if len(d.ramFields) != 0 {
		msgType = STATESERVER_OBJECT_ENTER_INTEREST_WITH_REQUIRED_OTHER
	}
	dg := NewDatagram()
	dg.AddServerHeader(location, Channel_t(d.do), uint16(msgType))
	dg.AddUint32(context)
	d.appendRequiredData(dg, true, false)
	if len(d.ramFields) != 0 {
		d.appendOtherData(dg, true, false)
	}
	d.RouteDatagram(dg)
}

func (d *DistributedObject) sendLocationEntry(location Channel_t) {
	msgType := STATESERVER_OBJECT_ENTER_LOCATION_WITH_REQUIRED
	if len(d.ramFields) != 0 {
		msgType = STATESERVER_OBJECT_ENTER_LOCATION_WITH_REQUIRED_OTHER
	}
	dg := NewDatagram()
	dg.AddServerHeader(location, Channel_t(d.do), uint16(msgType))
	d.appendRequiredData(dg, true, false)
	if len(d.ramFields) != 0 {
		d.appendOtherData(dg, true, false)
	}
	d.RouteDatagram(dg)
}

func (d *DistributedObject) sendAiEntry(ai Channel_t) {
	msgType := STATESERVER_OBJECT_ENTER_AI_WITH_REQUIRED
	if len(d.ramFields) != 0 {
		msgType = STATESERVER_OBJECT_ENTER_AI_WITH_REQUIRED_OTHER
	}
	dg := NewDatagram()
	dg.AddServerHeader(ai, Channel_t(d.do), uint16(msgType))
	d.appendRequiredData(dg, false, false)
	if len(d.ramFields) != 0 {
		d.appendOtherData(dg, false, false)
	}
	d.RouteDatagram(dg)
}

func (d *DistributedObject) sendOwnerEntry(owner Channel_t) {
	msgType := STATESERVER_OBJECT_ENTER_OWNER_WITH_REQUIRED
	if len(d.ramFields) != 0 {
		msgType = STATESERVER_OBJECT_ENTER_OWNER_WITH_REQUIRED_OTHER
	}
	dg := NewDatagram()
	dg.AddServerHeader(owner, Channel_t(d.do), uint16(msgType))
	d.appendRequiredData(dg, true, true)
	if len(d.ramFields) != 0 {
		d.appendOtherData(dg, true, true)
	}
	d.RouteDatagram(dg)
}

func (d *DistributedObject) handleLocationChange(parent Doid_t, zone Zone_t, sender Channel_t) {
	var targets []Channel_t
	oldParent := d.parent
	oldZone := d.zone

	if d.aiChannel != INVALID_CHANNEL {
		targets = append(targets, d.aiChannel)
	}

	if d.ownerChannel != INVALID_CHANNEL {
		targets = append(targets, d.ownerChannel)
	}

	if parent == d.do {
		d.log.Warn("Object cannot be parented to itself.")
		return
	}

	// Parent change
	if parent != oldParent {
		if oldParent != INVALID_DOID {
			d.UnsubscribeChannel(ParentToChildren(d.parent))
			targets = append(targets, Channel_t(d.parent))
			targets = append(targets, LocationAsChannel(d.parent, d.zone))
		}

		d.parent = parent
		d.zone = zone

		if parent != INVALID_DOID {
			d.SubscribeChannel(ParentToChildren(parent))
			if !d.explicitAi {
				// Retrieve parent AI
				dg := NewDatagram()
				dg.AddServerHeader(Channel_t(parent), Channel_t(d.do), STATESERVER_OBJECT_GET_AI)
				dg.AddUint32(d.context)
				d.RouteDatagram(dg)
				d.context++
			}
			targets = append(targets, Channel_t(parent))
		} else if !d.explicitAi {
			d.aiChannel = INVALID_CHANNEL
		}
	} else if zone != oldZone {
		d.zone = zone
		targets = append(targets, Channel_t(d.parent))
		targets = append(targets, LocationAsChannel(d.parent, d.zone))
	}

	// Broadcast location change message
	dg := NewDatagram()
	dg.AddMultipleServerHeader(targets, sender, STATESERVER_OBJECT_CHANGING_LOCATION)
	dg.AddDoid(d.do)
	dg.AddLocation(parent, zone)
	dg.AddLocation(oldParent, oldZone)

	d.parentSynchronized = false

	if parent != INVALID_DOID {
		d.sendLocationEntry(LocationAsChannel(parent, zone))
	}
}

func (d *DistributedObject) handleAiChange(ai Channel_t, sender Channel_t, explicit bool) {
	var targets []Channel_t
	oldAi := d.aiChannel
	if ai == oldAi {
		return
	}

	if oldAi != INVALID_CHANNEL {
		targets = append(targets, oldAi)
	}

	if len(d.zoneObjects) != 0 {
		// Notify children of the change
		targets = append(targets, ParentToChildren(d.do))
	}

	d.aiChannel = ai
	d.explicitAi = explicit

	dg := NewDatagram()
	dg.AddMultipleServerHeader(targets, sender, STATESERVER_OBJECT_CHANGING_AI)
	dg.AddDoid(d.do)
	dg.AddChannel(ai)
	dg.AddChannel(oldAi)
	d.RouteDatagram(dg)

	if ai != INVALID_CHANNEL {
		d.log.Debugf("Sending AI entry to %d", ai)
		d.sendAiEntry(ai)
	}
}

func (d *DistributedObject) annihilate(sender Channel_t, notifyParent bool) {
	var targets []Channel_t
	if d.parent != INVALID_DOID {
		targets = append(targets, LocationAsChannel(d.parent, d.zone))
		if notifyParent {
			dg := NewDatagram()
			dg.AddServerHeader(Channel_t(d.parent), sender, STATESERVER_OBJECT_CHANGING_LOCATION)
			dg.AddDoid(d.do)
			dg.AddLocation(INVALID_DOID, 0)
			dg.AddLocation(d.parent, d.zone)
			d.RouteDatagram(dg)
		}
	}

	if d.ownerChannel != INVALID_CHANNEL {
		targets = append(targets, d.ownerChannel)
	}

	if d.aiChannel != INVALID_CHANNEL {
		targets = append(targets, d.aiChannel)
	}

	dg := NewDatagram()
	dg.AddMultipleServerHeader(targets, sender, STATESERVER_OBJECT_DELETE_RAM)
	dg.AddDoid(d.do)
	d.RouteDatagram(dg)

	d.deleteChildren(sender)
	delete(d.stateserver.objects, d.do)
	d.log.Debug("Deleted object.")

	d.Cleanup()
}

func (d *DistributedObject) deleteChildren(sender Channel_t) {
	if len(d.zoneObjects) != 0 {
		dg := NewDatagram()
		dg.AddServerHeader(ParentToChildren(d.do), sender, STATESERVER_OBJECT_DELETE_CHILDREN)
		dg.AddDoid(d.do)
		d.RouteDatagram(dg)
	}
}

func (d *DistributedObject) wakeChildren() {
	dg := NewDatagram()
	dg.AddServerHeader(ParentToChildren(d.do), Channel_t(d.do), STATESERVER_OBJECT_GET_LOCATION)
	dg.AddUint32(STATESERVER_CONTEXT_WAKE_CHILDREN)
	d.RouteDatagram(dg)
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
