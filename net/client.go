package net

import (
	. "astrongo/util"
	"bytes"
	"encoding/binary"
	"errors"
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
}

func NewClient(tr Transport, handler DatagramHandler) *Client {
	client := &Client{tr: tr, handler: handler}
	client.initialize()
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
		if c.buff.Len() > int(sz+Dgsize) {
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

			c.Lock()
			go c.handler.ReceiveDatagram(dg)
			c.Unlock()
		} else {
			break
		}
	}
}

func (c *Client) processInput(len int, data []byte) {
	c.Mutex.Lock()

	// Check if we have enough data for a single datagram
	if c.buff.Len() == 0 && len > Dgsize {
		sz := binary.LittleEndian.Uint32(data[0:Dgsize])
		if sz == uint32(len-Dgsize) {
			// We have enough data for a full datagram; send it off
			dg := NewDatagram()
			dg.Write(data[Dgsize:])
			go c.handler.ReceiveDatagram(dg)
			c.Mutex.Unlock()
			return
		}
	}

	c.buff.Write(data)
	c.Mutex.Unlock()
	c.defragment()
}

func (c *Client) read() {
	buff := make([]byte, BUFF_SIZE)
	if n, err := c.tr.Read(buff); err == nil {
		c.processInput(n, buff[0:n])
		c.read()
	} else {
		c.disconnect(err)
	}
}

func (c *Client) SendDatagram(datagram Datagram) {
	var dg Datagram
	dg = NewDatagram()

	c.Mutex.Lock()
	defer c.Mutex.Unlock()

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
	case <-time.After(60 * time.Second):
		c.disconnect(errors.New("write timeout"))
	}

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
