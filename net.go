// +build console

package main

import (
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/containerd/console"
)

func main() {
	term, err := console.ConsoleFromFile(os.Stdin)
	if err != nil {
		panic(err)
	}

	if err := term.SetRaw(); err != nil {
		panic(err)
	}

	conn, err := net.Dial("tcp", "vm2.k0s.io:12345")
	if err != nil {
		panic(err)
	}

	// conn.Write([]byte("go run utty.go\nsleep 1 && reset\n"))
	go func() {
		if _, err := io.Copy(conn, os.Stdin); err != nil {
			log.Println(err)
		}
		term.Reset()
		os.Exit(1)
	}()

	go func() {
		if _, err := io.Copy(os.Stdout, conn); err != nil {
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
