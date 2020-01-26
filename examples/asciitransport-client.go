// +build console

package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/btwiuse/asciitransport"
	"github.com/containerd/console"
)

func main() {
	log.Println("Press ESC twice to exit.")

	var (
		conn net.Conn
		err  error
	)
	for {
		conn, err = net.Dial("tcp", ":12345")
		if err != nil {
			log.Println(err)
			time.Sleep(time.Second)
			continue
		}
		break
	}

	term, err := console.ConsoleFromFile(os.Stdin)
	if err != nil {
		panic(err)
	}

	if err = term.SetRaw(); err != nil {
		panic(err)
	}

	opts := []asciitransport.Opt{
		asciitransport.WithLogger(os.Stdout),
		asciitransport.WithReader(os.Stdin),
		// asciitransport.WithWriter(os.Stdout),
	}
	client := asciitransport.Client(conn, opts...)

	// send
	// r
	go func() {
		sig := make(chan os.Signal)
		signal.Notify(sig, syscall.SIGWINCH)

		for {
			currentSize, err := term.Size()
			if err != nil {
				log.Println(err)
				continue
			}

			// log.Println(currentSize)
			client.Resize(
				uint(currentSize.Height),
				uint(currentSize.Width),
			)

			switch <-sig {
			case syscall.SIGWINCH:
			}
		}
	}()

	<-client.Done()
	term.Reset()
}
