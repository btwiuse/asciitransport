// +build console

package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/btwiuse/asciitransport"
	"github.com/containerd/console"
)

func main() {
	fmt.Println("Press ESC twice to exit.")

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

	client := asciitransport.Client(conn)

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

func exit() {
	exec.Command("reset").Run()
	os.Exit(1)
}
