package messagedirector

import (
	"astrongo/core"
	. "astrongo/test"
	. "astrongo/util"
	"github.com/apex/log"
	"os"
	"testing"
	"time"
)

var mainClient, client1, client2 *TestMDConnection

func TestMain(m *testing.M) {
	// SETUP
	// Silence the (very annoying) logger while we're testing
	log.SetHandler(log.HandlerFunc(func(*log.Entry) error { return nil }))
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

	// Send a test datagram on the first connection
	dg = (&TestDatagram{}).Create([]Channel_t{123456789}, 0, 1234)
	dg.AddUint32(0xDEADBEEF)
	client1.SendDatagram(*dg)
	client2.Expect(t, *dg, false)
	mainClient.Expect(t, *dg, false)

	// Unsubscribe on the first connection
	dg = (&TestDatagram{}).CreateRemoveChannel(123456789)
	client1.SendDatagram(*dg)
	client1.ExpectNone(t)
	client2.ExpectNone(t)
	// MD should not unsubscribe from parent
	mainClient.ExpectNone(t)

	// Send another datagram on the second connection
	dg = (&TestDatagram{}).Create([]Channel_t{123456789}, 0, 1234)
	dg.AddUint32(0xDEADBEEF)
	client2.SendDatagram(*dg)
	mainClient.Expect(t, *dg, false) // Should be sent upwards
	client2.ExpectNone(t)            // Should not be relayed
	client1.ExpectNone(t)            // Should not be echoed back

	// CLose the second connection, auto-unsubscribing it
	client2.Close()
	client2 = (&TestMDConnection{}).Connect(":57123", "client #2")
	client1.ExpectNone(t)
	// MD should unsubscribe from parent
	mainClient.Expect(t, *(&TestDatagram{}).CreateRemoveChannel(123456789), false)
}

func TestMD_Multi(t *testing.T) {
	mainClient.Flush()
	client1.Flush()
	client2.Flush()

	// Subscribe to a pair of channels on the first client
	for _, ch := range []Channel_t{1111, 2222} {
		dg := (&TestDatagram{}).CreateAddChannel(ch)
		client1.SendDatagram(*dg)
	}

	// Subscribe to a pair of channels on the second
	for _, ch := range []Channel_t{2222, 3333} {
		dg := (&TestDatagram{}).CreateAddChannel(ch)
		client2.SendDatagram(*dg)
	}

	time.Sleep(200 * time.Millisecond)
	mainClient.Flush()

	// A datagram on channel 2222 should be delivered on both clients
	dg := (&TestDatagram{}).Create([]Channel_t{2222}, 0, 1337)
	dg.AddUint32(0xDEADBEEF)
	mainClient.SendDatagram(*dg)
	client1.Expect(t, *dg, false)
	client2.Expect(t, *dg, false)

	// A datagram to channels 1111 and 3333 should be delivered to both as well
	dg = (&TestDatagram{}).Create([]Channel_t{1111, 3333}, 0, 1337)
	dg.AddUint32(0xDEADBEEF)
	mainClient.SendDatagram(*dg)
	client1.Expect(t, *dg, false)
	client2.Expect(t, *dg, false)

	// A datagram should only be delivered once if multiple channels match
	dg = (&TestDatagram{}).Create([]Channel_t{1111, 2222}, 0, 1337)
	dg.AddUint32(0xDEADBEEF)
	mainClient.SendDatagram(*dg)
	client1.Expect(t, *dg, false)
	client1.ExpectNone(t)
	client2.Expect(t, *dg, false)

	// Verify behavior for datagrams with duplicate recipients
	dg = (&TestDatagram{}).Create([]Channel_t{1111, 2222, 3333, 1111, 1111,
		2222, 3333, 3333, 2222}, 0, 1337)
	dg.AddUint32(0xDEADBEEF)
	mainClient.SendDatagram(*dg)
	client1.Expect(t, *dg, false)
	client1.ExpectNone(t)
	client2.Expect(t, *dg, false)
	client2.ExpectNone(t)

	// Send the same message through the first client; verify behavior
	client1.SendDatagram(*dg)
	mainClient.Expect(t, *dg, false)
	mainClient.ExpectNone(t)
	client2.Expect(t, *dg, false)
	client2.ExpectNone(t)
	client1.ExpectNone(t)

	// Unsubscribe from the channels
	for _, ch := range []Channel_t{1111, 2222, 3333} {
		dg := (&TestDatagram{}).CreateRemoveChannel(ch)
		client1.SendDatagram(*dg)
		client2.SendDatagram(*dg)
	}
}
