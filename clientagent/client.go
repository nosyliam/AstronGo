package clientagent

import (
	"astrongo/core"
	"astrongo/messagedirector"
)

type Client struct {
	messagedirector.MDParticipantBase

	config core.Role
	ca     *ClientAgent
}

func (c *Client) init(config core.Role, ca *ClientAgent) *Client {
	client := &Client{config: config, ca: ca}
	return client
}
