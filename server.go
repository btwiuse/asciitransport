package asciitransport

import (
	"io"
	"net"
	"sync"
	"time"
)

type AsciiTransportServer interface {
	ResizeEvent() <-chan *ResizeEvent
	InputEvent() <-chan *InputEvent
	Output([]byte)
	Done() <-chan struct{}
	Close()
}

func Server(conn net.Conn, opts ...Opt) AsciiTransportServer {
	at := &AsciiTransport{
		conn:      conn,
		quit:      make(chan struct{}),
		closeonce: &sync.Once{},
		start:     time.Now(),
		iech:      make(chan *InputEvent),
		oech:      make(chan *OutputEvent),
		rech:      make(chan *ResizeEvent),
		isClient:  false,
	}
	pr, pw := io.Pipe()
	go func() {
		io.Copy(pw, conn)
		at.Close()
	}()
	at.goReadConn(pr)
	at.goWriteConn(conn)
	for _, opt := range opts {
		opt(at)
	}
	return at
}
