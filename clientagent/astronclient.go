package clientagent

import (
	"astrongo/core"
	"astrongo/eventlogger"
	"astrongo/net"
	. "astrongo/util"
	"fmt"
	"github.com/apex/log"
	gonet "net"
	"sync"
	"time"
)

type InterestPermission int

const (
	INTERESTS_ENABLED InterestPermission = iota
	INTERESTS_VISIBLE
	INTERESTS_DISABLED
)

// AstronClient provides backing functions for the more backend-centric Client class to use.
type AstronClient struct {
	Client

	conn   gonet.Conn
	client *net.Client
	config core.Role
	log    *log.Entry
	lock   sync.Mutex

	cleanDisconnect  bool
	allowedInterests InterestPermission
	heartbeat        *time.Ticker
	finish           chan bool
}

func NewAstronClient(config core.Role, ca *ClientAgent, conn gonet.Conn) *AstronClient {
	client := &AstronClient{
		config: config,
		conn:   conn,
		log: log.WithFields(log.Fields{
			"name": fmt.Sprintf("Client (%s)", config.Bind),
		}),
	}
	client.init(config, ca)
	switch config.Client.Add_Interest {
	case "enabled":
		client.allowedInterests = INTERESTS_ENABLED
	case "visible":
		client.allowedInterests = INTERESTS_VISIBLE
	default:
		client.allowedInterests = INTERESTS_DISABLED
	}

	socket := net.NewSocketTransport(conn,
		time.Duration(config.Client.Keepalive)*time.Second, config.Client.Write_Buffer_Size)
	client.client = net.NewClient(socket, client)

	if config.Client.Heartbeat_Timeout != 0 {
		client.heartbeat = time.NewTicker(time.Duration(config.Client.Heartbeat_Timeout) * time.Second)
		go client.startHeartbeat()
	}

	if !client.client.Local() {
		event := eventlogger.NewLoggedEvent("client-connected", "")
		event.Add("remote_address", conn.RemoteAddr().String())
		event.Add("local_address", conn.LocalAddr().String())
		event.Send()
	}

	return client
}

func (a *AstronClient) startHeartbeat() {
	// Even though it is unnecessary, the heartbeat is contained in a select statement so that
	//  the ticker can be replaced each time a heartbeat is sent.
	select {
	case <-a.heartbeat.C:
		// Time to disconnect!
		a.lock.Lock()
		a.sendDisconnect(CLIENT_DISCONNECT_NO_HEARTBEAT, "Server timed out while waiting for heartbeat.", false)
		a.lock.Unlock()
	}
}

func (a *AstronClient) sendDisconnect(reason uint16, error string, security bool) {
	if a.client.Connected() {
		a.Client.sendDisconnect(reason, error, security)

		resp := NewDatagram()
		resp.AddUint16(CLIENT_EJECT)
		resp.AddUint16(reason)
		resp.AddString(error)
		a.client.SendDatagram(resp)

		a.cleanDisconnect = true
		a.client.Close()
	}
}

func (a *AstronClient) receiveDisconnect(err error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	if !a.cleanDisconnect && !a.client.Local() {
		event := eventlogger.NewLoggedEvent("client-lost", "")
		event.Add("remote_address", a.conn.RemoteAddr().String())
		event.Add("local_address", a.conn.LocalAddr().String())
		event.Add("reason", err.Error())
		event.Send()
	}

	a.heartbeat.Stop()
}

func (a *AstronClient) forwardDatagram(dg Datagram) {
	a.client.SendDatagram(dg)
}

func (a *AstronClient) handleDrop() {
	a.cleanDisconnect = true
	a.client.Close()
}

func (a *AstronClient) ReceiveDatagram(dg Datagram) {
	a.lock.Lock()
	defer a.lock.Unlock()

	dgi := NewDatagramIterator(&dg)
	finish := make(chan bool)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(DatagramIteratorEOF); ok {
				}
				finish <- true
			}
		}()

		switch a.state {
		case CLIENT_STATE_NEW:
			a.handleIntroduction(dgi)
		case CLIENT_STATE_ANONYMOUS:
			a.handleAuthentication(dgi)
		case CLIENT_STATE_ESTABLISHED:
			a.handleAuthenticated(dgi)
		}

		finish <- true
	}()

	<-finish
	if len(dgi.ReadRemainder()) != 0 {
		a.sendDisconnect(CLIENT_DISCONNECT_OVERSIZED_DATAGRAM, "Datagram contains excess data.", true)

	}
}

func (a *AstronClient) handleIntroduction(dgi *DatagramIterator) {
	msgType := dgi.ReadUint16()
	if msgType != CLIENT_HELLO {
		a.sendDisconnect(CLIENT_DISCONNECT_NO_HELLO, "First packet is not CLIENT_HELLO", false)
		return
	}

	hash := dgi.ReadUint32()
	version := dgi.ReadString()

	if version != a.ca.config.Version {
		a.sendDisconnect(CLIENT_DISCONNECT_BAD_VERSION,
			fmt.Sprintf("Client version mismatch: client=%s, server=%s", version, a.ca.config.Version), false)
	}

	if hash != core.Hash {
		a.sendDisconnect(CLIENT_DISCONNECT_BAD_VERSION,
			fmt.Sprintf("Client DC hash mismatch: client=0x%x, server=0x%x", hash, core.Hash), false)
	}

	resp := NewDatagram()
	resp.AddUint16(CLIENT_HELLO_RESP)
	a.client.SendDatagram(resp)

	a.state = CLIENT_STATE_ANONYMOUS
}

func (a *AstronClient) handleAuthentication(dgi *DatagramIterator) {
	msgType := dgi.ReadUint16()
	switch msgType {
	case CLIENT_DISCONNECT:
		event := eventlogger.NewLoggedEvent("client-disconnected", "")
		event.Add("who", a.conn.RemoteAddr().String())
		event.Send()

		a.cleanDisconnect = true
		a.client.Close()
	case CLIENT_OBJECT_SET_FIELD:
		// TODO
	case CLIENT_HEARTBEAT:
		a.handleHeartbeat()
	default:
		a.sendDisconnect(CLIENT_DISCONNECT_INVALID_MSGTYPE,
			fmt.Sprintf("Message type %d not allowed prior to authentication.", msgType), true)
	}
}

func (a *AstronClient) handleAuthenticated(dgi *DatagramIterator) {
	msgType := dgi.ReadUint16()
	switch msgType {
	case CLIENT_DISCONNECT:
		event := eventlogger.NewLoggedEvent("client-disconnected", "")
		event.Add("who", a.conn.RemoteAddr().String())
		event.Send()

		a.cleanDisconnect = true
		a.client.Close()
	case CLIENT_OBJECT_SET_FIELD:
	case CLIENT_OBJECT_LOCATION:
	case CLIENT_ADD_INTEREST:
	case CLIENT_ADD_INTEREST_MULTIPLE:
	case CLIENT_REMOVE_INTEREST:
	case CLIENT_HEARTBEAT:
		a.handleHeartbeat()
	default:
		a.sendDisconnect(CLIENT_DISCONNECT_INVALID_MSGTYPE,
			fmt.Sprintf("Message type %d not valid.", msgType), true)
	}
}

func (a *AstronClient) handleHeartbeat() {
	a.heartbeat = time.NewTicker(time.Duration(a.config.Client.Heartbeat_Timeout) * time.Second)
}
