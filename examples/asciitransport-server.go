// +build console

package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/btwiuse/asciitransport"
	"github.com/containerd/console"
)

func handle(conn net.Conn) {
	logname := fmt.Sprintf("AT-server-%s.log", time.Now().Format("20060102-150405"))
	log.Println("writing to", logname)
	logfile, err := os.Create(logname)
	if err != nil {
		panic(err)
	}
	opts := []asciitransport.Opt{
		asciitransport.WithLogger(logfile),
	}
	server := asciitransport.Server(conn, opts...)
	term, name, _ := console.NewPty()
	log.Println("running bash with tty", name)

	// send
	// o
	go func() {
		// make([]byte, 0, 4096) causes 0 return
		for buf := make([]byte, 4096); ; {
			n, err := term.Read(buf)
			if err != nil {
				log.Println(err)
				break
			}
			// log.Println(string(buf[:n]))
			server.Output(buf[:n])
		}
		server.Close()
	}()

	// recv
	// i
	go func() {
		for {
			ie := <-server.InputEvent()
			_, err := io.Copy(term, strings.NewReader(ie.Data))
			if err != nil {
				log.Println(err)
				break
			}
		}
		server.Close()
	}()

	go func() {
		for {
			re := <-server.ResizeEvent()
			sz := console.WinSize{
				Width:  uint16(re.Width),
				Height: uint16(re.Height),
			}
			err := term.Resize(sz)
			if err != nil {
				log.Println(err)
				break
			}
		}
		server.Close()
	}()

	<-server.Done()
	log.Println(name, "detached", term.Close())
}

func main() {
	port := ":12345"
	log.Println("listening on", port)
	ln, err := net.Listen("tcp", port)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handle(conn)
	}
}
