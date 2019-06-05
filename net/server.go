// Original code derived from https://github.com/ortuman/jackal

package net

import (
	"context"
	"github.com/apex/log"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
)

// A server is a representation of a listening socket. Structures embedding Server must
// implement the appropriate callbacks.
type Server interface {
}

// NetworkServer is a base class which provides methods that accept and manage connections.
type NetworkServer struct {
	Server

	inConns   sync.Map
	outConns  sync.Map
	ln        net.Listener
	listening uint32
}

func (s *NetworkServer) start(bindAddr string, port int) {
	address := bindAddr + ":" + strconv.Itoa(port)

	if err := s.listenConn(address); err != nil {
		log.Fatalf("%v", err)
	}
}

func (s *NetworkServer) shutdown(ctx context.Context) error {
	if atomic.CompareAndSwapUint32(&s.listening, 1, 0) {
		if err := s.ln.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (s *NetworkServer) listenConn(address string) error {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	s.ln = ln

	atomic.StoreUint32(&s.listening, 1)
	for atomic.LoadUint32(&s.listening) == 1 {
		_, err := ln.Accept()
		if err == nil {
			// TODO
			continue
		}
	}
	return nil
}
