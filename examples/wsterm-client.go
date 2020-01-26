// +build console

package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/btwiuse/wetty/pkg/utils"
	"github.com/containerd/console"
	"github.com/gorilla/websocket"
)

func main() {
	term, err := console.ConsoleFromFile(os.Stdin)
	if err != nil {
		panic(err)
	}

	if err := term.SetRaw(); err != nil {
		panic(err)
	}

	dialer := &websocket.Dialer{}
	conn, _, err := dialer.Dial("ws://127.0.0.1:12345/", nil)
	// conn, err := net.Dial("tcp", "vm2.k0s.io:12345")
	if err != nil {
		panic(err)
	}
	rwc := utils.WsConnToReadWriter(conn)

	// conn.Write([]byte("go run utty.go\nsleep 1 && reset\n"))
	go func() {
		if _, err := io.Copy(rwc, os.Stdin); err != nil {
			log.Println(err)
		}
		term.Reset()
		os.Exit(1)
	}()

	go func() {
		if _, err := io.Copy(os.Stdout, rwc); err != nil {
			log.Println(err)
		}
		term.Reset()
		os.Exit(1)
	}()

	select {}
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGWINCH)
	for {
		switch <-sig {
		case syscall.SIGWINCH:
			currentSize, err := term.Size()
			if err != nil {
				log.Println(err)
				continue
			}

			log.Println(currentSize)
		}
	}
}
