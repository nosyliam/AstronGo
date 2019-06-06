// Original code derived from https://github.com/ortuman/jackal

package net

import (
	"context"
	"net"
	"sync/atomic"
	"time"
)

// Server is an interface which allows a network listening mechanism to pass accepted connections to
//  an actual server, like a CA or MD
type Server interface {
	handleConnect(net.Conn)
}

// NetworkServer is a base class which provides methods that accept connections.
type NetworkServer struct {
	Handler Server

	keepAlive time.Duration
	ln        net.Listener
	listening uint32
}

func (s *NetworkServer) Start(bindAddr string, errChan chan error) {
	if err := s.listenConn(bindAddr, errChan); err != nil {
		errChan <- err
	}
}

func (s *NetworkServer) Shutdown(ctx context.Context) error {
	if atomic.CompareAndSwapUint32(&s.listening, 1, 0) {
		if err := s.ln.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (s *NetworkServer) listenConn(address string, errChan chan error) error {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	s.ln = ln

	errChan <- nil
	atomic.StoreUint32(&s.listening, 1)
	for atomic.LoadUint32(&s.listening) == 1 {
		conn, err := ln.Accept()
		if err == nil {
			s.Handler.handleConnect(conn)
			continue
		}
	}
	return nil
}
