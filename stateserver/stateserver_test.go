package stateserver

import (
	"astrongo/core"
	"astrongo/messagedirector"
	. "astrongo/test"
	. "astrongo/util"
	"encoding/hex"
	"fmt"
	"github.com/apex/log"
	"github.com/tj/assert"
	"os"
	"testing"
	"time"
)

func connect(ch Channel_t) *TestChannelConnection {
	conn := (&TestChannelConnection{}).Create("127.0.0.1:57123", fmt.Sprintf("Channel (%d)", ch), ch)
	conn.Timeout = 200
	return conn
}

func appendMeta(dg *Datagram, doid Doid_t, parent Doid_t, zone Zone_t, dclass uint16) {
	if doid != 6969 {
		dg.AddDoid(doid)
	}
	if parent != 6969 {
		dg.AddDoid(parent)
	}
	if zone != 6969 {
		dg.AddZone(zone)
	}
	if dclass != 6969 {
		dg.AddUint16(dclass)
	}
}

func instantiateObject(conn *TestChannelConnection, sender Channel_t, doid Doid_t, parent Doid_t, zone Zone_t, required uint16) {
	dg := (&TestDatagram{}).Create([]Channel_t{100100}, sender, STATESERVER_CREATE_OBJECT_WITH_REQUIRED)
	appendMeta(dg, doid, parent, zone, DistributedTestObject1)
	dg.AddUint16(required)
	conn.SendDatagram(*dg)
}

func deleteObject(conn *TestChannelConnection, sender Channel_t, doid Doid_t) {
	dg := (&TestDatagram{}).Create([]Channel_t{Channel_t(doid)}, sender, STATESERVER_OBJECT_DELETE_RAM)
	dg.AddDoid(doid)
	conn.SendDatagram(*dg)
}

func TestMain(m *testing.M) {
	// SETUP
	// Silence the (very annoying) logger while we're testing
	log.SetHandler(log.HandlerFunc(func(*log.Entry) error { return nil }))

	StartDaemon(
		core.ServerConfig{MessageDirector: struct {
			Bind    string
			Connect string
		}{Bind: "127.0.0.1:57123"},
			General: struct {
				Eventlogger string
				DC_Files    []string
			}{Eventlogger: "", DC_Files: []string{"dclass/parse/test.dc"}}})
	if err := core.LoadDC(); err != nil {
		os.Exit(1)
	}
	messagedirector.Start()
	NewStateServer(core.Role{Control: 100100})

	code := m.Run()

	// TEARDOWN
	os.Exit(code)
}

func TestStateServer_CreateDelete(t *testing.T) {
	ai, parent, children :=
		connect(LocationAsChannel(5000, 1500)),
		connect(5000),
		connect(ParentToChildren(101000000))
	children.Timeout = 400

	test := func() {
		instantiateObject(ai, 5, 101000000, 5000, 1500, 1337)

		var received bool
		for n := 0; n < 2; n++ {
			dg := parent.ReceiveMaybe()
			assert.True(t, dg != nil, "Parent did not receive ChangingLocation and/or GetAI")
			dgi := (&TestDatagram{}).Set(dg)
			msgType := dgi.MessageType()
			// Object should tell the parent that it's arriving
			if ok, why := dgi.MatchesHeader([]Channel_t{5000}, 5, STATESERVER_OBJECT_CHANGING_LOCATION, -1); ok {
				received = true
				dgi.SeekPayload()
				dgi.ReadChannel()                                  // Sender
				dgi.ReadUint16()                                   // Message type
				assert.Equal(t, dgi.ReadDoid(), Doid_t(101000000)) // ID
				assert.Equal(t, dgi.ReadDoid(), Doid_t(5000))      // New Parent
				assert.Equal(t, dgi.ReadZone(), Zone_t(1500))      // New Zone
				assert.Equal(t, dgi.ReadDoid(), INVALID_DOID)      // Old Parent
				assert.Equal(t, dgi.ReadZone(), INVALID_ZONE)      // Old Zone
				// Object should also ask for its AI, which we are not testing here
			} else if ok, why = dgi.MatchesHeader([]Channel_t{5000}, 101000000, STATESERVER_OBJECT_GET_AI, -1); ok {
				continue
			} else {
				t.Errorf("Failed object creation test! (msgtype=%d): %s\nDatagram dump:\n%s", msgType, why, hex.Dump(dgi.Dg.Bytes()))
			}
		}

		assert.True(t, received)

		// Object should announce its entry to the zone
		dg := (&TestDatagram{}).Create([]Channel_t{LocationAsChannel(5000, 1500)},
			101000000, STATESERVER_OBJECT_ENTER_LOCATION_WITH_REQUIRED)
		appendMeta(dg, 101000000, 5000, 1500, DistributedTestObject1)
		dg.AddUint32(1337)
		ai.Expect(t, *dg, false)

		dg = (&TestDatagram{}).Create([]Channel_t{ParentToChildren(101000000)},
			101000000, STATESERVER_OBJECT_GET_LOCATION)
		dg.AddUint32(STATESERVER_CONTEXT_WAKE_CHILDREN)
		children.Expect(t, *dg, false)

		// Test for DeleteRam
		deleteObject(ai, 5, 101000000)

		// Object should inform the parent that it is leaving
		dg = (&TestDatagram{}).Create([]Channel_t{5000}, 5, STATESERVER_OBJECT_CHANGING_LOCATION)
		dg.AddDoid(101000000)
		appendMeta(dg, 6969, INVALID_DOID, INVALID_ZONE, 6969) // New Location
		appendMeta(dg, 6969, 5000, 1500, 6969)                 // Old Location
		parent.Expect(t, *dg, false)

		// Object should announce it's disappearance
		dg = (&TestDatagram{}).Create([]Channel_t{LocationAsChannel(5000, 1500)},
			5, STATESERVER_OBJECT_DELETE_RAM)
		dg.AddDoid(101000000)
		ai.Expect(t, *dg, false)
	}

	// Run tests twice to verify that ids can be reused
	test()
	test()

	// Cleanup
	parent.Close()
	ai.Close()
	children.Close()
}

func TestStateServer_Broadcast(t *testing.T) {
	ai := connect(LocationAsChannel(5000, 1500))

	// Broadcast for location test
	// Start with the creation of a DTO2
	dg := (&TestDatagram{}).Create([]Channel_t{100100}, 5, STATESERVER_CREATE_OBJECT_WITH_REQUIRED)
	appendMeta(dg, 101000001, 5000, 1500, DistributedTestObject2)
	ai.SendDatagram(*dg)

	// Ignore the entry message
	time.Sleep(50 * time.Millisecond)
	ai.Flush()

	// Send an update to setB2
	dg = (&TestDatagram{}).Create([]Channel_t{101000001}, 5, STATESERVER_OBJECT_SET_FIELD)
	dg.AddDoid(101000001)
	dg.AddUint16(SetB2)
	dg.AddUint32(0xDEADBEEF)
	ai.SendDatagram(*dg)

	// Object should broadcast the update
	dg = (&TestDatagram{}).Create([]Channel_t{LocationAsChannel(5000, 1500)}, 5, STATESERVER_OBJECT_SET_FIELD)
	dg.AddDoid(101000001)
	dg.AddUint16(SetB2)
	dg.AddUint32(0xDEADBEEF)
	ai.Expect(t, *dg, false)

	// Cleanup
	deleteObject(ai, 5, 101000001)
	ai.Close()
}

func TestStateServer_Airecv(t *testing.T) {
	conn := connect(5)
	conn.AddChannel(1300)

	instantiateObject(conn, 5, 101000002, 5000, 1500, 1337)
	time.Sleep(100 * time.Millisecond)

	dg := (&TestDatagram{}).Create([]Channel_t{101000002}, 5, STATESERVER_OBJECT_SET_AI)
	dg.AddChannel(1300)
	conn.SendDatagram(*dg)
	time.Sleep(100 * time.Millisecond)
	conn.Flush()

	dg = (&TestDatagram{}).Create([]Channel_t{101000002}, 5, STATESERVER_OBJECT_SET_FIELD)
	dg.AddDoid(101000002)
	dg.AddUint16(SetBA1)
	dg.AddUint16(0xF00D)
	conn.SendDatagram(*dg)

	// AI should receive the message
	dg = (&TestDatagram{}).Create([]Channel_t{LocationAsChannel(5000, 1500), 1300}, 5, STATESERVER_OBJECT_SET_FIELD)
	dg.AddDoid(101000002)
	dg.AddUint16(SetBA1)
	dg.AddUint16(0xF00D)
	conn.Expect(t, *dg, false)

	// AI should not get its own reflected messages back
	dg = (&TestDatagram{}).Create([]Channel_t{101000002}, 1300, STATESERVER_OBJECT_SET_FIELD)
	dg.AddDoid(101000002)
	dg.AddUint16(SetBA1)
	dg.AddUint16(0xF00D)
	conn.SendDatagram(*dg)

	conn.ExpectNone(t)

	// Test for AI notification of object deletion
	deleteObject(conn, 5, 101000002)

	// AI should receive the delete
	dg = (&TestDatagram{}).Create([]Channel_t{LocationAsChannel(5000, 1500), 1300}, 5, STATESERVER_OBJECT_DELETE_RAM)
	dg.AddDoid(101000002)
	conn.Expect(t, *dg, false)

	conn.Close()
}

func TestStateServer_SetAI(t *testing.T) {
	conn := connect(5)
	conn.AddChannel(0)

	ai1Chan, ai2Chan := Channel_t(1000), Channel_t(2000)
	ai1, ai2 := connect(ai1Chan), connect(ai2Chan)

	do1, do2 := Channel_t(133337), Channel_t(133338)
	obj1, obj2 := connect(do1), connect(do2)
	children1, children2 := connect(ParentToChildren(Doid_t(do1))), connect(ParentToChildren(Doid_t(do2)))

	// Test for an object without children, AI, or optional fields
	instantiateObject(conn, 5, Doid_t(do1), 0, 0, 1337)
	conn.ExpectNone(t)
	time.Sleep(50 * time.Millisecond)
	children1.Flush()

	// Give DO #1 to AI #1
	dg := (&TestDatagram{}).Create([]Channel_t{do1}, 5, STATESERVER_OBJECT_SET_AI)
	dg.AddChannel(ai1Chan)
	conn.SendDatagram(*dg)
	time.Sleep(50 * time.Millisecond)
	obj1.Flush()

	// DO #1 should announce its presence to AI #1
	dg = (&TestDatagram{}).Create([]Channel_t{ai1Chan}, do1, STATESERVER_OBJECT_ENTER_AI_WITH_REQUIRED)
	appendMeta(dg, Doid_t(do1), 0, 0, DistributedTestObject1)
	dg.AddUint32(1337)
	ai1.Expect(t, *dg, false)
	children1.ExpectNone(t)

	// Test for an object with an existing AI
	// Give DO #1 to AI #2
	dg = (&TestDatagram{}).Create([]Channel_t{do1}, 5, STATESERVER_OBJECT_SET_AI)
	dg.AddChannel(ai2Chan)
	conn.SendDatagram(*dg)
	time.Sleep(50 * time.Millisecond)
	obj1.Flush()

	// DO #1 should tell AI #1 that it is changing AI
	dg = (&TestDatagram{}).Create([]Channel_t{ai1Chan}, 5, STATESERVER_OBJECT_CHANGING_AI)
	dg.AddDoid(Doid_t(do1)) // ID
	dg.AddChannel(ai2Chan)  // New AI
	dg.AddChannel(ai1Chan)  // Old AI
	ai1.Expect(t, *dg, false)

	// It should also inform AI #2 that it is entering
	dg = (&TestDatagram{}).Create([]Channel_t{ai2Chan}, do1, STATESERVER_OBJECT_ENTER_AI_WITH_REQUIRED)
	appendMeta(dg, Doid_t(do1), 0, 0, DistributedTestObject1)
	dg.AddUint32(1337)
	ai2.Expect(t, *dg, false)

	// Test for child AI handling on creation
	// Instantiate a new object beneath DO #1
	instantiateObject(conn, 5, Doid_t(do2), Doid_t(do1), 1500, 1337)

	// DO #1 should receive two messages from the child
	var context uint32
	for n := 0; n < 2; n++ {
		dg = obj1.ReceiveMaybe()
		assert.True(t, dg != nil, "Parent did not receive ChangingLocation and/or GetAI")
		dgi := (&TestDatagram{}).Set(dg)
		if ok, _ := dgi.MatchesHeader([]Channel_t{do1}, do2, STATESERVER_OBJECT_GET_AI, -1); ok {
			context = dgi.ReadUint32()
		} else if ok, _ := dgi.MatchesHeader([]Channel_t{do1}, 5, STATESERVER_OBJECT_CHANGING_LOCATION, -1); ok {
			continue
		} else {
			t.Error("Received unexpected or non-matching header")
		}
	}

	// The parent should reply with AI #2 and a location acknowledgement
	dg = (&TestDatagram{}).Create([]Channel_t{do2}, do1, STATESERVER_OBJECT_GET_AI_RESP)
	dg.AddUint32(context)
	dg.AddDoid(Doid_t(do1))
	dg.AddChannel(ai2Chan)
	ack := (&TestDatagram{}).Create([]Channel_t{do2}, do1, STATESERVER_OBJECT_LOCATION_ACK)
	ack.AddDoid(Doid_t(do1))
	ack.AddZone(1500)
	obj2.ExpectMany(t, []Datagram{*dg, *ack}, false, true)

	// We should also receive a wake children message
	dg = (&TestDatagram{}).Create([]Channel_t{ParentToChildren(Doid_t(do2))}, do2, STATESERVER_OBJECT_GET_LOCATION)
	dg.AddUint32(STATESERVER_CONTEXT_WAKE_CHILDREN)
	children2.Expect(t, *dg, false)

	// DO #2 should the announce its presence to AI #2
	dg = (&TestDatagram{}).Create([]Channel_t{ai2Chan}, do2, STATESERVER_OBJECT_ENTER_AI_WITH_REQUIRED)
	appendMeta(dg, Doid_t(do2), Doid_t(do1), 1500, DistributedTestObject1)
	dg.AddUint32(1337)
	ai2.Expect(t, *dg, false)
	children2.ExpectNone(t)

	// Test for child AI handling on reparent
	// Delete DO #2 and close our channel to it to let it generate for the next test
	deleteObject(conn, 5, Doid_t(do2))
	obj2.Close()
	time.Sleep(50 * time.Millisecond)
	children2.Flush()
	obj1.Flush()
	obj2.Flush()
	ai2.Flush()

	// Recreate DO #2 w/o a parent
	instantiateObject(conn, 5, Doid_t(do2), 0, 0, 1337)

	// Set the location of DO #2 to a zone of the first object
	dg = (&TestDatagram{}).Create([]Channel_t{do2}, 5, STATESERVER_OBJECT_SET_LOCATION)
	appendMeta(dg, 6969, Doid_t(do1), 1500, 6969)
	conn.SendDatagram(*dg)

	// Ignore location change messages
	time.Sleep(10 * time.Millisecond)
	obj2.Flush()
	conn.Flush()
	children2.Flush()

	// DO #1 is expecting two messages from the child
	for n := 0; n < 2; n++ {
		obj1.Timeout = 2000
		dg = obj1.ReceiveMaybe()
		assert.True(t, dg != nil, "Parent did not receive ChangingLocation and/or GetAI")
		dgi := (&TestDatagram{}).Set(dg)
		if ok, _ := dgi.MatchesHeader([]Channel_t{do1}, do2, STATESERVER_OBJECT_GET_AI, -1); ok {
			context = dgi.ReadUint32()
		} else if ok, _ := dgi.MatchesHeader([]Channel_t{do1}, 5, STATESERVER_OBJECT_CHANGING_LOCATION, -1); ok {
			continue
		} else {
			t.Error("Received unexpected or non-matching header")
		}
	}

	// DO #2 should also announce its presence to AI #2
	dg = (&TestDatagram{}).Create([]Channel_t{ai2Chan}, do2, STATESERVER_OBJECT_ENTER_AI_WITH_REQUIRED)
	appendMeta(dg, Doid_t(do2), Doid_t(do1), 1500, DistributedTestObject1)
	dg.AddUint32(1337)
	ai2.Expect(t, *dg, false)
	children2.ExpectNone(t)

}
