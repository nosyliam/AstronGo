// Modified version of cli.go from github.com/apex/log
package core

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/fatih/color"
	colorable "github.com/mattn/go-colorable"
)

// Default handler outputting to stderr.
var Log = NewLogger(os.Stderr)

var bold = color.New(color.Bold)
var grey = color.New(color.FgHiBlack)

// Strings mapping.
var Strings = [...]string{
	log.DebugLevel: "DEBUG",
	log.InfoLevel:  "INFO",
	log.WarnLevel:  "WARNING",
	log.ErrorLevel: "ERROR",
	log.FatalLevel: "FATAL",
}

// Colors mapping.
var Colors = [...]*color.Color{
	log.DebugLevel: color.New(color.FgWhite),
	log.InfoLevel:  color.New(color.FgBlue),
	log.WarnLevel:  color.New(color.FgYellow),
	log.ErrorLevel: color.New(color.FgRed),
	log.FatalLevel: color.New(color.FgRed),
}

// Handler implementation.
type Handler struct {
	mu     sync.Mutex
	Writer io.Writer
}

// New handler.
func NewLogger(w io.Writer) *Handler {
	if f, ok := w.(*os.File); ok {
		return &Handler{
			Writer: colorable.NewColorable(f),
		}
	}

	return &Handler{
		Writer: w,
	}
}

// HandleLog implements log.Handler.
func (h *Handler) HandleLog(e *log.Entry) error {
	color := Colors[e.Level]
	level := Strings[e.Level]
	name := e.Fields.Get("name")
	t := time.Now()

	h.mu.Lock()
	defer h.mu.Unlock()

	grey.Fprintf(h.Writer, "[%s] ", t.Format("2006-01-02 01:02:03"))
	color.Fprintf(h.Writer, bold.Sprintf("%*s: ", 1, level))
	fmt.Fprintf(h.Writer, "%s: %s", name, e.Message)

	fmt.Fprintln(h.Writer)

	return nil
}
