// +build console

package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/containerd/console"
)

func main() {
	fmt.Println("Press ESC twice to exit.")
	term, err := console.ConsoleFromFile(os.Stdin)
	if err != nil {
		panic(err)
	}

	if err := term.SetRaw(); err != nil {
		panic(err)
	}

	conn, err := net.Dial("tcp", ":12345")
	if err != nil {
		panic(err)
	}

	client := Client(conn)

	// send
	// i
	go func() {
		// make([]byte, 0, 4096) causes 0 return
		for buf := make([]byte, 4096); ; {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				log.Println(err)
				break
			}
			// log.Println(n)
			// time.Sleep(time.Second)
			client.Input(buf[:n])
		}
		exit()
	}()

	// recv
	// o
	go func() {
		for {
			oe := <-client.OutputEvent()
			continue // discard
			_, err := io.Copy(os.Stdout, strings.NewReader(oe.Data))
			if err != nil {
				log.Println(err)
				break
			}
		}
		exit()
	}()

	// send
	// r
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

			// log.Println(currentSize)
			client.Resize(
				uint(currentSize.Height),
				uint(currentSize.Width),
			)
		}
	}
}
