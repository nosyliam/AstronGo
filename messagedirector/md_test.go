package messagedirector

import (
	"astrongo/core"
	. "astrongo/test"
	. "astrongo/util"
	"os"
	"testing"
)

var mainClient, client1, client2 *TestMDConnection

func TestMain(m *testing.M) {
	// SETUP
	upstream := StartUpstream("127.0.0.1:57124")

	StartDaemon(
		core.ServerConfig{MessageDirector: struct {
			Bind    string
			Connect string
		}{Bind: "127.0.0.1:57123", Connect: "127.0.0.1:57124"}})
	Start()

	for {
		if upstream.Server != nil {
			break
		}
	}
	mainClient = (&TestMDConnection{}).Set(*upstream.Server, "main")

	client1 = (&TestMDConnection{}).Connect(":57123", "client #1")
	client2 = (&TestMDConnection{}).Connect(":57123", "client #2")

	code := m.Run()

	// TEARDOWN
	mainClient.Close()
	client1.Close()
	client2.Close()
	os.Exit(code)
}

func TestMD_Single(t *testing.T) {
	mainClient.Flush()

	// Send a datagram
	dg := (&TestDatagram{}).Create([]Channel_t{1234}, 4321, 1337)
	dg.AddString("HELLO #2!")
	client1.SendDatagram(*dg)

	// The MD should pass it to the main client
	mainClient.Expect(t, *dg, false)
}

func TestMD_Subscribe(t *testing.T) {
	mainClient.Flush()
	client1.Flush()
	client2.Flush()

	// Subscribe to a channel
	dg := (&TestDatagram{}).CreateAddChannel(123456789)
	client1.SendDatagram(*dg)
	client1.ExpectNone(t)
	mainClient.Expect(t, *dg, false)

	// Send a test datagram on the other connection
	dg = (&TestDatagram{}).Create([]Channel_t{123456789}, 0, 1234)
	dg.AddUint32(0xDEADBEEF)
	client2.SendDatagram(*dg)
	client1.Expect(t, *dg, false)
	// MD should relay the message upwards
	mainClient.Expect(t, *dg, false)

	// Subscribe on the other connection
	dg = (&TestDatagram{}).CreateAddChannel(123456789)
	client2.SendDatagram(*dg)
	client2.ExpectNone(t)
	// MD should not ask for the channel again
	mainClient.ExpectNone(t)

}
