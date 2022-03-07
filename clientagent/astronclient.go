package clientagent

import (
	"astrongo/core"
	"astrongo/eventlogger"
	"astrongo/net"
	. "astrongo/util"
	"fmt"
	gonet "net"
	"time"
)

type InterestPermission int

const (
	INTERESTS_ENABLED InterestPermission = iota
	INTERESTS_VISIBLE
	INTERESTS_DISABLED
)

// N.B. The purpose of this file is to separate implementations of ReceiveDatagram
//  and HandleDatagram and their associated functions-- normally, this would be done
//  by having two separate classes Client and AstronClient, but Go has zero support
//  for the virtual functions required to achieve this level of organization. Thus, two
//  distinct files still exist, but implement functions to the same class.

func (c *Client) init(config core.Role, conn gonet.Conn) {
	switch config.Client.Add_Interest {
	case "enabled":
		c.allowedInterests = INTERESTS_ENABLED
	case "visible":
		c.allowedInterests = INTERESTS_VISIBLE
	default:
		c.allowedInterests = INTERESTS_DISABLED
	}
	if config.Client.Heartbeat_Timeout != 0 {
		c.heartbeat = time.NewTicker(time.Duration(config.Client.Heartbeat_Timeout) * time.Second)
		go c.startHeartbeat()
	}

	socket := net.NewSocketTransport(conn,
		time.Duration(config.Client.Keepalive)*time.Second, config.Client.Write_Buffer_Size)
	c.client = net.NewClient(socket, c, time.Duration(config.Client.Keepalive)*time.Second)

	if !c.client.Local() {
		event := eventlogger.NewLoggedEvent("client-connected", "")
		event.Add("remote_address", conn.RemoteAddr().String())
		event.Add("local_address", conn.LocalAddr().String())
		c.logEvent(event)
	}
}

func (c *Client) startHeartbeat() {
	// Even though it is unnecessary, the heartbeat is contained in a select statement so that
	//  the ticker can be replaced each time a heartbeat is sent.
	select {
	case <-c.heartbeat.C:
		// Time to disconnect!
		c.lock.Lock()
		c.sendDisconnect(CLIENT_DISCONNECT_NO_HEARTBEAT, "Server timed out while waiting for heartbeat.", false)
		c.lock.Unlock()
	}
}

func (c *Client) receiveDisconnect(err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if !c.cleanDisconnect && !c.client.Local() {
		event := eventlogger.NewLoggedEvent("client-lost", "")
		event.Add("remote_address", c.conn.RemoteAddr().String())
		event.Add("local_address", c.conn.LocalAddr().String())
		event.Add("reason", err.Error())
		event.Send()
	}

	c.heartbeat.Stop()
	c.annihilate()
}

func (c *Client) forwardDatagram(dg Datagram) {
	c.client.SendDatagram(dg)
}

func (c *Client) handleDrop() {
	c.cleanDisconnect = true
	c.client.Close()
}

func (c *Client) ReceiveDatagram(dg Datagram) {
	c.lock.Lock()
	defer c.lock.Unlock()

	dgi := NewDatagramIterator(&dg)
	finish := make(chan bool)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(DatagramIteratorEOF); ok {
					c.sendDisconnect(CLIENT_DISCONNECT_TRUNCATED_DATAGRAM, "Datagram unexpectedly ended while iterating.", false)
				}
				finish <- true
			}
		}()

		switch c.state {
		case CLIENT_STATE_NEW:
			c.handleIntroduction(dgi)
		case CLIENT_STATE_ANONYMOUS:
			c.handleAuthentication(dgi)
		case CLIENT_STATE_ESTABLISHED:
			c.handleAuthenticated(dgi)
		}

		finish <- true
	}()

	<-finish
	if len(dgi.ReadRemainder()) != 0 {
		c.sendDisconnect(CLIENT_DISCONNECT_OVERSIZED_DATAGRAM, "Datagram contains excess datc.", true)

	}
}

func (c *Client) handleIntroduction(dgi *DatagramIterator) {
	msgType := dgi.ReadUint16()
	if msgType != CLIENT_HELLO {
		c.sendDisconnect(CLIENT_DISCONNECT_NO_HELLO, "First packet is not CLIENT_HELLO", false)
		return
	}

	hash := dgi.ReadUint32()
	version := dgi.ReadString()

	if version != c.ca.config.Version {
		c.sendDisconnect(CLIENT_DISCONNECT_BAD_VERSION,
			fmt.Sprintf("Client version mismatch: client=%s, server=%s", version, c.cc.config.Version), false)
	}

	if hash != core.Hash {
		c.sendDisconnect(CLIENT_DISCONNECT_BAD_VERSION,
			fmt.Sprintf("Client DC hash mismatch: client=0x%x, server=0x%x", hash, core.Hash), false)
	}

	resp := NewDatagram()
	resp.AddUint16(CLIENT_HELLO_RESP)
	c.client.SendDatagram(resp)

	c.state = CLIENT_STATE_ANONYMOUS
}

func (c *Client) handleAuthentication(dgi *DatagramIterator) {
	msgType := dgi.ReadUint16()
	switch msgType {
	case CLIENT_DISCONNECT:
		event := eventlogger.NewLoggedEvent("client-disconnected", "")
		event.Add("who", c.conn.RemoteAddr().String())
		event.Send()

		c.cleanDisconnect = true
		c.client.Close()
	case CLIENT_OBJECT_SET_FIELD:
		// TODO
	case CLIENT_HEARTBEAT:
		c.handleHeartbeat()
	default:
		c.sendDisconnect(CLIENT_DISCONNECT_INVALID_MSGTYPE,
			fmt.Sprintf("Message type %d not allowed prior to authentication.", msgType), true)
	}
}

func (c *Client) handleAuthenticated(dgi *DatagramIterator) {
	msgType := dgi.ReadUint16()
	switch msgType {
	case CLIENT_DISCONNECT:
		event := eventlogger.NewLoggedEvent("client-disconnected", "")
		event.Add("who", c.conn.RemoteAddr().String())
		event.Send()

		c.cleanDisconnect = true
		c.client.Close()
	case CLIENT_OBJECT_SET_FIELD:
	case CLIENT_OBJECT_LOCATION:
	case CLIENT_ADD_INTEREST:
	case CLIENT_ADD_INTEREST_MULTIPLE:
	case CLIENT_REMOVE_INTEREST:
	case CLIENT_HEARTBEAT:
		c.handleHeartbeat()
	default:
		c.sendDisconnect(CLIENT_DISCONNECT_INVALID_MSGTYPE,
			fmt.Sprintf("Message type %d not valid.", msgType), true)
	}
}

func (c *Client) handleHeartbeat() {
	c.heartbeat = time.NewTicker(time.Duration(c.config.Client.Heartbeat_Timeout) * time.Second)
}

func (c *Client) handleAddOwnership(do Doid_t, parent Doid_t, zone Zone_t, dc uint16, dgi *DatagramIterator, other bool) {
	msgType := CLIENT_ENTER_OBJECT_REQUIRED_OWNER
	if other {
		msgType = CLIENT_ENTER_OBJECT_REQUIRED_OTHER_OWNER
	}

	resp := NewDatagram()
	resp.AddUint16(uint16(msgType))
	resp.AddDoid(do)
	resp.AddLocation(parent, zone)
	resp.AddUint16(dc)
	resp.AddData(dgi.ReadRemainder())
	c.client.SendDatagram(resp)
}

func (c *Client) handleRemoveOwnership(do Doid_t) {
	resp := NewDatagram()
	resp.AddUint16(CLIENT_OBJECT_LEAVING_OWNER)
	resp.AddDoid(do)
	c.client.SendDatagram(resp)
}

func (c *Client) handleSetFields(do Doid_t, fields uint16, dgi *DatagramIterator) {
	resp := NewDatagram()
	resp.AddUint16(CLIENT_OBJECT_SET_FIELDS)
	resp.AddDoid(do)
	resp.AddUint16(fields)
	resp.AddData(dgi.ReadRemainder())
	c.client.SendDatagram(resp)
}

func (c *Client) handleSetField(do Doid_t, field uint16, dgi *DatagramIterator) {
	resp := NewDatagram()
	resp.AddUint16(CLIENT_OBJECT_SET_FIELD)
	resp.AddDoid(do)
	resp.AddUint16(field)
	resp.AddData(dgi.ReadRemainder())
	c.client.SendDatagram(resp)
}

func (c *Client) handleRemoveInterest(id uint16, context uint32) {
	resp := NewDatagram()
	resp.AddUint16(CLIENT_REMOVE_INTEREST)
	resp.AddUint32(context)
	resp.AddUint16(id)
	c.client.SendDatagram(resp)
}

func (c *Client) handleAddInterest(i Interest, context uint32) {
	msgType := uint16(CLIENT_ADD_INTEREST)
	if len(i.zones) > 0 {
		msgType = uint16(CLIENT_ADD_INTEREST_MULTIPLE)
	}

	resp := NewDatagram()
	resp.AddUint16(msgType)
	resp.AddUint32(context)
	resp.AddUint16(i.id)
	resp.AddDoid(i.parent)
	if len(i.zones) > 0 {
		resp.AddUint16(uint16(len(i.zones)))
	}
	for _, zone := range i.zones {
		resp.AddZone(zone)
	}
	c.client.SendDatagram(resp)
}

func (c *Client) handleRemoveObject(do Doid_t) {
	resp := NewDatagram()
	resp.AddUint16(CLIENT_OBJECT_LEAVING)
	resp.AddDoid(do)
	c.client.SendDatagram(resp)
}

func (c *Client) handleAddObject(do Doid_t, parent Doid_t, zone Zone_t, dc uint16, dgi *DatagramIterator, other bool) {
	msgType := CLIENT_ENTER_OBJECT_REQUIRED
	if other {
		msgType = CLIENT_ENTER_OBJECT_REQUIRED_OTHER
	}

	resp := NewDatagram()
	resp.AddUint16(uint16(msgType))
	resp.AddDoid(do)
	resp.AddLocation(parent, zone)
	resp.AddUint16(dc)
	resp.AddData(dgi.ReadRemainder())
	c.client.SendDatagram(resp)
}

func (c *Client) handleInterestDone(interestId uint16, context uint32) {
	resp := NewDatagram()
	resp.AddUint16(CLIENT_DONE_INTEREST_RESP)
	resp.AddUint32(context)
	resp.AddUint16(interestId)
	c.client.SendDatagram(resp)
}
