// Original code derived from https://github.com/ortuman/jackal

package net

import (
	"astrongo/util"
	"bufio"
	"io"
	"net"
	"time"
)

const socketBuffSize = 4096

type socketTransport struct {
	conn       net.Conn
	rw         io.ReadWriter
	br         *bufio.Reader
	bw         *bufio.Writer
	keepAlive  time.Duration
	compressed bool
}

// NewSocketTransport creates a socket class stream transport.
func NewSocketTransport(conn net.Conn, keepAlive time.Duration) Transport {
	s := &socketTransport{
		conn:      conn,
		rw:        conn,
		br:        bufio.NewReaderSize(conn, socketBuffSize),
		bw:        bufio.NewWriterSize(conn, socketBuffSize),
		keepAlive: keepAlive,
	}
	return s
}

func (s *socketTransport) Read(p []byte) (n int, err error) {
	if s.keepAlive > 0 {
		s.conn.SetReadDeadline(time.Now().Add(s.keepAlive))
	}

	return s.br.Read(p)
}

func (s *socketTransport) Write(p []byte) (n int, err error) {
	return s.bw.Write(p)
}

func (s *socketTransport) WriteDatagram(datagram util.Datagram) (n int, err error) {
	return s.bw.Write(datagram.Bytes())
}

func (s *socketTransport) Close() error {
	return s.conn.Close()
}

// Flush writes any buffered data to the underlying io.Writer.
func (s *socketTransport) Flush() error {
	return s.bw.Flush()
}
