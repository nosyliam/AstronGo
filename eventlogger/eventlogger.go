package eventlogger

import (
	"astrongo/core"
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

	handleInterrupts()
	go listen()
}

func createLog() {
	if logfile != nil {
		logfile.Close()
	}

	t := time.Now()
	logfile, err := os.OpenFile(strftime.Format(core.Config.Eventlogger.Output, t), os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		EventLoggerLog.Fatalf("failed to open logfile: %s", err)
		return
	}

	logfile.Write([]byte(""))
}

func handleInterrupts() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		server.Close()
	}()
}

func processPacket(data []byte, addr *net.UDPAddr) {
	var out map[string]interface{}
	err := msgpack.Unmarshal(data, &out)
	if err == nil {

	}
}

func listen() {
	buff := make([]byte, 1024)
	for {
		n, addr, err := server.ReadFromUDP(buff)
		if err != nil {
			EventLoggerLog.Fatalf("failed to read from eventlogger socket: %s", err)
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
