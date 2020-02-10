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

	client *net.Client
	config core.Role
	log    *log.Entry
	lock   sync.Mutex

	cleanDisconnect  bool
	allowedInterests InterestPermission
}

func NewAstronClient(config core.Role, ca *ClientAgent, conn gonet.Conn) *AstronClient {
	client := &AstronClient{
		config: config,
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
		// TODO: Implement heartbeat
	}

	if !client.client.Local() {
		event := eventlogger.NewLoggedEvent("client-connected", "AstronClient")
		event.Add("remote_address", conn.RemoteAddr().String())
		event.Add("local_address", conn.LocalAddr().String())
		event.Send()
	}

	return client
}

func (a *AstronClient) sendDisconnect(reason uint16, error string, security bool) {
	if a.client.Connected() {
		a.client.Close()
	}
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

	}
}

func (a *AstronClient) handleIntroduction(dgi *DatagramIterator) {

}

func (a *AstronClient) handleAuthentication(dgi *DatagramIterator) {

}

func (a *AstronClient) handleAuthenticated(dgi *DatagramIterator) {

}
