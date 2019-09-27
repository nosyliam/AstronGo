// Original code derived from https://github.com/ortuman/jackal

package net

import (
	"astrongo/util"
	"io"
)

// Transport represents a stream transport mechanism.
type Transport interface {
	io.ReadWriteCloser

	// WriteString writes a datagram to the transport
	WriteDatagram(datagram util.Datagram) (n int, err error)

	// Flush writes any buffered data to the underlying io.Writer.
	Flush() chan error
}
