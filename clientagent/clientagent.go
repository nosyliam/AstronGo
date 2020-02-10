package clientagent

import (
	"astrongo/core"
	"astrongo/messagedirector"
	"astrongo/net"
	. "astrongo/util"
	"fmt"
	"github.com/apex/log"
	gonet "net"
)

type ChannelTracker struct {
	next Channel_t
	max  Channel_t
	free []Channel_t
	log  *log.Entry
}

type ClientAgent struct {
	net.Server

	tracker *ChannelTracker
	config  core.Role
	log     *log.Entry

	rng             messagedirector.Range
	interestTimeout int
}

func NewChannelTracker(min Channel_t, max Channel_t, log *log.Entry) *ChannelTracker {
	return &ChannelTracker{next: min, max: max}
}

func (c *ChannelTracker) alloc() Channel_t {
	var ch Channel_t
	if c.next <= c.max {
		c.next++
		return c.next
	} else if len(c.free) != 0 {
		ch, c.free = c.free[0], c.free[1:]
		return ch
	} else {
		c.log.Fatalf("CA has no more available channels.")
	}
}

func NewClientAgent(config core.Role) *ClientAgent {
	ca := &ClientAgent{
		config: config,
		log: log.WithFields(log.Fields{
			"name": fmt.Sprintf("ClientAgent (%s)", config.Bind),
		}),
	}
	ca.tracker = NewChannelTracker(Channel_t(config.Channels.Min), Channel_t(config.Channels.Max), ca.log)

	ca.rng = messagedirector.Range{Channel_t(config.Channels.Min), Channel_t(config.Channels.Max)}
	if ca.rng.Size() <= 0 {
		ca.log.Fatal("Failed to instantiate CA: invalid channel range")
		return nil
	}

	ca.interestTimeout = config.Tuning.Interest_Timeout

	server := &net.NetworkServer{Handler: ca}
	errChan := make(chan error)
	go func() {
		err := <-errChan
		switch err {
		case nil:
			ca.log.Debugf("Opened listening socket at %s", config.Bind)
		default:
			ca.log.Fatal(err.Error())
		}
	}()

	server.Start(config.Bind, errChan)
	return ca
}

func (c *ClientAgent) HandleConnect(conn gonet.Conn) {
	// NOTE: AstronGo will not support multiple client types.
	c.log.Debugf("Incoming connection from %s", conn.RemoteAddr())
	NewAstronClient(c.config, c, conn)
}

func (c *ClientAgent) Allocate() Channel_t {
	return c.tracker.alloc()
}
