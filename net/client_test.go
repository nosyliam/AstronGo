package net

import (
	"astrongo/util"
	"bufio"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
	"time"
)

type MDParticipantFake struct{}

var queue = make(chan util.Datagram)

func (m *MDParticipantFake) RouteDatagram(datagram util.Datagram) {
	queue <- datagram
}

func (m *MDParticipantFake) HandleDatagram(datagram util.Datagram) {
	queue <- datagram
}

func (m *MDParticipantFake) Terminate() {}

var participant *MDParticipantFake
var netclient *Client

func TestClient_SendDatagram(t *testing.T) {
	dg := util.NewDatagram()
	dg.WriteString("hello")

	go netclient.sendDatagram(dg)
	reader := bufio.NewReaderSize(server, socketBuffSize)
	buff := make([]byte, 1024)
	data, err := reader.Read(buff)
	if err != nil {
		t.Error(err)
	}

	require.EqualValues(t, data, []byte{5, 0, 0, 0, 'h', 'e', 'l', 'l', 'o'})
}

func init() {
	server, client = net.Pipe()
	socket = &socketTransport{
		conn:      client,
		rw:        client,
		br:        bufio.NewReaderSize(client, socketBuffSize),
		bw:        bufio.NewWriterSize(client, socketBuffSize),
		keepAlive: 50 * time.Millisecond,
	}
	netclient = NewClient(Transport(socket), participant)
}
