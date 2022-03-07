package clientagent

import (
	"astrongo/core"
	"astrongo/dclass/dc"
	"astrongo/eventlogger"
	"astrongo/messagedirector"
	"astrongo/net"
	. "astrongo/util"
	"fmt"
	"github.com/apex/log"
	gonet "net"
	"sync"
	"time"
)

type ClientState int

const (
	CLIENT_STATE_NEW ClientState = iota
	CLIENT_STATE_ANONYMOUS
	CLIENT_STATE_ESTABLISHED
)

type DeclaredObject struct {
	do Doid_t
	dc *dc.Class
}

type OwnedObject struct {
	DeclaredObject
	parent Doid_t
	zone   Zone_t
}

type VisibleObject struct {
	DeclaredObject
	parent Doid_t
	zone   Zone_t
}

type Interest struct {
	id     uint16
	parent Doid_t
	zones  []Zone_t
}

func (i *Interest) hasZone(zone Zone_t) bool {
	for _, z := range i.zones {
		if z == zone {
			return true
		}
	}
	return false
}

type Client struct {
	sync.Mutex
	messagedirector.MDParticipantBase

	// Client properties
	config core.Role
	ca     *ClientAgent
	log    *log.Entry

	allocatedChannel Channel_t
	channel          Channel_t
	state            ClientState
	context          uint32

	seenObjects       []Doid_t
	sessionObjects    []Doid_t
	historicalObjects []Doid_t

	visibleObjects   map[Doid_t]VisibleObject
	declaredObjects  map[Doid_t]DeclaredObject
	ownedObjects     map[Doid_t]OwnedObject
	pendingObjects   map[Doid_t]uint32
	interests        map[uint16]Interest
	pendingInterests map[uint32]*InterestOperation
	sendableFields   map[Doid_t][]uint16

	// AstronClient properties
	conn   gonet.Conn
	client *net.Client
	lock   sync.Mutex

	cleanDisconnect  bool
	allowedInterests InterestPermission
	heartbeat        *time.Ticker
	finish           chan bool
}

func (c *Client) NewClient(config core.Role, ca *ClientAgent, conn gonet.Conn) *Client {
	client := &Client{
		config:           config,
		ca:               ca,
		visibleObjects:   make(map[Doid_t]VisibleObject),
		declaredObjects:  make(map[Doid_t]DeclaredObject),
		ownedObjects:     make(map[Doid_t]OwnedObject),
		pendingObjects:   make(map[Doid_t]uint32),
		interests:        make(map[uint16]Interest),
		pendingInterests: make(map[uint32]*InterestOperation),
		sendableFields:   make(map[Doid_t][]uint16),
	}

	client.init(config, conn)
	client.Init(c)

	client.allocatedChannel = ca.Allocate()
	if client.allocatedChannel == 0 {
		c.sendDisconnect(CLIENT_DISCONNECT_GENERIC, "Client capacity reached", false)
		return nil
	}
	client.channel = client.allocatedChannel

	client.log = log.WithFields(log.Fields{
		"name": fmt.Sprintf("Client (%d)", client.channel),
	})

	client.SubscribeChannel(client.channel)
	client.SubscribeChannel(BCHAN_CLIENTS)

	return client
}

func (c *Client) sendDisconnect(reason uint16, error string, security bool) {
	// TODO: Implement security loglevel
	var eventType string
	if security {
		c.log.Errorf("[SECURITY] Ejecting client (%s): %s", reason, error)
		eventType = "client-ejected-security"
	} else {
		c.log.Errorf("Ejecting client (%s): %s", reason, error)
		eventType = "client-ejected"
	}

	event := eventlogger.NewLoggedEvent(eventType, "")
	event.Add("reason_code", string(reason))
	event.Add("reason_msg", error)
	c.logEvent(event)

	if c.client.Connected() {
		resp := NewDatagram()
		resp.AddUint16(CLIENT_EJECT)
		resp.AddUint16(reason)
		resp.AddString(error)
		c.client.SendDatagram(resp)

		c.cleanDisconnect = true
		c.client.Close()
	}
}

func (c *Client) logEvent(event eventlogger.LoggedEvent) {
	event.Add("sender", fmt.Sprintf("Client: %d", c.channel))
	event.Send()
}

func (c *Client) annihilate() {
	c.Lock()
	defer c.Unlock()

	if c.IsTerminated() {
		return
	}

	c.ca.Tracker.free(c.channel)

	// Delete all session object
	for len(c.sessionObjects) > 0 {
		var do Doid_t
		do, c.sessionObjects = c.sessionObjects[0], c.sessionObjects[1:]
		c.log.Debugf("Client exited, deleting session object ID=%d", do)
		dg := NewDatagram()
		dg.AddServerHeader(Channel_t(do), c.channel, STATESERVER_OBJECT_DELETE_RAM)
		dg.AddDoid(do)
		c.RouteDatagram(dg)
	}

	for _, int := range c.pendingInterests {
		int.finish()
	}

	c.Cleanup()
}

func (c *Client) lookupInterests(parent Doid_t, zone Zone_t) []Interest {
	var interests []Interest
	for _, int := range c.interests {
		if parent == int.parent && int.hasZone(zone) {
			interests = append(interests, int)
		}
	}
	return interests
}

func (c *Client) buildInterest(dgi *DatagramIterator, multiple bool) Interest {
	int := Interest{
		id:     dgi.ReadUint16(),
		parent: dgi.ReadDoid(),
	}

	count := uint16(1)
	if multiple {
		count = dgi.ReadUint16()
	}

	for count != 0 {
		int.zones = append(int.zones, dgi.ReadZone())
		count--
	}

	return int
}

func (c *Client) addInterest(i Interest, context uint32, caller Channel_t) {
	var zones []Zone_t

	for _, zone := range i.zones {
		if len(c.lookupInterests(i.parent, zone)) == 0 {
			zones = append(zones, zone)
		}
	}

	if prevInt, ok := c.interests[i.id]; ok {
		// This interest already exists, so it is being altered
		var killedZones []Zone_t

		for _, zone := range prevInt.zones {
			if len(c.lookupInterests(i.parent, zone)) > 1 {
				// Another interest has this zone, so ignore it
				continue
			}

			if i.parent != prevInt.parent || i.hasZone(zone) {
				killedZones = append(killedZones, zone)
			}
		}

		c.closeZones(prevInt.parent, killedZones)
	}
	c.interests[i.id] = i

	if len(zones) == 0 {
		// We aren't requesting any new zones, so let the client know we finished
		c.notifyInterestDone(i.id, []Channel_t{caller})
		c.handleInterestDone(i.id, context)
		return
	}

	// Build a new IOP otherwise
	c.context++
	iop := NewInterestOperation(c, c.config.Tuning.Interest_Timeout, i.id,
		context, c.context, i.parent, zones, caller)
	c.pendingInterests[c.context] = iop

	resp := NewDatagram()
	resp.AddServerHeader(Channel_t(i.parent), c.channel, STATESERVER_OBJECT_GET_ZONES_OBJECTS)
	resp.AddUint32(c.context)
	resp.AddDoid(i.parent)
	resp.AddUint16(uint16(len(zones)))
	for _, zone := range zones {
		resp.AddZone(zone)
		c.SubscribeChannel(LocationAsChannel(i.parent, zone))
	}
	c.RouteDatagram(resp)
}

func (c *Client) removeInterest(i Interest, context uint32, caller Channel_t) {
	var zones []Zone_t

	for _, zone := range i.zones {
		if len(c.lookupInterests(i.parent, zone)) == 1 {
			zones = append(zones, zone)
		}
	}

	c.closeZones(i.parent, zones)
	c.notifyInterestDone(i.id, []Channel_t{caller})
	c.handleInterestDone(i.id, context)

	delete(c.interests, i.id)
}

func (c *Client) closeZones(parent Doid_t, zones []Zone_t) {
	var toRemove []Doid_t

	for _, obj := range c.visibleObjects {
		if obj.parent != parent {
			// Object does not belong to the parent in question
			continue
		}

		for i := range zones {
			if zones[i] == obj.zone {
				for i := range c.sessionObjects {
					if c.sessionObjects[i] == obj.do {
						c.sendDisconnect(CLIENT_DISCONNECT_SESSION_OBJECT_DELETED,
							"A session object has unexpectedly left interest.", false)
						return
					}
				}

				c.handleRemoveObject(obj.do)
				for i, o := range c.seenObjects {
					if o == obj.do {
						c.seenObjects = append(c.seenObjects[:i], c.seenObjects[i+1:]...)
					}
				}
				toRemove = append(toRemove, obj.do)
			}
		}
	}

	for _, do := range toRemove {
		delete(c.visibleObjects, do)
	}

	for _, zone := range zones {
		c.UnsubscribeChannel(LocationAsChannel(parent, zone))
	}
}

func (c *Client) historicalObject(do Doid_t) bool {
	for i := range c.historicalObjects {
		if c.historicalObjects[i] == do {
			return true
		}
	}
	return false
}

func (c *Client) lookupObject(do Doid_t) *dc.Class {
	// Search UberDOGs
	for i := range core.Uberdogs {
		if core.Uberdogs[i].Id == do {
			return core.Uberdogs[i].Class
		}
	}

	// Check the object cache
	if obj, ok := c.ownedObjects[do]; ok {
		return obj.dc
	}

	for i := range c.seenObjects {
		if c.seenObjects[i] == do {
			if obj, ok := c.visibleObjects[do]; ok {
				return obj.dc
			}
		}
	}

	// Check declared objects
	if obj, ok := c.declaredObjects[do]; ok {
		return obj.dc
	}

	// We don't know :(
	return nil
}

func (c *Client) tryQueuePending(do Doid_t, dg Datagram) bool {
	if context, ok := c.pendingObjects[do]; ok {
		if iop, ok := c.pendingInterests[context]; ok {
			iop.pendingQueue <- dg
			return true
		}
	}
	return false
}

func (c *Client) handleObjectEntrance(dgi *DatagramIterator, other bool) {
	do, parent, zone, dc := dgi.ReadDoid(), dgi.ReadDoid(), dgi.ReadZone(), dgi.ReadUint16()

	delete(c.pendingObjects, do)

	for i := range c.seenObjects {
		if c.seenObjects[i] == do {
			return
		}
	}

	if _, ok := c.ownedObjects[do]; ok {
		for i := range c.sessionObjects {
			if c.sessionObjects[i] == do {
				return
			}
		}
	}

	if _, ok := c.visibleObjects[do]; !ok {
		cls, _ := core.DC.Class(int(dc))
		c.visibleObjects[do] = VisibleObject{
			DeclaredObject: DeclaredObject{
				do: do,
				dc: cls,
			},
			parent: parent,
			zone:   zone,
		}
	}
	c.seenObjects = append(c.seenObjects, do)

	c.handleAddObject(do, parent, zone, dc, dgi, other)
}

func (c *Client) notifyInterestDone(interestId uint16, callers []Channel_t) {
	if len(callers) == 0 {
		return
	}

	resp := NewDatagram()
	resp.AddMultipleServerHeader(callers, c.channel, CLIENTAGENT_DONE_INTEREST_RESP)
	resp.AddChannel(c.channel)
	resp.AddUint16(interestId)
	c.RouteDatagram(resp)
}

func (c *Client) HandleDatagram(dg Datagram, dgi *DatagramIterator) {
	c.Lock()
	defer c.Unlock()

	sender := dgi.ReadChannel()
	msgType := dgi.ReadUint16()
	if sender == c.channel {
		return
	}

	switch msgType {
	case CLIENTAGENT_EJECT:
		reason, error := dgi.ReadUint16(), dgi.ReadString()
		c.sendDisconnect(reason, error, false)
	case CLIENTAGENT_DROP:
		c.handleDrop()
	case CLIENTAGENT_SET_STATE:
		c.state = ClientState(dgi.ReadUint16())
	case CLIENTAGENT_ADD_INTEREST:
		c.context++
		int := c.buildInterest(dgi, false)
		c.handleAddInterest(int, c.context)
		c.addInterest(int, c.context, sender)
	case CLIENTAGENT_ADD_INTEREST_MULTIPLE:
		c.context++
		int := c.buildInterest(dgi, true)
		c.handleAddInterest(int, c.context)
		c.addInterest(int, c.context, sender)
	case CLIENTAGENT_REMOVE_INTEREST:
		c.context++
		id := dgi.ReadUint16()
		int := c.interests[id]
		c.handleRemoveInterest(id, c.context)
		c.removeInterest(int, c.context, sender)
	case CLIENTAGENT_SET_CLIENT_ID:
		if c.channel != c.allocatedChannel {
			c.UnsubscribeChannel(c.channel)
		}

		c.channel = dgi.ReadChannel()
		c.SubscribeChannel(c.channel)
	case CLIENTAGENT_SEND_DATAGRAM:
		c.client.SendDatagram(dg)
	case CLIENTAGENT_OPEN_CHANNEL:
		c.SubscribeChannel(dgi.ReadChannel())
	case CLIENTAGENT_CLOSE_CHANNEL:
		c.UnsubscribeChannel(dgi.ReadChannel())
	case CLIENTAGENT_ADD_POST_REMOVE:
		c.AddPostRemove(c.allocatedChannel, *dgi.ReadDatagram())
	case CLIENTAGENT_CLEAR_POST_REMOVES:
		c.ClearPostRemoves(c.allocatedChannel)
	case CLIENTAGENT_DECLARE_OBJECT:
		do, dc := dgi.ReadDoid(), dgi.ReadUint16()

		if _, ok := c.declaredObjects[do]; ok {
			c.log.Warnf("Received object declaration for previously declared object %d", do)
			return
		}

		cls, _ := core.DC.Class(int(dc))
		c.declaredObjects[do] = DeclaredObject{
			do: do,
			dc: cls,
		}
	case CLIENTAGENT_UNDECLARE_OBJECT:
		do := dgi.ReadDoid()

		if _, ok := c.declaredObjects[do]; ok {
			c.log.Warnf("Received object de-declaration for previously declared object %d", do)
			return
		}

		delete(c.declaredObjects, do)
	case CLIENTAGENT_SET_FIELDS_SENDABLE:
		do, count := dgi.ReadDoid(), dgi.ReadUint16()

		var fields []uint16
		for count != 0 {
			fields = append(fields, dgi.ReadUint16())
		}
		c.sendableFields[do] = fields
	case CLIENTAGENT_ADD_SESSION_OBJECT:
		do := dgi.ReadDoid()
		for _, d := range c.sessionObjects {
			if d == do {
				c.log.Warnf("Received add sesion object with existing ID=%d", do)
			}
		}

		c.log.Debugf("Added session object with ID %d", do)
		c.sessionObjects = append(c.sessionObjects, do)
	case CLIENTAGENT_REMOVE_SESSION_OBJECT:
		do := dgi.ReadDoid()
		for _, d := range c.sessionObjects {
			if d == do {
				break
			}
			c.log.Warnf("Received remove sesion object with non-existant ID=%d", do)
		}

		c.log.Debugf("Removed session object with ID %d", do)
		for i, o := range c.sessionObjects {
			if o == do {
				c.sessionObjects = append(c.sessionObjects[:i], c.sessionObjects[i+1:]...)
			}
		}
	case CLIENTAGENT_GET_TLVS:
		resp := NewDatagram()
		resp.AddServerHeader(sender, c.channel, CLIENTAGENT_GET_TLVS_RESP)
		resp.AddUint32(dgi.ReadUint32())
		// resp.AddDataBlob(c.client.Tlvs())
		c.RouteDatagram(resp)
		// TODO: Implement HAProxy
	case CLIENTAGENT_GET_NETWORK_ADDRESS:
		resp := NewDatagram()
		resp.AddServerHeader(sender, c.channel, CLIENTAGENT_GET_NETWORK_ADDRESS_RESP)
		resp.AddUint32(dgi.ReadUint32())
		resp.AddString(c.client.RemoteIP())
		resp.AddUint16(c.client.RemotePort())
		resp.AddString(c.client.LocalIP())
		resp.AddUint16(c.client.LocalPort())
		c.RouteDatagram(resp)
	case STATESERVER_OBJECT_SET_FIELD:
		do := dgi.ReadDoid()
		if c.lookupObject(do) == nil {
			if c.tryQueuePending(do, dg) {
				return
			}
			c.log.Warnf("Received server-side field update for unknown object %d", do)
		}

		if sender != c.channel {
			field := dgi.ReadUint16()
			c.handleSetField(do, field, dgi)
		}
	case STATESERVER_OBJECT_SET_FIELDS:
		do := dgi.ReadDoid()
		if c.lookupObject(do) == nil {
			if c.tryQueuePending(do, dg) {
				return
			}
			c.log.Warnf("Received server-side multi-field update for unknown object %d", do)
		}

		if sender != c.channel {
			fields := dgi.ReadUint16()
			c.handleSetField(do, fields, dgi)
		}
	case STATESERVER_OBJECT_DELETE_RAM:
		do := dgi.ReadDoid()
		if c.lookupObject(do) == nil {
			if c.tryQueuePending(do, dg) {
				return
			}
			c.log.Warnf("Received server-side object delete for unknown object %d", do)
		}

		for i, so := range c.sessionObjects {
			if so == do {
				c.sessionObjects = append(c.sessionObjects[:i], c.sessionObjects[i+1:]...)
				c.sendDisconnect(CLIENT_DISCONNECT_SESSION_OBJECT_DELETED,
					fmt.Sprintf("The session object with id %d has been unexpectedly deleted", do), false)
			}
		}

		for i, so := range c.seenObjects {
			if so == do {
				c.seenObjects = append(c.seenObjects[:i], c.seenObjects[i+1:]...)
				c.handleRemoveObject(do)
			}
		}

		if _, ok := c.ownedObjects[do]; ok {
			c.handleRemoveOwnership(do)
			delete(c.ownedObjects, do)
		}

		c.historicalObjects = append(c.historicalObjects, do)
		delete(c.visibleObjects, do)
	case STATESERVER_OBJECT_ENTER_OWNER_WITH_REQUIRED_OTHER:
		fallthrough
	case STATESERVER_OBJECT_ENTER_OWNER_WITH_REQUIRED:
		do, parent, zone, dc := dgi.ReadDoid(), dgi.ReadDoid(), dgi.ReadZone(), dgi.ReadUint16()

		if _, ok := c.ownedObjects[do]; !ok {
			cls, _ := core.DC.Class(int(dc))
			c.ownedObjects[do] = OwnedObject{
				DeclaredObject: DeclaredObject{
					do: do,
					dc: cls,
				},
				parent: parent,
				zone:   zone,
			}
		}

		other := msgType == STATESERVER_OBJECT_ENTER_OWNER_WITH_REQUIRED_OTHER
		c.handleAddOwnership(do, parent, zone, dc, dgi, other)
	case STATESERVER_OBJECT_ENTER_LOCATION_WITH_REQUIRED:
		fallthrough
	case STATESERVER_OBJECT_ENTER_LOCATION_WITH_REQUIRED_OTHER:
		do, parent, zone := dgi.ReadDoid(), dgi.ReadDoid(), dgi.ReadZone()
		for id, iop := range c.pendingInterests {
			if iop.parent == parent && iop.hasZone(zone) {
				iop.pendingQueue <- dg
				c.pendingObjects[do] = id
			}
		}
	case STATESERVER_OBJECT_ENTER_INTEREST_WITH_REQUIRED:
		fallthrough
	case STATESERVER_OBJECT_ENTER_INTEREST_WITH_REQUIRED_OTHER:
	case STATESERVER_OBJECT_GET_ZONE_COUNT_RESP:
	case STATESERVER_OBJECT_CHANGING_LOCATION:
	case STATESERVER_OBJECT_CHANGING_OWNER:
	default:
		c.log.Errorf("Received unknown server msgtype %d", msgType)
	}
}

type InterestOperation struct {
	hasTotal bool
	finished bool
	total    int

	client         *Client
	interestId     uint16
	clientContext  uint32
	requestContext uint32
	parent         Doid_t

	zones   []Zone_t
	callers []Channel_t

	generateQueue chan Datagram
	pendingQueue  chan Datagram
}

func NewInterestOperation(client *Client, timeout int, interestId uint16,
	clientContext uint32, requestContext uint32, parent Doid_t, zones []Zone_t, caller Channel_t) *InterestOperation {
	iop := &InterestOperation{
		client:         client,
		interestId:     interestId,
		clientContext:  clientContext,
		requestContext: requestContext,
		parent:         parent,
		zones:          zones,
		generateQueue:  make(chan Datagram, 100),
		pendingQueue:   make(chan Datagram, 100),
		callers:        []Channel_t{caller},
	}

	// Timeout
	go func() {
		time.Sleep(time.Duration(timeout) * time.Second)
		if !iop.finished {
			client.log.Warnf("Interest operation timed out; forcing")
			iop.finish()
		}
	}()

	return iop
}

func (i *InterestOperation) hasZone(zone Zone_t) bool {
	for _, z := range i.zones {
		if z == zone {
			return true
		}
	}
	return false
}

func (i *InterestOperation) setExpected(total int) {
	if !i.hasTotal {
		i.total = total
		i.hasTotal = true
	}
}

func (i *InterestOperation) ready() bool {
	return i.hasTotal && len(i.generateQueue) >= i.total
}

func (i *InterestOperation) finish() {
	// We need to acquire our client's lock because we can't risk
	//  concurrent writes to pendingInterests
	i.client.Lock()
	defer i.client.Unlock()

	close(i.generateQueue)
	close(i.pendingQueue)
	i.finished = true

	for generate := range i.generateQueue {
		dgi := NewDatagramIterator(&generate)
		dgi.SeekPayload()
		dgi.Skip(Chansize) // Skip sender

		msgType := dgi.ReadUint16()
		other := msgType == STATESERVER_OBJECT_ENTER_INTEREST_WITH_REQUIRED_OTHER

		dgi.Skip(Dgsize) // Skip request context
		i.client.handleObjectEntrance(dgi, other)
	}

	// Send out interest done messages
	i.client.notifyInterestDone(i.interestId, i.callers)
	i.client.handleInterestDone(i.interestId, i.clientContext)

	// Delete the IOP
	delete(i.client.pendingInterests, i.requestContext)
	for dg := range i.pendingQueue {
		dgi := NewDatagramIterator(&dg)
		dgi.SeekPayload()
		i.client.HandleDatagram(dg, dgi)
	}
}
