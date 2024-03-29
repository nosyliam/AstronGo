package net

import (
	. "astrongo/util"
	"bytes"
	"encoding/binary"
	"errors"
	gonet "net"
	"sync"
	"time"
)

const BUFF_SIZE = 4096

// DatagramHandler is an interface for which structures that can accept datagrams may
//  implement to accept datagrams from a client, such as an MD participant.
type DatagramHandler interface {
	// Handles a message received from the client
	ReceiveDatagram(Datagram)
	// Handles a message received from the MD
	HandleDatagram(Datagram, *DatagramIterator)

	Terminate(error)
}

type Client struct {
	sync.Mutex
	tr      Transport
	handler DatagramHandler
	buff    bytes.Buffer
	timeout time.Duration

	remote *gonet.TCPAddr
	local  *gonet.TCPAddr
}

func NewClient(tr Transport, handler DatagramHandler, timeout time.Duration) *Client {
	client := &Client{
		tr:      tr,
		handler: handler,
		remote:  tr.Conn().RemoteAddr().(*gonet.TCPAddr),
		local:   tr.Conn().LocalAddr().(*gonet.TCPAddr),
	}
	client.initialize()
	client.timeout = timeout
	return client
}

func (c *Client) initialize() {
	go c.read()
}

func (c *Client) shutdown() {
	c.tr.Close()
}

func (c *Client) defragment() {
	for c.buff.Len() > Dgsize {
		data := c.buff.Bytes()
		sz := binary.LittleEndian.Uint32(data[0:Dgsize])
		if c.buff.Len() >= int(sz+Dgsize) {
			overreadSz := c.buff.Len() - int(sz) - int(Dgsize)
			dg := NewDatagram()
			dg.Write(data[Dgsize : sz+Dgsize])
			if 0 < overreadSz {
				c.buff.Truncate(0)
				c.buff.Write(data[sz+Dgsize : sz+Dgsize+uint32(overreadSz)])
			} else {
				// No overread
				c.buff.Truncate(0)
			}

			go c.handler.ReceiveDatagram(dg)
		} else {
			return
		}
	}
}

func (c *Client) processInput(len int, data []byte) {
	c.Lock()

	// Check if we have enough data for a single datagram
	if c.buff.Len() == 0 && len >= Dgsize {
		sz := binary.LittleEndian.Uint32(data[0:Dgsize])
		if sz == uint32(len-Dgsize) {
			// We have enough data for a full datagram; send it off
			dg := NewDatagram()
			dg.Write(data[Dgsize:])
			go c.handler.ReceiveDatagram(dg)
			c.Unlock()
			return
		}
	}

	c.buff.Write(data)
	c.defragment()
	c.Unlock()

}

func (c *Client) read() {
	buff := make([]byte, BUFF_SIZE)
	if n, err := c.tr.Read(buff); err == nil {
		go c.processInput(n, buff[0:n])
		c.read()
	} else {
		c.disconnect(err)
	}
}

func (c *Client) SendDatagram(datagram Datagram) {
	var dg Datagram
	dg = NewDatagram()

	c.Lock()

	dg.AddSize(Dgsize_t(datagram.Len()))
	dg.Write(datagram.Bytes())

	if _, err := c.tr.WriteDatagram(dg); err != nil {
		c.disconnect(err)
	}

	select {
	case err := <-c.tr.Flush():
		if err != nil {
			c.disconnect(err)
		}
	case <-time.After(c.timeout):
		c.disconnect(errors.New("write timeout"))
	}

	c.Unlock()
}

func (c *Client) Close() {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.tr.Close()
}

func (c *Client) disconnect(err error) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.tr.Close()
	c.handler.Terminate(err)
}

func (c *Client) Local() bool {
	return true
}

func (c *Client) Connected() bool {
	return !c.tr.Closed()
}

func (c *Client) RemoteIP() string {
	return c.remote.IP.String()
}

func (c *Client) RemotePort() uint16 {
	return uint16(c.remote.Port)
}

func (c *Client) LocalIP() string {
	return c.local.IP.String()
}

func (c *Client) LocalPort() uint16 {
	return uint16(c.local.Port)
}
