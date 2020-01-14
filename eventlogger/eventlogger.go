package eventlogger

import (
	"astrongo/core"
	"encoding/json"
	"github.com/apex/log"
	"github.com/jehiah/go-strftime"
	"github.com/vmihailenco/msgpack"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var EventLoggerLog *log.Entry

var logfile *os.File
var server *net.UDPConn

type LoggedEvent struct {
	keys map[string]interface{}
}

func NewLoggedEvent(tp string, sender string) LoggedEvent {
	le := LoggedEvent{make(map[string]interface{})}
	le.keys["type"] = tp
	le.keys["sender"] = sender
	return le
}

func (l *LoggedEvent) add(key string, val string) {
	l.keys[key] = val
}

func (l *LoggedEvent) send() {
	msg, err := msgpack.Marshal(l.keys)
	if err != nil {
		EventLoggerLog.Warnf("failed to marshal %s event: %s", l.keys["type"], err)
	}

	processPacket(msg, &net.UDPAddr{IP: []byte{0, 0, 0, 0}})
}

func StartEventLogger() {
	if core.Config.Eventlogger.Bind == "" {
		core.Config.Eventlogger.Bind = "0.0.0.0:7197"
	}

	if core.Config.Eventlogger.Output == "" {
		core.Config.Eventlogger.Output = "events-%Y%m%d-%H%M%S.log"
	}

	createLog()

	EventLoggerLog.Info("Opening UDP socket...")
	addr, err := net.ResolveUDPAddr("udp", core.Config.Eventlogger.Bind)
	if err != nil {
		EventLoggerLog.Fatalf("Unable to open socket: %s", err)
	}

	server, err = net.ListenUDP("udp", addr)
	if err != nil {
		EventLoggerLog.Fatalf("Unable to open socket: %s", err)
	}

	event := NewLoggedEvent("log-opened", "EventLogger")
	event.add("msg", "Log opened upon Event Logger startup.")
	event.send()

	handleInterrupts()
	go listen()
}

func createLog() {
	var err error
	if logfile != nil {
		logfile.Close()
	}

	t := time.Now()
	logfile, err = os.OpenFile(strftime.Format(core.Config.Eventlogger.Output, t), os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		EventLoggerLog.Fatalf("failed to open logfile: %s", err)
		return
	}

	logfile.Truncate(0)
	logfile.Seek(0, 0)
}

func handleInterrupts() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		server.Close()
		logfile.Sync()
		logfile.Close()
	}()
}

func processPacket(data []byte, addr *net.UDPAddr) {
	var out map[string]interface{}
	err := msgpack.Unmarshal(data, &out)
	if err != nil {
		EventLoggerLog.Warnf("failed to unmarshal event from client %s: %s", addr.IP, err)
		return
	}

	out["_time"] = strftime.Format("%Y-%m-%d %H:%M:%S%z", time.Now())
	final, _ := json.Marshal(out)
	_, err = logfile.WriteString(string(final) + "\n")
	if err != nil {
		EventLoggerLog.Fatalf("failed to write to logfile: %s", err)
	}
	logfile.Sync()
}

func listen() {
	buff := make([]byte, 1024)
	for {
		n, addr, err := server.ReadFromUDP(buff)
		if err != nil {
			// If the socket is unreadable the daemon is probably closed
			break
		}

		processPacket(buff[0:n], addr)
	}
}

func init() {
	EventLoggerLog = log.WithFields(log.Fields{
		"name": "MD",
	})
}
