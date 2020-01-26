// +build console

package main

import (
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/btwiuse/consoled/asciitransport"
	"github.com/containerd/console"
)

func handle(conn net.Conn) {
	server := asciitransport.Server(conn)
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
		exit()
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
		exit()
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
		exit()
	}()
}

func main() {
	ln, err := net.Listen("tcp", ":12345")
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

func exit() {
	exec.Command("reset").Run()
	os.Exit(1)
}
