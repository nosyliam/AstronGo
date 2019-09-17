package net

import (
	"astrongo/util"
	"bytes"
	"encoding/binary"
	"errors"
	"sync"
	"time"
)

// DatagramHandler is an interface for which structures that can accept datagrams may
//  implement to accept datagrams from a client, such as an MD participant.
type DatagramHandler interface {
	HandleDatagram(util.Datagram)
	Terminate()
}

type Client struct {
	sync.Mutex
	tr      Transport
	handler DatagramHandler
	buff    bytes.Buffer
}

func NewClient(tr Transport, handler DatagramHandler) *Client {
	client := &Client{tr: tr, handler: handler}
	return client
}

func (c *Client) initialize() {
	go c.read()
}

func (c *Client) shutdown() {
	c.tr.Close()
}

func (c *Client) defragment() {
	for c.buff.Len() > util.Dgsize {
		data := c.buff.Bytes()
		sz := binary.LittleEndian.Uint32(data[0:util.Dgsize])
		if c.buff.Len() > int(sz+util.Dgsize) {
			overreadSz := c.buff.Len() - int(sz) - int(util.Dgsize)
			dg := util.NewDatagram()
			dg.Write(data[util.Dgsize : sz+util.Dgsize])
			if 0 < overreadSz {
				c.buff.Truncate(0)
				c.buff.Write(data[sz+util.Dgsize : sz+util.Dgsize+uint32(overreadSz)])
			} else {
				// No overread
				c.buff.Truncate(0)
			}

			c.Lock()
			c.handler.HandleDatagram(dg)
			c.Unlock()
		} else {
			break
		}
	}
}

func (c *Client) processInput(len int, data []byte) {
	c.Mutex.Lock()

	// Check if we have enough data for a single datagram
	if c.buff.Len() == 0 && len > util.Dgsize {
		sz := binary.LittleEndian.Uint32(data[0:util.Dgsize])
		if sz == uint32(len-util.Dgsize) {
			// We have enough data for a full datagram; send it off
			dg := util.NewDatagram()
			dg.Write(data[util.Dgsize:])
			c.handler.HandleDatagram(dg)
			c.Mutex.Unlock()
		}
	}

	c.buff.Write(data)
	c.Mutex.Unlock()
	c.defragment()
}

func (c *Client) read() {
	buff := make([]byte, 1024)
	if n, err := c.tr.Read(buff); err == nil {
		c.processInput(n, buff[0:n])
		c.read()
	} else {
		c.disconnect(err)
	}
}

func (c *Client) sendDatagram(datagram util.Datagram) {
	var dg util.Datagram
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	dg.AddSize(util.Dgsize_t(datagram.Len()))
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

func (c *Client) disconnect(err error) {
	c.handler.Terminate()
}
