package clientagent

import (
	"astrongo/core"
	"astrongo/messagedirector"
	. "astrongo/util"
	"fmt"
	"github.com/apex/log"
)

type ClientState int

const (
	CLIENT_STATE_NEW ClientState = iota
	CLIENT_STATE_ANONYMOUS
	CLIENT_STATE_ESTABLISHED
)

type Client struct {
	messagedirector.MDParticipantBase

	config core.Role
	ca     *ClientAgent
	log    *log.Entry

	channel Channel_t
	state   ClientState
}

func (c *Client) init(config core.Role, ca *ClientAgent) *Client {
	client := &Client{config: config, ca: ca}

	client.channel = ca.Allocate()
	if client.channel == 0 {
		// TODO: kick for capacity
	}

	client.log = log.WithFields(log.Fields{
		"name": fmt.Sprintf("Client (%d)", client.channel),
	})

	client.SubscribeChannel(client.channel)
	client.SubscribeChannel(BCHAN_CLIENTS)

	return client
}
