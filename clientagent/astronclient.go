package clientagent

import (
	"astrongo/core"
	"astrongo/net"
	gonet "net"
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

	cleanDisconnect  bool
	allowedInterests InterestPermission
}

func NewAstronClient(config core.Role, ca *ClientAgent, conn gonet.Conn) *AstronClient {
	client := &AstronClient{config: config}
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

	return client
}
