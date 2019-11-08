package messagedirector

import (
	"astrongo/core"
	"astrongo/net"
	. "astrongo/util"
	"fmt"
	"github.com/apex/log"
	gonet "net"
)

var MDLog *log.Entry

var MD *MessageDirector

type MessageDirector struct {
	net.Server
	net.NetworkServer

	// Connections within the context of the MessageDirector are represented as
	// participants; however, clients and objects on the SS may function as participants
	// as well. The MD will keep track of them and what channels they subscribe and route data to them.
	participants []MDParticipant

	// MD participants may directly queue datagarams to be routed by inserting it into the
	// queue channel, where they will be processed asynchronously
	Queue chan Datagram
}

func init() {
	MDLog = log.WithFields(log.Fields{
		"name": "MD",
	})
}

func Start() {
	MD = &MessageDirector{}
	MD.Queue = make(chan Datagram)
	MD.participants = make([]MDParticipant, 0)
	MD.Handler = MD.Server

	bindAddr := core.Config.MessageDirector.Bind
	if bindAddr == "" {
		bindAddr = "127.0.0.1:7199"
	}

	errChan := make(chan error)
	go func() {
		err := <-errChan
		switch err {
		case nil:
			MDLog.Info(fmt.Sprintf("Opened listening socket at %s", bindAddr))
		default:
			MDLog.Fatal(err.Error())
		}
	}()
	MD.Start(bindAddr, errChan)
	go MD.queueLoop()
}

func (m *MessageDirector) queueLoop() {
	finish := make(chan bool)
	for dg := range MD.Queue {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					if _, ok := r.(DatagramIteratorEOF); ok {
						MDLog.Error("MD received truncated datagram header from an unknown participant")
					}
					finish <- true
				}
			}()

			var receivers []Channel_t
			dgi := NewDatagramIterator(&dg)
			chanCount := dgi.ReadUint8()
			for n := 0; uint8(n) < chanCount; n++ {
				receivers = append(receivers, dgi.ReadChannel())
			}

			// Send out datagram to every receiver
			for _, recv := range receivers {
				channelMap.Channel(recv) <- dg
			}
		}()
		<-finish
	}
}

func (m *MessageDirector) handleConnect(conn gonet.Conn) {
	NewMDParticipant(conn)
}

func (m *MessageDirector) PreroutePostRemove(sender Channel_t, dg Datagram) {
	// TODO
}

func (m *MessageDirector) RecallPostRemoves(sender Channel_t) {
	// TODO
}
