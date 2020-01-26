// +build console

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/btwiuse/pretty"
	"github.com/cirocosta/asciinema-edit/cast"
)

type AsciiTransportClient interface {
	OutputEvent() <-chan *OutputEvent
	Input([]byte)
	Resize(uint, uint)
}

type AsciiTransportServer interface {
	ResizeEvent() <-chan *ResizeEvent
	InputEvent() <-chan *InputEvent
	Output([]byte)
}

func Client(conn net.Conn) AsciiTransportClient {
	at := &AsciiTransport{
		conn:     conn,
		start:    time.Now(),
		logger:   log.New(os.Stderr, "AT-client", 0),
		iech:     make(chan *InputEvent),
		oech:     make(chan *OutputEvent),
		rech:     make(chan *ResizeEvent),
		isClient: true,
	}
	at.goReadConn()
	at.goWriteConn()
	return at
}

func Server(conn net.Conn) AsciiTransportServer {
	at := &AsciiTransport{
		conn:     conn,
		start:    time.Now(),
		logger:   log.New(os.Stderr, "AT-server", 0),
		iech:     make(chan *InputEvent),
		oech:     make(chan *OutputEvent),
		rech:     make(chan *ResizeEvent),
		isClient: false,
	}
	at.goReadConn()
	at.goWriteConn()
	return at
}

type AsciiTransport struct {
	conn     net.Conn
	start    time.Time
	logger   *log.Logger
	iech     chan *InputEvent
	oech     chan *OutputEvent
	rech     chan *ResizeEvent
	isClient bool
}

type ResizeEvent cast.Header
type OutputEvent Event
type InputEvent Event
type Event cast.Event

func (e *Event) UnmarshalJSON(buf []byte) error {
	tmp := []interface{}{&e.Time, &e.Type, &e.Data}
	wantLen := len(tmp)
	if err := json.Unmarshal(buf, &tmp); err != nil {
		return err
	}
	if g, w := len(tmp), wantLen; g != w {
		return fmt.Errorf("wrong number of fields in Notification: %d != %d", g, w)
	}
	return nil
}

func (e *ResizeEvent) String() string { return pretty.JsonString(e) }
func (e *InputEvent) String() string {
	return pretty.JsonString([]interface{}{&e.Time, &e.Type, &e.Data})
}
func (e *OutputEvent) String() string {
	return pretty.JsonString([]interface{}{&e.Time, &e.Type, &e.Data})
}

func (c *AsciiTransport) goReadConn() {
	go func() {
		scanner := bufio.NewScanner(c.conn)
		for scanner.Scan() {
			var (
				buf  = scanner.Bytes()
				line = scanner.Text()
			)
			if line[0] == '[' {
				var (
					e   = new(Event)
					err = json.Unmarshal(buf, e)
				)
				if err != nil {
					log.Println(err)
					continue
				}
				switch e.Type {
				case "i":
					var (
						ie  = (*InputEvent)(e)
						str = ie.String()
					)
					c.logger.Print(str)
					// consumed by reading <-AsciiTransportServer.OutputEvent()
					c.iech <- ie
				case "o":
					var (
						oe  = (*OutputEvent)(e)
						str = oe.String()
					)
					c.logger.Print(str)
					// consumed by reading <-AsciiTransportClient.OutputEvent()
					c.oech <- oe
				default:
					log.Println("unknown message:", e)
				}
			}
			if line[0] == '{' {
				var (
					re  = new(ResizeEvent)
					err = json.Unmarshal(buf, re)
				)
				if err != nil {
					log.Println(err)
					continue
				}
				var (
					str = re.String()
				)
				c.logger.Print(str)
				// consumed by reading <-AsciiTransportServer.ResizeEvent()
				c.rech <- re
			}
		}
	}()
}

func (c *AsciiTransport) goWriteConn() {
	var (
		clientInput2Server = func() {
			for {
				var (
					ie  = <-c.iech
					str = ie.String()
				)
				c.logger.Print(str)
				_, err := io.Copy(c.conn, strings.NewReader(str))
				if err != nil {
					log.Println(err)
					break
				}
			}
			exit()
		}
		clientResize2Server = func() {
			for {
				var (
					re  = <-c.rech
					str = re.String()
				)
				c.logger.Print(str)
				_, err := io.Copy(c.conn, strings.NewReader(str))
				if err != nil {
					log.Println(err)
					break
				}
			}
			exit()
		}
		serverOutput2Client = func() {
			for {
				var (
					oe  = <-c.oech
					str = oe.String()
				)
				c.logger.Print(str)
				_, err := io.Copy(c.conn, strings.NewReader(str))
				if err != nil {
					log.Println(err)
					break
				}
			}
			exit()
		}
	)
	if c.isClient {
		go clientInput2Server()
		go clientResize2Server()
	} else {
		go serverOutput2Client()
	}
}

func (c *AsciiTransport) OutputEvent() <-chan *OutputEvent { return c.oech }
func (s *AsciiTransport) InputEvent() <-chan *InputEvent   { return s.iech }
func (s *AsciiTransport) ResizeEvent() <-chan *ResizeEvent { return s.rech }

func (c *AsciiTransport) Input(buf []byte) {
	var (
		str = string(buf)
		e   = &Event{Time: 0, Type: "i", Data: str}
		ie  = (*InputEvent)(e)
	)

	if fmt.Sprintf("%q", str) == `"\x1b\x1b"` {
		exit()
	}

	c.iech <- ie
}

func (c *AsciiTransport) Output(buf []byte) {
	var (
		str = string(buf)
		e   = &Event{Time: 0, Type: "o", Data: str}
		oe  = (*OutputEvent)(e)
	)

	c.oech <- oe
}

func (c *AsciiTransport) Resize(height, width uint) {
	ie := &ResizeEvent{
		Width:  width,
		Height: height,
	}
	c.rech <- ie
}

func exit() {
	exec.Command("reset").Run()
	os.Exit(1)
}
