package net

import "astrongo/util"

// DatagramHandler is an interface for which structures that can accept datagrams may
//  implement to accept datagrams from a client, such as an MD participant.
type DatagramHandler interface {
	HandleDatagram(util.Datagram)
}

type Client struct {
	tr      Transport
	handler DatagramHandler
}

func NewClient(tr Transport) {
	client := &Client{tr: tr}

	go client.read()
}

func (c *Client) read() {

}
