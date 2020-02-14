package clientagent

import (
	"astrongo/core"
	"astrongo/dclass/dc"
	"astrongo/dclass/parse"
	"astrongo/messagedirector"
	"astrongo/net"
	. "astrongo/util"
	"errors"
	"fmt"
	gonet "net"
	"testing"
	"time"
)

var msgQueue chan Datagram

type MDParticipantFake struct {
	messagedirector.MDParticipantBase
}

func (m *MDParticipantFake) ReceiveDatagram(datagram Datagram) {
	msgQueue <- datagram
}

func (m *MDParticipantFake) HandleDatagram(datagram Datagram, dgi *DatagramIterator) {
	msgQueue <- datagram
}

func (m *MDParticipantFake) Terminate(error) {}

func connectClient() (*net.Client, error) {
	conn, err := gonet.Dial("tcp", "127.0.0.1:7170")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to connect to CA: %s", err))
	}

	participant := &MDParticipantFake{}
	socket := net.NewSocketTransport(conn, 60*time.Second, 4096)
	client := net.NewClient(socket, participant)
	participant.Init()
	return client, nil
}

func TestAstronClient_Heartbeat(t *testing.T) {

}

func init() {
	if messagedirector.MD != nil {
		core.Config = &core.ServerConfig{MessageDirector: struct {
			Bind    string
			Connect string
		}{Bind: "127.0.0.1:7199", Connect: ""}}
		messagedirector.Start()
	}

	dct, err := parse.ParseFile("dclass/parse/test.dc")
	if err != nil {
		panic(fmt.Sprintf("test dclass parse failed: %s", err))
	}

	dcf := dct.Traverse()
	hashgen := dc.NewHashGenerator()
	dcf.GenerateHash(hashgen)
	core.Hash = hashgen.Hash()
	core.DC = dcf
}
